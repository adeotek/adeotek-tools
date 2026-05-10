import * as pty from 'node-pty'
import type { WebSocket } from 'ws'
import type { FastifyInstance } from 'fastify'
import { db } from '../db/schema'

const IDLE_TIMEOUT_MS = 30 * 60 * 1000
const ANSI_RE = /\x1B\[[0-9;]*[A-Za-z]|\x1B\][^\x07]*\x07|\x1B[()][AB012]/g

interface ClientMessage {
  type: 'input' | 'resize' | 'interrupt' | 'chat'
  data?: string
  cols?: number
  rows?: number
}

interface ServerMessage {
  type: 'output' | 'message' | 'status'
  data?: string
  role?: string
  content?: string
  state?: string
}

class ActiveSession {
  private proc: pty.IPty
  private idleTimer: NodeJS.Timeout | null = null
  private messageBuffer = ''
  private inAssistantBlock = false
  private sockets = new Set<WebSocket>()

  constructor(readonly id: string, workdir: string) {
    const claudeBin = process.env.CLAUDE_BIN ?? 'claude'
    this.proc = pty.spawn(claudeBin, [], {
      name: 'xterm-color',
      cols: 220,
      rows: 50,
      cwd: workdir,
      env: { ...process.env, TERM: 'xterm-color' },
    })

    this.proc.onData((data) => this.onPtyData(data))
    this.proc.onExit(() => {
      this.broadcast({ type: 'status', state: 'idle' })
      db.prepare('UPDATE sessions SET ended_at = ? WHERE id = ?').run(Date.now(), id)
    })
    this.resetIdle()
  }

  attach(ws: WebSocket) {
    this.sockets.add(ws)
    this.broadcast({ type: 'status', state: 'running' })
    ws.on('close', () => this.sockets.delete(ws))
  }

  send(msg: ClientMessage, sessionId: string) {
    this.resetIdle()
    switch (msg.type) {
      case 'input':
        if (msg.data) this.proc.write(msg.data)
        break
      case 'resize':
        if (msg.cols && msg.rows) this.proc.resize(msg.cols, msg.rows)
        break
      case 'interrupt':
        this.proc.write('\x03') // Ctrl+C
        break
      case 'chat': {
        const text = msg.data as string
        // Forward to PTY as input (with newline if not already ending with one)
        const ptyInput = text.endsWith('\n') ? text : text + '\n'
        this.proc.write(ptyInput)
        // Persist as user message
        db.prepare(
          'INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)',
        ).run(sessionId, 'user', text.replace(/\n$/, ''), Date.now())
        break
      }
    }
  }

  kill() {
    this.proc.kill()
    if (this.idleTimer) clearTimeout(this.idleTimer)
  }

  private onPtyData(data: string) {
    this.broadcast({ type: 'output', data })
    this.parseAssistantContent(data)
  }

  private parseAssistantContent(raw: string) {
    const clean = raw.replace(ANSI_RE, '')
    this.messageBuffer += clean

    // Claude Code wraps assistant turns between recognizable markers
    if (!this.inAssistantBlock && this.messageBuffer.includes('❯❯❯')) {
      this.inAssistantBlock = true
      this.messageBuffer = this.messageBuffer.split('❯❯❯').pop() ?? ''
    }

    if (this.inAssistantBlock) {
      const endIdx = this.messageBuffer.indexOf('❖')
      if (endIdx !== -1) {
        const content = this.messageBuffer.slice(0, endIdx).trim()
        if (content) {
          this.broadcast({ type: 'message', role: 'assistant', content })
          db.prepare(
            'INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)',
          ).run(this.id, 'assistant', content, Date.now())
        }
        this.messageBuffer = this.messageBuffer.slice(endIdx + 1)
        this.inAssistantBlock = false
      }
    }

    if (this.messageBuffer.length > 8_192) {
      this.messageBuffer = this.messageBuffer.slice(-4_096)
    }
  }

  private broadcast(msg: ServerMessage) {
    const payload = JSON.stringify(msg)
    for (const ws of this.sockets) {
      if (ws.readyState === ws.OPEN) ws.send(payload)
    }
  }

  private resetIdle() {
    if (this.idleTimer) clearTimeout(this.idleTimer)
    this.idleTimer = setTimeout(() => this.kill(), IDLE_TIMEOUT_MS)
  }
}

class SessionManager {
  private sessions = new Map<string, ActiveSession>()

  getOrCreate(id: string, workdir: string): ActiveSession {
    if (!this.sessions.has(id)) {
      const session = new ActiveSession(id, workdir)
      this.sessions.set(id, session)
    }
    return this.sessions.get(id)!
  }

  get(id: string): ActiveSession | undefined {
    return this.sessions.get(id)
  }

  kill(id: string) {
    this.sessions.get(id)?.kill()
    this.sessions.delete(id)
  }
}

export const sessionManager = new SessionManager()

export async function sessionWsRoutes(fastify: FastifyInstance) {
  fastify.get<{ Params: { id: string } }>(
    '/ws/session/:id',
    { websocket: true },
    (socket, req) => {
      const { id } = req.params

      const row = db.prepare('SELECT * FROM sessions WHERE id = ?').get(id) as
        | { workdir: string; ended_at: number | null }
        | undefined

      if (!row) {
        socket.send(JSON.stringify({ type: 'status', state: 'error', data: 'session not found' }))
        socket.close()
        return
      }

      // Send message history to the connecting client
      const history = db
        .prepare('SELECT role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC')
        .all(id)
      if (history.length > 0) {
        socket.send(JSON.stringify({ type: 'history', messages: history }))
      }

      // Clear ended_at if resuming a previously stopped session
      if (row.ended_at !== null) {
        db.prepare('UPDATE sessions SET ended_at = NULL WHERE id = ?').run(id)
      }

      const session = sessionManager.getOrCreate(id, row.workdir)
      session.attach(socket)

      socket.on('message', (raw: Buffer | string) => {
        try {
          const msg = JSON.parse(raw.toString()) as ClientMessage
          session.send(msg, id)
        } catch {
          // ignore malformed frames
        }
      })

      socket.on('close', () => {
        // socket removed from session's set by the attach() listener above
      })
    },
  )
}
