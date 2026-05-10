# Session Naming Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let users assign optional names to sessions at creation time and rename them later from the active-session header.

**Architecture:** Name is stored as a nullable `TEXT` column on the `sessions` table, returned in the existing `GET /api/sessions` payload, set via an extended `POST /api/sessions` body, and updated via a new `PATCH /api/sessions/:id` endpoint. The frontend propagates the name through `SessionContext` state and renders it in `SessionHeader` with an inline edit mode.

**Tech Stack:** Fastify + better-sqlite3 (backend), React + TypeScript + Tailwind (frontend), `make lint` (`tsc --noEmit`) for validation.

---

## File Map

| File | Change |
|------|--------|
| `backend/src/db/schema.ts` | Migration: add `name TEXT` column |
| `backend/src/routes/sessions.ts` | POST accepts `name`; new PATCH endpoint |
| `frontend/src/hooks/useDashboard.ts` | Add `name: string \| null` to `Session` interface |
| `frontend/src/context/SessionContext.tsx` | Add `name` to state; update `SESSION_CREATED`, `RESUME_SESSION`; add `SESSION_RENAMED` |
| `frontend/src/components/NewSessionModal.tsx` | Add optional Name field; update `onStart` signature |
| `frontend/src/components/SessionList.tsx` | Display `name ?? lastSegment(workdir)` as row title |
| `frontend/src/components/SessionHeader.tsx` | Show name; Pencil rename button; inline edit mode |
| `frontend/src/views/DashboardView.tsx` | Plumb name through create/resume; `handleRenameSession`; updated SessionHeader props |

---

## Task 1: DB Migration — add `name` column

**Files:**
- Modify: `backend/src/db/schema.ts`

- [ ] **Open `backend/src/db/schema.ts`** and find the existing `claude_session_id` migration block (lines ~55-58). Add the `name` migration immediately after it:

```typescript
  // Migrate: add claude_session_id to sessions if missing
  const sessionCols = db.pragma('table_info(sessions)') as Array<{ name: string }>
  if (!sessionCols.find((c) => c.name === 'claude_session_id')) {
    db.prepare('ALTER TABLE sessions ADD COLUMN claude_session_id TEXT').run()
  }
  // Migrate: add name to sessions if missing
  if (!sessionCols.find((c) => c.name === 'name')) {
    db.prepare('ALTER TABLE sessions ADD COLUMN name TEXT').run()
  }
```

- [ ] **Validate**

```bash
cd /path/to/claude-code-dashboard && make lint
```

Expected: no errors.

- [ ] **Commit**

```bash
git add backend/src/db/schema.ts
git commit -m "feat(db): add name column to sessions"
```

---

## Task 2: Backend API — POST accepts name, new PATCH endpoint

**Files:**
- Modify: `backend/src/routes/sessions.ts`

- [ ] **Update the POST route** to accept an optional `name` field and store it. Replace the existing `POST /api/sessions` handler:

```typescript
  fastify.post<{ Body: { workdir: string; name?: string } }>('/api/sessions', async (req, reply) => {
    const { workdir, name } = req.body
    if (!workdir || typeof workdir !== 'string') {
      return reply.status(400).send({ error: 'workdir is required' })
    }

    const id = uuidv4()
    db.prepare(
      'INSERT INTO sessions (id, workdir, name, started_at) VALUES (?, ?, ?, ?)',
    ).run(id, workdir, name?.trim() || null, Date.now())

    return reply.status(201).send({ sessionId: id })
  })
```

- [ ] **Add the PATCH route** for renaming. Insert it after the existing `POST /api/sessions/:id/stop` handler:

```typescript
  fastify.patch<{ Params: { id: string }; Body: { name: string } }>('/api/sessions/:id', async (req, reply) => {
    const { id } = req.params
    const { name } = req.body

    const session = db.prepare('SELECT id FROM sessions WHERE id = ?').get(id)
    if (!session) {
      return reply.status(404).send({ error: 'Session not found' })
    }

    db.prepare('UPDATE sessions SET name = ? WHERE id = ?').run(
      typeof name === 'string' ? name.trim() || null : null,
      id,
    )
    return reply.send({ ok: true })
  })
```

- [ ] **Validate**

```bash
make lint
```

Expected: no errors.

- [ ] **Commit**

```bash
git add backend/src/routes/sessions.ts
git commit -m "feat(api): POST sessions accepts name; add PATCH sessions/:id for rename"
```

---

## Task 3: Frontend types — Session interface + SessionContext

**Files:**
- Modify: `frontend/src/hooks/useDashboard.ts`
- Modify: `frontend/src/context/SessionContext.tsx`

