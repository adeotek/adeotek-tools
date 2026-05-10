import { spawn, ChildProcess } from 'child_process'
import * as readline from 'readline'
import type { WebSocket } from 'ws'
import type { FastifyInstance } from 'fastify'
import { db } from '../db/schema'

const IDLE_TIMEOUT_MS = 30 * 60 * 1000

// ─── Types ────────────────────────────────────────────────────────────────────

type ContentBlock =
  | { type: 'text'; text: string }
  | { type: 'thinking'; thinking: string }
  | { type: 'tool_use'; id: string; name: string; input: Record<string, unknown> }
  | { type: 'tool_result'; tool_use_id: string; content: string | ContentBlock[]; is_error?: boolean }

interface StreamEvent {
  type: string
  subtype?: string
  session_id?: string
  result?: string
  is_error?: boolean
  message?: {
    id?: string
    content?: ContentBlock[]
    stop_reason?: string | null
  }
}

interface ServerMessage {
  type: 'output' | 'message' | 'status' | 'history'
  data?: string
  role?: string
  content?: string
  state?: string
  messages?: Array<{ role: string; content: string; created_at: number }>
}

interface ClientMessage {
  type: 'chat' | 'input' | 'resize' | 'interrupt'
  data?: string
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function getBypassPermissions(): boolean {
  const row = db
    .prepare('SELECT value FROM settings WHERE key = ?')
    .get('bypass_permissions') as { value: string } | undefined
  return row ? row.value === 'true' : true
}

function summarizeToolInput(name: string, input: Record<string, unknown>): string {
  if (name === 'Bash') return String(input.command ?? '').slice(0, 120)
  if (name === 'Read') return String(input.file_path ?? '')
  if (name === 'Write' || name === 'Edit') return String(input.file_path ?? '')
  if (name === 'Glob') return String(input.pattern ?? '')
  if (name === 'Grep') return String(input.pattern ?? '')
  return JSON.stringify(input).slice(0, 120)
}

function parseEvent(line: string): StreamEvent | null {
  try {
    return JSON.parse(line) as StreamEvent
  } catch {
    return null
  }
}

// ─── ActiveSession ────────────────────────────────────────────────────────────

class ActiveSession {
  private claudeSessionId: string | null
  private currentProc: ChildProcess | null = null
  private isRunning = false
  private sockets = new Set<WebSocket>()
  private idleTimer: NodeJS.Timeout | null = null

  constructor(
    readonly id: string,
    private readonly workdir: string,
    claudeSessionId: string | null = null,
  ) {
    this.claudeSessionId = claudeSessionId
    this.resetIdle()
  }

  attach(ws: WebSocket) {
    this.sockets.add(ws)
    ws.send(JSON.stringify({ type: 'status', state: this.isRunning ? 'running' : 'idle' }))
    ws.on('close', () => this.sockets.delete(ws))
  }

