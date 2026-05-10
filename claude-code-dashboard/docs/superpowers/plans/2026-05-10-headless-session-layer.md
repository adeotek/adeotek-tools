# Headless Session Layer Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the broken PTY + ANSI-scraping session backend with `claude --print --output-format stream-json` per-turn subprocesses, giving reliable chat message extraction and a real-time terminal log.

**Architecture:** Each user message spawns a `claude --print --output-format stream-json --verbose --include-partial-messages --input-format stream-json` subprocess. The message is written to stdin; stdout is parsed as JSONL events. Tool calls, thinking blocks, and text chunks go to the terminal log (`type:'output'`); the final response goes to the chat view (`type:'message'`). The Claude-assigned `session_id` from the `result` event is stored in the DB and passed as `--resume` on subsequent turns. A `settings` table stores the bypass-permissions flag.

**Tech Stack:** Node.js `child_process.spawn` (stdlib — no new deps), `readline` (stdlib), `better-sqlite3`, Fastify, React + TypeScript, Tailwind.

**Spec:** `docs/superpowers/specs/2026-05-10-headless-session-layer-design.md`

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `backend/src/db/schema.ts` | Modify | Add `claude_session_id` column + `settings` table |
| `backend/src/routes/settings.ts` | Create | `GET/POST /api/settings` |
| `backend/src/server.ts` | Modify | Register `settingsRoutes` |
| `backend/src/ws/session.ts` | Rewrite | `ActiveSession` using `child_process` + stream-json parser |
| `backend/package.json` | Modify | Remove `node-pty` dependency |
| `frontend/src/views/SettingsView.tsx` | Modify | Add bypass-permissions toggle |
| `frontend/src/views/DashboardView.tsx` | Modify | Remove unused PTY resize wiring |

---

## Task 1: DB Schema — add `claude_session_id` column and `settings` table

**Files:**
- Modify: `backend/src/db/schema.ts`

- [ ] **Step 1: Add the migration code to `schema.ts`**

Open `backend/src/db/schema.ts`. After the `oauth_cache` table creation and before `return db`, add:

```typescript
  // Migrate: add claude_session_id to sessions if missing
  const sessionCols = db.pragma('table_info(sessions)') as Array<{ name: string }>
  if (!sessionCols.find((c) => c.name === 'claude_session_id')) {
    db.prepare('ALTER TABLE sessions ADD COLUMN claude_session_id TEXT').run()
  }

  db.prepare(`
    CREATE TABLE IF NOT EXISTS settings (
      key   TEXT PRIMARY KEY,
      value TEXT NOT NULL
    )
  `).run()
```

The full `initDb` function should now end with:

```typescript
  db.prepare(`
    CREATE TABLE IF NOT EXISTS oauth_cache (
      key TEXT PRIMARY KEY,
      data TEXT NOT NULL,
      cached_at INTEGER NOT NULL
    )
  `).run()

  // Migrate: add claude_session_id to sessions if missing
  const sessionCols = db.pragma('table_info(sessions)') as Array<{ name: string }>
  if (!sessionCols.find((c) => c.name === 'claude_session_id')) {
    db.prepare('ALTER TABLE sessions ADD COLUMN claude_session_id TEXT').run()
  }

  db.prepare(`
    CREATE TABLE IF NOT EXISTS settings (
      key   TEXT PRIMARY KEY,
      value TEXT NOT NULL
    )
  `).run()

  return db
}

export const db = initDb()
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add backend/src/db/schema.ts
git commit -m "feat: add claude_session_id column and settings table"
```

---

## Task 2: Settings route — `GET/POST /api/settings`

**Files:**
- Create: `backend/src/routes/settings.ts`

- [ ] **Step 1: Create the file**

```typescript
import type { FastifyInstance } from 'fastify'
import { db } from '../db/schema'

const DEFAULTS: Record<string, string> = {
  bypass_permissions: 'true',
}

export async function settingsRoutes(fastify: FastifyInstance) {
  fastify.get('/api/settings', async (_req, reply) => {
    const rows = db.prepare('SELECT key, value FROM settings').all() as Array<{
      key: string
      value: string
    }>
    const result: Record<string, string> = { ...DEFAULTS }
    for (const row of rows) result[row.key] = row.value
    return reply.send(result)
  })

  fastify.post<{ Body: Record<string, string> }>('/api/settings', async (req, reply) => {
    const entries = Object.entries(req.body)
    if (entries.length === 0) return reply.status(400).send({ error: 'empty body' })
    const upsert = db.prepare('INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)')
    const tx = db.transaction(() => {
      for (const [key, value] of entries) upsert.run(key, String(value))
    })
    tx()
    return reply.send({ ok: true })
  })
}
```