- [ ] **Add `name` to the `Session` interface** in `frontend/src/hooks/useDashboard.ts`. Replace the existing interface:

```typescript
export interface Session {
  id: string
  workdir: string
  name: string | null
  model: string | null
  started_at: number
  ended_at: number | null
  is_active: boolean
  message_count: number
}
```

- [ ] **Update `SessionState`** in `frontend/src/context/SessionContext.tsx` — add `name`:

```typescript
export interface SessionState {
  sessionId: string | null
  workdir: string | null
  name: string | null
  model: string | null
  wsState: 'disconnected' | 'connecting' | 'running' | 'idle' | 'error'
  messages: Message[]
  workingTimeMs: number
  runningStartedAt: number | null
}
```

- [ ] **Update the `Action` union** — add optional `name` to `SESSION_CREATED` and `RESUME_SESSION`, add `SESSION_RENAMED`:

```typescript
type Action =
  | { type: 'SESSION_CREATED'; sessionId: string; workdir: string; name?: string }
  | { type: 'SESSION_CLEARED' }
  | { type: 'RESUME_SESSION'; id: string; workdir: string; name?: string }
  | { type: 'WS_STATE'; state: SessionState['wsState']; timestamp: number }
  | { type: 'MESSAGE_ADDED'; message: Message }
  | { type: 'HISTORY_LOADED'; messages: Message[] }
  | { type: 'MODEL_SET'; model: string }
  | { type: 'SESSION_RENAMED'; name: string | null }
```

- [ ] **Update `initial` state** — add `name: null`:

```typescript
const initial: SessionState = {
  sessionId: null,
  workdir: null,
  name: null,
  model: null,
  wsState: 'disconnected',
  messages: [],
  workingTimeMs: 0,
  runningStartedAt: null,
}
```

- [ ] **Update the reducer** — add `name` to `SESSION_CREATED` and `RESUME_SESSION` cases, add `SESSION_RENAMED` case:

```typescript
    case 'SESSION_CREATED':
      return { ...state, sessionId: action.sessionId, workdir: action.workdir, name: action.name ?? null, messages: [], wsState: 'connecting', workingTimeMs: 0, runningStartedAt: null }
    case 'SESSION_CLEARED':
      return { ...initial }
    case 'RESUME_SESSION':
      return { ...state, sessionId: action.id, workdir: action.workdir, name: action.name ?? null, messages: [], wsState: 'connecting', workingTimeMs: 0, runningStartedAt: null }
    // ... existing WS_STATE, MESSAGE_ADDED, HISTORY_LOADED, MODEL_SET cases unchanged ...
    case 'SESSION_RENAMED':
      return { ...state, name: action.name }
```

- [ ] **Validate**

```bash
make lint
```

Expected: no errors.

- [ ] **Commit**

```bash
git add frontend/src/hooks/useDashboard.ts frontend/src/context/SessionContext.tsx
git commit -m "feat(types): add name field to Session and SessionState"
```

---

## Task 4: NewSessionModal — optional Name field

**Files:**
- Modify: `frontend/src/components/NewSessionModal.tsx`
- Modify: `frontend/src/views/DashboardView.tsx`

- [ ] **Replace `NewSessionModal.tsx`** with the updated version that adds the optional Name field:

```typescript
import { useState } from 'react'
import { FolderOpen } from 'lucide-react'

interface NewSessionModalProps {
  onStart: (sessionId: string, workdir: string, name: string | null) => void
}

export default function NewSessionModal({ onStart }: NewSessionModalProps) {
  const [name, setName] = useState('')
  const [workdir, setWorkdir] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!workdir.trim()) return
    setLoading(true)
    setError(null)

    try {
      const body: { workdir: string; name?: string } = { workdir: workdir.trim() }
      if (name.trim()) body.name = name.trim()

      const res = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
      if (!res.ok) throw new Error(await res.text())
      const { sessionId } = (await res.json()) as { sessionId: string }
      onStart(sessionId, workdir.trim(), name.trim() || null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create session')
      setLoading(false)
    }
  }

  return (
    <div className="flex-1 flex items-center justify-center p-6">
      <div className="w-full max-w-md bg-bg-surface border border-border rounded-lg p-6 space-y-4">
        <div className="flex items-center gap-2">
          <FolderOpen size={16} className="text-accent" />
          <h3 className="text-text-primary text-sm font-medium">Start new session</h3>
        </div>
        <p className="text-text-muted text-xs">
          Enter the working directory where Claude Code will run.
        </p>
        <form onSubmit={handleSubmit} className="space-y-3">
          <div className="space-y-1">
            <label className="text-text-dim text-xs">
              Name <span className="text-text-muted">(optional)</span>
            </label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g. API refactor"
              className="w-full bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent"
            />
          </div>
          <div className="space-y-1">
            <label className="text-text-dim text-xs">Working directory</label>
            <input
              type="text"
              value={workdir}
              onChange={(e) => setWorkdir(e.target.value)}
              placeholder="/home/user/projects/my-app"
              autoFocus
              className="w-full bg-bg-panel border border-border rounded-md px-3 py-2 text-sm text-text-primary placeholder-text-dim focus:outline-none focus:border-accent"
            />
          </div>
          {error && <p className="text-status-red text-xs">{error}</p>}
          <button
            type="submit"
            disabled={loading || !workdir.trim()}
            className="w-full bg-accent hover:bg-accent-hover disabled:opacity-40 disabled:cursor-not-allowed text-black text-sm font-medium py-2 rounded-md transition-colors"
          >
            {loading ? 'Starting…' : 'Start session'}
          </button>
        </form>
      </div>
    </div>
  )
}
```