  sendMessage(text: string) {
    if (this.isRunning) {
      this.broadcast({ type: 'status', state: 'running' })
      return
    }
    this.resetIdle()
    this.isRunning = true
    this.broadcast({ type: 'status', state: 'running' })

    // Show user prompt in terminal log
    this.broadcast({ type: 'output', data: `\r\n❯ ${text}\r\n` })

    // Persist user message to DB
    db.prepare(
      'INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)',
    ).run(this.id, 'user', text, Date.now())

    const claudeBin = process.env.CLAUDE_BIN ?? 'claude'
    const bypassPermissions = getBypassPermissions()

    const args = [
      '--print',
      '--output-format', 'stream-json',
      '--verbose',
      '--include-partial-messages',
      '--input-format', 'stream-json',
    ]
    if (this.claudeSessionId) args.push('--resume', this.claudeSessionId)
    if (bypassPermissions) args.push('--dangerously-skip-permissions')

    const proc = spawn(claudeBin, args, {
      cwd: this.workdir,
      env: { ...process.env },
      stdio: ['pipe', 'pipe', 'pipe'],
    })
    this.currentProc = proc

    if (!proc.stdin || !proc.stdout || !proc.stderr) {
      this.isRunning = false
      this.broadcast({ type: 'status', state: 'error' })
      return
    }

    const inputMsg =
      JSON.stringify({ type: 'user', message: { role: 'user', content: text } }) + '\n'
    proc.stdin!.write(inputMsg)
    proc.stdin!.end()

    proc.on('error', (err) => {
      this.currentProc = null
      this.isRunning = false
      this.broadcast({ type: 'output', data: `\r\nError: ${err.message}\r\n` })
      this.broadcast({ type: 'status', state: 'error' })
    })

    proc.stdin!.on('error', () => { /* stdin EPIPE on early process exit */ })

    const rl = readline.createInterface({ input: proc.stdout! })

    // Track per-message text position to emit incremental chunks (partial messages
    // from --include-partial-messages are cumulative, not incremental)
    let lastMsgId = ''
    let lastEmittedLen = 0
    let thinkingEmitted = false
    const emittedToolIds = new Set<string>()

    rl.on('line', (line) => {
      this.resetIdle()
      const event = parseEvent(line)
      if (!event) return

      if (event.type === 'system' && event.subtype === 'init') {
        const sid = event.session_id?.slice(0, 8) ?? '?'
        this.broadcast({ type: 'output', data: `[session ${sid}]\r\n` })
        return
      }

      if (event.type === 'assistant' && event.message?.content) {
        const msgId = event.message.id ?? ''
        if (msgId !== lastMsgId) {
          lastMsgId = msgId
          lastEmittedLen = 0
          thinkingEmitted = false
          emittedToolIds.clear()
        }
        for (const block of event.message.content) {
          if (block.type === 'thinking') {
            if (!thinkingEmitted) {
              thinkingEmitted = true
              const preview = block.thinking.slice(0, 300)
              this.broadcast({ type: 'output', data: `\x1b[2m💭 ${preview}\x1b[0m\r\n` })
            }
          } else if (block.type === 'tool_use') {
            if (!emittedToolIds.has(block.id)) {
              emittedToolIds.add(block.id)
              const summary = summarizeToolInput(block.name, block.input ?? {})
              this.broadcast({ type: 'output', data: `⚙ ${block.name}: ${summary}\r\n` })
            }
          } else if (block.type === 'text') {
            const fullText = block.text
            const chunk = fullText.slice(lastEmittedLen)
            if (chunk) this.broadcast({ type: 'output', data: chunk })
            lastEmittedLen = fullText.length
          }
        }
      }

      if (event.type === 'user' && event.message?.content) {
        for (const block of event.message.content) {
          if (block.type === 'tool_result') {
            const raw = block.content
            const content = (typeof raw === 'string' ? raw : JSON.stringify(raw)).slice(0, 500)
            this.broadcast({ type: 'output', data: `→ ${content}\r\n` })
          }
        }
      }

      if (event.type === 'result') {
        if (event.session_id) {
          this.claudeSessionId = event.session_id
          db.prepare('UPDATE sessions SET claude_session_id = ? WHERE id = ?').run(
            event.session_id,
            this.id,
          )
        }
        const finalText = event.result ?? ''
        this.broadcast({ type: 'message', role: 'assistant', content: finalText })
        if (finalText) {
          db.prepare(
            'INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)',
          ).run(this.id, 'assistant', finalText, Date.now())
        }
        this.isRunning = false
        this.broadcast({ type: 'status', state: 'idle' })
        this.resetIdle()
      }
    })

    proc.stderr!.on('data', () => { /* suppress hook/debug noise */ })

    rl.on('close', () => {
      this.currentProc = null
      if (this.isRunning) {
        this.isRunning = false
        this.broadcast({ type: 'status', state: 'error' })
      }
    })
  }

  kill() {
    db.prepare('UPDATE sessions SET ended_at = ? WHERE id = ?').run(Date.now(), this.id)
    if (this.currentProc) {
      this.currentProc.kill('SIGTERM')
      this.currentProc = null
    }
    if (this.idleTimer) clearTimeout(this.idleTimer)
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

// ─── SessionManager ───────────────────────────────────────────────────────────

class SessionManager {
  private sessions = new Map<string, ActiveSession>()

  getOrCreate(id: string, workdir: string, claudeSessionId: string | null = null): ActiveSession {
    if (!this.sessions.has(id)) {
      this.sessions.set(id, new ActiveSession(id, workdir, claudeSessionId))
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

// ─── WebSocket Route ──────────────────────────────────────────────────────────

export async function sessionWsRoutes(fastify: FastifyInstance) {
  fastify.get<{ Params: { id: string } }>(
    '/ws/session/:id',
    { websocket: true },
    (socket, req) => {
      const { id } = req.params

      const row = db.prepare('SELECT * FROM sessions WHERE id = ?').get(id) as
        | { workdir: string; ended_at: number | null; claude_session_id: string | null }
        | undefined

      if (!row) {
        socket.send(JSON.stringify({ type: 'status', state: 'error', data: 'session not found' }))
        socket.close()
        return
      }

      // Send message history on connect
      const history = db
        .prepare(
          'SELECT role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC',
        )
        .all(id)
      if (history.length > 0) {
        socket.send(JSON.stringify({ type: 'history', messages: history }))
      }

      // Clear ended_at so resumed sessions show as active
      if (row.ended_at !== null) {
        db.prepare('UPDATE sessions SET ended_at = NULL WHERE id = ?').run(id)
      }

      const session = sessionManager.getOrCreate(id, row.workdir, row.claude_session_id ?? null)
      session.attach(socket)

      socket.on('message', (raw: Buffer | string) => {
        try {
          const msg = JSON.parse(raw.toString()) as ClientMessage
          if (msg.type === 'chat' && msg.data) {
            session.sendMessage(msg.data.replace(/\n$/, ''))
          }
          // type:'input' and type:'resize' are PTY-only — silently ignored
        } catch {
          // ignore malformed frames
        }
      })
    },
  )
}