- [ ] **Step 2: Register the route in `backend/src/server.ts`**

Add the import at the top of `server.ts`:

```typescript
import { settingsRoutes } from './routes/settings'
```

Add the registration inside `start()` alongside the other route registrations:

```typescript
  await fastify.register(settingsRoutes)
```

- [ ] **Step 3: Verify it compiles**

```bash
cd backend && npx tsc --noEmit
```

Expected: no errors.

- [ ] **Step 4: Smoke-test the endpoints**

Start the backend: `cd backend && npm run dev`

In another terminal:
```bash
curl -s http://localhost:9998/api/settings | python3 -m json.tool
# Expected: {"bypass_permissions": "true"}

curl -s -X POST http://localhost:9998/api/settings \
  -H 'Content-Type: application/json' \
  -d '{"bypass_permissions":"false"}' | python3 -m json.tool
# Expected: {"ok": true}

curl -s http://localhost:9998/api/settings | python3 -m json.tool
# Expected: {"bypass_permissions": "false"}
```

Stop the backend (`Ctrl+C`).

- [ ] **Step 5: Commit**

```bash
git add backend/src/routes/settings.ts backend/src/server.ts
git commit -m "feat: add GET/POST /api/settings route"
```

---

## Task 3: Rewrite `backend/src/ws/session.ts`

This is the core change. Replace the entire file with the `child_process`-based implementation.

**Files:**
- Rewrite: `backend/src/ws/session.ts`

- [ ] **Step 1: Replace the entire file contents**