- [ ] **Update `handleSessionStart` in `DashboardView.tsx`** to accept and dispatch `name`. Find the existing function and replace it:

```typescript
  function handleSessionStart(sessionId: string, workdir: string, name: string | null) {
    dispatch({ type: 'SESSION_CREATED', sessionId, workdir, ...(name ? { name } : {}) })
    if (account?.model) dispatch({ type: 'MODEL_SET', model: account.model })
    setShowModal(false)
    refresh()
  }
```

- [ ] **Validate**

```bash
make lint
```

Expected: no errors.

- [ ] **Commit**

```bash
git add frontend/src/components/NewSessionModal.tsx frontend/src/views/DashboardView.tsx
git commit -m "feat(ui): add optional Name field to new session modal"
```

---

## Task 5: SessionList — display name with workdir fallback

**Files:**
- Modify: `frontend/src/components/SessionList.tsx`

- [ ] **Find the session title span** (currently `{lastSegment(session.workdir)}`) and update it to prefer the session name:

```typescript
                  <span className="text-text-primary text-sm font-medium truncate">
                    {session.name ?? lastSegment(session.workdir)}
                  </span>
```

That is the only change to this file.

- [ ] **Validate**

```bash
make lint
```

Expected: no errors.

- [ ] **Commit**

```bash
git add frontend/src/components/SessionList.tsx
git commit -m "feat(ui): show session name in list, fall back to workdir segment"
```

---

## Task 6: SessionHeader — name display + rename button + edit mode

**Files:**
- Modify: `frontend/src/components/SessionHeader.tsx`

- [ ] **Replace `SessionHeader.tsx`** with the full updated version below. Key additions: `sessionName` and `onRename` optional props; name display before workdir chip; Pencil rename button; inline edit mode with Save/Cancel/Enter/Escape.

```typescript
import { useState, useEffect } from 'react'
import { Plus, List, Square, Pencil } from 'lucide-react'
import { useSession } from '../context/SessionContext'

interface SessionHeaderProps {
  onNewSession: () => void
  onStopSession: () => void
  onSessionsList: () => void
  totalTokens: number
  sessionStartedAt: number | null
  sessionName?: string | null
  onRename?: (name: string) => Promise<void>
}

function formatModelName(model: string): string {
  const s = model.replace(/^claude-/, '').replace(/-\d{8}$/, '')
  const m = s.match(/^(opus|sonnet|haiku)-(\d+)-(\d+)/)
  if (m) return `${m[1].charAt(0).toUpperCase() + m[1].slice(1)} ${m[2]}.${m[3]}`
  const m2 = s.match(/^(\d+)-(?:(\d+)-)?(\w+)$/)
  if (m2) {
    const family = m2[3].charAt(0).toUpperCase() + m2[3].slice(1)
    return m2[2] ? `${family} ${m2[1]}.${m2[2]}` : `${family} ${m2[1]}`
  }
  return s.charAt(0).toUpperCase() + s.slice(1)
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`
  return String(n)
}

function formatCreatedAt(ts: number): string {
  const d = new Date(ts)
  const now = new Date()
  const sameDay = d.toDateString() === now.toDateString()
  const time = d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  if (sameDay) return time
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' }) + ' ' + time
}

function formatDuration(ms: number): string {
  const totalSec = Math.floor(ms / 1000)
  const h = Math.floor(totalSec / 3600)
  const m = Math.floor((totalSec % 3600) / 60)
  const s = totalSec % 60
  if (h > 0) return `${h}h ${m}m ${s}s`
  if (m > 0) return `${m}m ${s}s`
  return `${s}s`
}

interface StatChipProps { label: string; value: string; valueClass?: string }
function StatChip({ label, value, valueClass = 'text-text-secondary' }: StatChipProps) {
  return (
    <span className="flex items-center gap-1 text-xs">
      <span className="text-text-dim uppercase tracking-wider">{label}</span>
      <span className={`font-medium ${valueClass}`}>{value}</span>
    </span>
  )
}

export default function SessionHeader({
  onNewSession,
  onStopSession,
  onSessionsList,
  totalTokens,
  sessionStartedAt,
  sessionName,
  onRename,
}: SessionHeaderProps) {
  const { state } = useSession()
  const [, setTick] = useState(0)
  const [renaming, setRenaming] = useState(false)
  const [nameInput, setNameInput] = useState('')
  const [renameError, setRenameError] = useState<string | null>(null)
  const [renameSaving, setRenameSaving] = useState(false)

  // Tick only while Claude is actively running — idle time doesn't count
  useEffect(() => {
    if (state.wsState !== 'running') return
    const timer = setInterval(() => setTick((t) => t + 1), 1000)
    return () => clearInterval(timer)
  }, [state.wsState])

  const workingMs =
    state.workingTimeMs +
    (state.runningStartedAt != null ? Date.now() - state.runningStartedAt : 0)

  function startRename() {
    setNameInput(sessionName ?? '')
    setRenameError(null)
    setRenaming(true)
  }

  function cancelRename() {
    setRenaming(false)
    setRenameError(null)
  }

  async function saveRename() {
    if (!onRename) return
    setRenameSaving(true)
    setRenameError(null)
    try {
      await onRename(nameInput.trim())
      setRenaming(false)
    } catch (err) {
      setRenameError(err instanceof Error ? err.message : 'Failed to rename')
    } finally {
      setRenameSaving(false)
    }
  }

  function onRenameKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (e.key === 'Enter') { e.preventDefault(); saveRename() }
    if (e.key === 'Escape') cancelRename()
  }

  return (
    <div className="px-3 py-2 border-b border-border-subtle bg-bg-surface flex items-center gap-3 flex-shrink-0 flex-wrap">
      {/* Status + name/workdir — or inline rename form */}
      {renaming ? (
        <div className="flex items-center gap-1.5 min-w-0 flex-wrap">
          <input
            value={nameInput}
            onChange={(e) => setNameInput(e.target.value)}
            onKeyDown={onRenameKeyDown}
            autoFocus
            placeholder="Session name (optional)"
            className="bg-bg-panel border border-accent rounded px-2 py-0.5 text-xs text-text-primary placeholder-text-dim focus:outline-none w-44"
          />
          <button
            onClick={saveRename}
            disabled={renameSaving}
            className="text-xs text-accent hover:text-accent-hover disabled:opacity-40 transition-colors"
          >
            {renameSaving ? '…' : 'Save'}
          </button>
          <button
            onClick={cancelRename}
            className="text-xs text-text-dim hover:text-text-secondary transition-colors"
          >
            ✕
          </button>
          {renameError && <span className="text-status-red text-xs">{renameError}</span>}
        </div>
      ) : (
        <div className="flex items-center gap-2 min-w-0">
          <div className={`w-2 h-2 rounded-full flex-shrink-0 ${
            state.wsState === 'running' ? 'bg-status-green' :
            state.wsState === 'idle'    ? 'bg-status-green/40' :
            state.wsState === 'error'   ? 'bg-status-red' : 'bg-text-dim'
          }`} />
          {sessionName && (
            <span className="text-text-primary text-xs font-medium truncate max-w-[160px]">
              {sessionName}
            </span>
          )}
          {state.workdir && (
            <span className="text-text-secondary text-xs bg-bg-elevated border border-border-subtle px-2 py-0.5 rounded truncate max-w-xs">
              {state.workdir}
            </span>
          )}
          {state.model && (
            <span className="text-accent text-xs font-medium hidden sm:block">{formatModelName(state.model)}</span>
          )}
        </div>
      )}

      {/* Session stat chips */}
      <div className="flex items-center gap-3 text-xs border-l border-border-subtle pl-3">
        {sessionStartedAt != null && (
          <StatChip label="created" value={formatCreatedAt(sessionStartedAt)} />
        )}
        {workingMs > 0 && (
          <StatChip label="dur" value={formatDuration(workingMs)} valueClass="text-status-green" />
        )}
        <StatChip label="tokens" value={formatTokens(totalTokens)} />
      </div>

      {/* Action buttons */}
      <div className="flex items-center gap-2 ml-auto">
        {onRename && !renaming && (
          <button
            onClick={startRename}
            className="flex items-center gap-1.5 text-text-muted hover:text-text-secondary text-xs bg-bg-elevated border border-border-subtle hover:border-border px-2 py-1 rounded transition-colors"
          >
            <Pencil size={11} />
            Rename
          </button>
        )}
        <button
          onClick={onNewSession}
          className="flex items-center gap-1.5 text-text-muted hover:text-text-secondary text-xs bg-bg-elevated border border-border-subtle hover:border-border px-2 py-1 rounded transition-colors"
        >
          <Plus size={11} />
          New session
        </button>
        <button
          onClick={onStopSession}
          className="flex items-center gap-1.5 text-text-muted hover:text-status-red text-xs bg-bg-elevated border border-border-subtle hover:border-status-red px-2 py-1 rounded transition-colors"
        >
          <Square size={11} />
          Stop session
        </button>
        <button
          onClick={onSessionsList}
          className="flex items-center gap-1.5 text-text-muted hover:text-text-secondary text-xs bg-bg-elevated border border-border-subtle hover:border-border px-2 py-1 rounded transition-colors"
        >
          <List size={11} />
          Sessions list
        </button>
      </div>
    </div>
  )
}
```

- [ ] **Validate**

```bash
make lint
```

Expected: no errors.

- [ ] **Commit**

```bash
git add frontend/src/components/SessionHeader.tsx
git commit -m "feat(ui): session name display and inline rename in SessionHeader"
```

---

## Task 7: DashboardView — wire name through resume + implement onRename

**Files:**
- Modify: `frontend/src/views/DashboardView.tsx`

- [ ] **Update `handleResume`** to pass the session name into the dispatch:

```typescript
  function handleResume(session: Session) {
    dispatch({ type: 'RESUME_SESSION', id: session.id, workdir: session.workdir, ...(session.name ? { name: session.name } : {}) })
    if (account?.model) dispatch({ type: 'MODEL_SET', model: account.model })
  }