```typescript
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
    if (this.isRunning) return
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

    const inputMsg =
      JSON.stringify({ type: 'user', message: { role: 'user', content: text } }) + '\n'
    proc.stdin!.write(inputMsg)
    proc.stdin!.end()

    const rl = readline.createInterface({ input: proc.stdout! })

    // Track per-message text position to emit incremental chunks (partial messages
    // from --include-partial-messages are cumulative, not incremental)
    let lastMsgId = ''
    let lastEmittedLen = 0

    rl.on('line', (line) => {
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
        }
        for (const block of event.message.content) {
          if (block.type === 'thinking') {
            // Only emit thinking on first occurrence to avoid repetition
            if (lastEmittedLen === 0) {
              const preview = block.thinking.slice(0, 300)
              this.broadcast({ type: 'output', data: `\x1b[2m💭 ${preview}\x1b[0m\r\n` })
            }
          } else if (block.type === 'tool_use') {
            const summary = summarizeToolInput(block.name, block.input ?? {})
            this.broadcast({ type: 'output', data: `⚙ ${block.name}: ${summary}\r\n` })
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
        if (finalText) {
          this.broadcast({ type: 'message', role: 'assistant', content: finalText })
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

    proc.on('close', () => {
      this.currentProc = null
      if (this.isRunning) {
        this.isRunning = false
        this.broadcast({ type: 'status', state: 'error' })
      }
    })
  }

  kill() {
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
```

- [ ] **Step 2: Verify it compiles**

```bash
cd backend && npx tsc --noEmit
```

Expected: no errors. If you see `Cannot find module 'node-pty'` — that's fine for now; node-pty will be removed in Task 4.

- [ ] **Step 3: Commit**

```bash
git add backend/src/ws/session.ts
git commit -m "feat: replace PTY session with headless stream-json subprocess"
```

---

## Task 4: Remove `node-pty` dependency

**Files:**
- Modify: `backend/package.json`

- [ ] **Step 1: Remove node-pty from package.json**

In `backend/package.json`, delete the `"node-pty"` line from `"dependencies"`.

The dependencies section should look like:

```json
"dependencies": {
  "@anthropic-ai/sdk": "^0.52.0",
  "@fastify/cors": "^10.0.1",
  "@fastify/static": "^9.1.3",
  "@fastify/websocket": "^11.0.1",
  "better-sqlite3": "^11.0.0",
  "fastify": "^5.2.1",
  "uuid": "^11.1.0"
}
```

- [ ] **Step 2: Re-install to update lockfile**

```bash
cd backend && npm install
```

Expected: `node_modules/node-pty` should no longer exist. Verify:

```bash
ls backend/node_modules/node-pty 2>&1
# Expected: "No such file or directory"
```

- [ ] **Step 3: Final lint check**

```bash
make lint
```

Expected: both backend and frontend pass with no errors.

- [ ] **Step 4: Commit**

```bash
git add backend/package.json backend/package-lock.json
git commit -m "chore: remove node-pty dependency"
```

---

## Task 5: Frontend — add bypass-permissions toggle to `SettingsView`

**Files:**
- Modify: `frontend/src/views/SettingsView.tsx`

- [ ] **Step 1: Replace the entire file**

```tsx
import { useState, useEffect } from 'react'
import { ArrowLeft, Save } from 'lucide-react'
import { NavLink } from 'react-router-dom'

const API_KEY = 'dashboard_anthropic_api_key'

export default function SettingsView() {
  const [apiKey, setApiKey] = useState(() => localStorage.getItem(API_KEY) ?? '')
  const [bypassPermissions, setBypassPermissions] = useState(true)
  const [saved, setSaved] = useState(false)
  const [settingsLoaded, setSettingsLoaded] = useState(false)

  useEffect(() => {
    fetch('/api/settings')
      .then((r) => r.json() as Promise<Record<string, string>>)
      .then((data) => {
        setBypassPermissions(data.bypass_permissions !== 'false')
        setSettingsLoaded(true)
      })
      .catch(() => setSettingsLoaded(true))
  }, [])

  async function handleSave() {
    if (apiKey) localStorage.setItem(API_KEY, apiKey)
    else localStorage.removeItem(API_KEY)

    await fetch('/api/settings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ bypass_permissions: String(bypassPermissions) }),
    }).catch(() => {})

    setSaved(true)
    setTimeout(() => setSaved(false), 2000)
  }

  return (
    <div className="p-6 space-y-6 max-w-md">
      <NavLink
        to="/"
        className="inline-flex items-center gap-1.5 text-xs text-accent hover:text-accent-hover transition-colors"
      >
        <ArrowLeft size={13} strokeWidth={1.5} />
        Back to Dashboard
      </NavLink>

      <h2 className="text-text-secondary text-xs uppercase tracking-widest">Settings</h2>

      {/* API key */}
      <div className="space-y-2">
        <label className="text-text-secondary text-xs uppercase tracking-widest block">
          Anthropic API Key
        </label>
        <p className="text-text-muted text-xs">
          Used to fetch billing usage totals in the Usage view. Stored in localStorage only.
        </p>
        <input
          type="password"
          value={apiKey}
          onChange={(e) => setApiKey(e.target.value)}
          placeholder="sk-ant-…"
          className="w-full bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent"
        />
      </div>

      {/* Bypass permissions toggle */}
      <div className="space-y-2">
        <label className="text-text-secondary text-xs uppercase tracking-widest block">
          Tool Permissions
        </label>
        <p className="text-text-muted text-xs">
          When enabled, Claude Code runs with <code className="text-text-secondary">--dangerously-skip-permissions</code> — the same trust level as running <code className="text-text-secondary">claude</code> directly in your terminal. When disabled, writes outside the workspace are sandboxed.
        </p>
        <button
          onClick={() => setBypassPermissions((v) => !v)}
          disabled={!settingsLoaded}
          className={`flex items-center gap-2 px-3 py-1.5 text-xs rounded border transition-colors disabled:opacity-40 ${
            bypassPermissions
              ? 'border-status-green text-status-green bg-status-green/10'
              : 'border-border-subtle text-text-muted bg-bg-elevated'
          }`}
        >
          <span className={`w-2 h-2 rounded-full ${bypassPermissions ? 'bg-status-green' : 'bg-text-dim'}`} />
          {bypassPermissions ? 'Bypass permissions: on' : 'Bypass permissions: off'}
        </button>
      </div>

      <button
        onClick={handleSave}
        className="flex items-center gap-2 bg-accent hover:bg-accent-hover text-black text-sm font-medium px-4 py-2 rounded-md transition-colors"
      >
        <Save size={14} />
        {saved ? 'Saved!' : 'Save'}
      </button>
    </div>
  )
}
```

- [ ] **Step 2: Verify it compiles**

```bash
make lint
```

Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/SettingsView.tsx
git commit -m "feat: add bypass-permissions toggle to settings"
```

---

## Task 6: Frontend — remove unused PTY wiring from `DashboardView`

**Files:**
- Modify: `frontend/src/views/DashboardView.tsx`

The `onTerminalMounted` callback was already `void`-ed but its wiring can be removed entirely since there is no PTY to resize.

- [ ] **Step 1: Remove the PTY resize callback**

In `frontend/src/views/DashboardView.tsx`, find and remove these lines:

```typescript
  // Wire terminal resize → backend PTY resize
  const onTerminalMounted = useCallback(() => {
    terminalRef.current?.sendResize((cols, rows) => {
      send({ type: 'resize', cols, rows })
    })
  }, [send])
  void onTerminalMounted
```

- [ ] **Step 2: Check for any remaining `useCallback` import**

If `useCallback` is no longer used elsewhere in the file, remove it from the import:

```typescript
// Before
import { useState, useRef, useCallback } from 'react'

// After (if useCallback removed)
import { useState, useRef } from 'react'
```

Check whether `useCallback` is still used for `onOutput`:
```typescript
const onOutput = useCallback((data: string) => {
  terminalRef.current?.write(data)
}, [])
```
If yes, keep `useCallback` in the import.

- [ ] **Step 3: Verify it compiles**

```bash
make lint
```

Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/DashboardView.tsx
git commit -m "chore: remove unused PTY resize wiring"
```

---

## Task 7: Integration test

No automated test suite exists in this repo. Follow these manual steps to verify end-to-end behaviour.

- [ ] **Step 1: Start the dev stack**

```bash
make dev
```

Open `http://localhost:5173` (or `http://<host-ip>:5173`).

- [ ] **Step 2: Create a new session**

Click **New session**, enter a valid project directory (e.g. `/home/<user>/projects/react-todo-app`), click **Start session**.

Expected:
- Chat view appears with the session header
- Status dot is green/dim (idle)
- No errors in the browser console

- [ ] **Step 3: Send a simple message**

In the chat input, type `What is this project about?` and press Enter.

Expected:
- Status dot goes bright green (running)
- Terminal drawer (open it with the toggle) shows: the user prompt line (`❯ What is this project about?`), a session ID banner, and streaming text chunks as Claude responds
- After ~5–15 seconds, the chat view shows the assistant's response as a message bubble
- Status dot returns to dim (idle)

- [ ] **Step 4: Send a follow-up message**

Type `List the main files` and press Enter.

Expected:
- Claude responds with file listing (uses `--resume` under the hood so it remembers the previous turn)
- Terminal shows `⚙ Bash: ls ...` or `⚙ Glob: ...` tool calls and `→ <result>` lines
- Chat view shows the final text response

- [ ] **Step 5: Verify tool calls appear in terminal**

Ask Claude to run a bash command: `run: echo hello world`.

Expected:
- Terminal shows `⚙ Bash: echo hello world` and `→ hello world`
- Chat shows Claude's final text response

- [ ] **Step 6: Test the bypass-permissions toggle**

Go to **Settings** (gear icon), turn off **Bypass permissions**, save. Return to Dashboard, send: `write hello to /tmp/test.txt`.

Expected:
- Claude reports it was blocked (sandbox mode)
- No file written to `/tmp/test.txt`

Re-enable bypass permissions in Settings. Send the same message again.

Expected:
- File is written (`cat /tmp/test.txt` shows `hello`)

- [ ] **Step 7: Test resume after server restart**

Stop the backend (`Ctrl+C` on `make dev`). Restart (`make dev`). Go to the Sessions list, click **Resume** on the previous session. Send a message: `What did I ask you before?`.

Expected:
- Claude references earlier context (because `--resume <claude_session_id>` re-attaches to the Claude-side session)

- [ ] **Step 8: Commit if any minor fixes were needed**

```bash
git add -A
git commit -m "fix: <describe any fixes made during integration test>"
```

---

## Self-Review Checklist (completed)

| Spec requirement | Covered by task |
|---|---|
| Replace node-pty with `child_process.spawn` | Task 3 |
| `--print --output-format stream-json --verbose --include-partial-messages --input-format stream-json` flags | Task 3 |
| `--resume <claude_session_id>` for multi-turn | Task 3 (constructor + `getOrCreate`) |
| `--dangerously-skip-permissions` when bypass enabled | Task 3 (`getBypassPermissions()`) |
| `claude_session_id` column in DB | Task 1 |
| `settings` table in DB | Task 1 |
| `GET/POST /api/settings` | Task 2 |
| Register settings route | Task 2 (Step 2) |
| `thinking` blocks → terminal with `💭` prefix | Task 3 |
| `tool_use` blocks → terminal with `⚙` prefix | Task 3 |
| `tool_result` blocks → terminal with `→` prefix | Task 3 |
| Partial text chunks (incremental diff) → terminal | Task 3 (`lastEmittedLen`) |
| Final text → chat view + DB insert | Task 3 (`result` event handler) |
| System init banner in terminal | Task 3 |
| User prompt shown in terminal | Task 3 |
| Remove node-pty dep | Task 4 |
| Bypass-permissions toggle in Settings UI | Task 5 |
| Remove PTY resize wiring | Task 6 |
| `claude_session_id` restored from DB on server restart | Task 3 (`getOrCreate` third arg) |