```

- [ ] **Add `handleRenameSession`** after `handleResume`:

```typescript
  async function handleRenameSession(name: string) {
    if (!state.sessionId) return
    const res = await fetch(`/api/sessions/${state.sessionId}`, {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: name || null }),
    })
    if (!res.ok) throw new Error(await res.text())
    dispatch({ type: 'SESSION_RENAMED', name: name || null })
    refresh()
  }
```

- [ ] **Update the `SessionHeader` JSX** to pass `sessionName` and `onRename`:

```typescript
          <SessionHeader
            onNewSession={handleNewSession}
            onStopSession={handleStopSession}
            onSessionsList={handleSessionsList}
            onRename={handleRenameSession}
            sessionName={state.name}
            totalTokens={(usage?.totals.inputTokens ?? 0) + (usage?.totals.outputTokens ?? 0)}
            sessionStartedAt={activeSession?.started_at ?? null}
          />
```

- [ ] **Validate**

```bash
make lint
```

Expected: no errors.

- [ ] **Commit**

```bash
git add frontend/src/views/DashboardView.tsx
git commit -m "feat(ui): wire session name through resume and rename flows"
```

---

## Self-Review Checklist

- **DB migration** — Task 1 ✓
- **POST /api/sessions accepts `name`** — Task 2 ✓
- **PATCH /api/sessions/:id** — Task 2 ✓
- **`Session` interface `name` field** — Task 3 ✓
- **`SessionState.name`, `SESSION_RENAMED` action** — Task 3 ✓
- **NewSessionModal Name field, tab order Name→Workdir→Submit** — Task 4 ✓ (autoFocus is on workdir, name comes first in DOM so tab order is natural)
- **`handleSessionStart` dispatches name** — Task 4 ✓
- **SessionList shows `name ?? lastSegment`** — Task 5 ✓
- **SessionHeader name display before workdir chip** — Task 6 ✓
- **Pencil Rename button left of action group** — Task 6 ✓
- **Edit mode: input + Save + Cancel + Enter/Escape** — Task 6 ✓
- **Error display in edit mode** — Task 6 ✓
- **Empty name clears (stored as null)** — Task 2 (`name.trim() || null`) + Task 7 (`name || null`) ✓
- **`handleResume` passes name** — Task 7 ✓
- **`handleRenameSession` PATCH + dispatch** — Task 7 ✓
