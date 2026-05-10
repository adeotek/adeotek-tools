# Headless Session Layer Design

**Date:** 2026-05-10
**Branch:** feat/claude-code-dashboard
**Status:** Approved

## Problem

The current backend spawns `claude` as a PTY process and parses heuristic markers (`❯❯❯` / `❖`) from raw ANSI output to extract assistant messages. These markers no longer appear in Claude Code v2.1.138+, making the chat view completely broken. The approach is fundamentally fragile — it breaks whenever Claude Code changes its TUI.

## Goal

Replace the PTY + ANSI-scraping layer with `claude --print --output-format stream-json`, which emits structured JSON events. Chat messages are extracted reliably from `assistant` events; the terminal log is fed formatted text derived from the same event stream.

## Architecture Overview

```
User message (WS frame type:'chat')
    │
    ▼
ActiveSession.sendMessage()
    │
    ▼
spawn: claude --print --output-format stream-json --verbose
              --include-partial-messages --input-format stream-json
              [--resume <claude_session_id>]
              [--dangerously-skip-permissions]   ← configurable
    │
    ├─ stdin ← {"type":"user","message":{"role":"user","content":"..."}}
    │
    └─ stdout (JSONL) → EventParser
         ├─ tool_use / tool_result  → type:'output' (terminal log)
         ├─ partial text chunk      → type:'output' (terminal log)
         ├─ final text (end_turn)   → type:'output' + type:'message' (chat)
         ├─ system/init             → type:'output' session banner
         └─ result                  → capture claude_session_id, persist to DB
```

One subprocess per user turn. The process exits naturally after the `result` event. Session continuity across turns is maintained by storing `claude_session_id` (from the `result` event) and passing `--resume <claude_session_id>` on every turn after the first.

## Backend — ActiveSession Class

Replaces `node-pty` with `child_process.spawn`.

### State

```typescript
class ActiveSession {
  readonly id: string                      // dashboard session UUID
  private claudeSessionId: string | null   // Claude-assigned; used for --resume
  private bypassPermissions: boolean       // from DB settings
  private currentProc: ChildProcess | null // non-null only during a turn
  private isRunning: boolean
  private sockets: Set<WebSocket>
  private idleTimer: NodeJS.Timeout | null
}
```

### Turn Lifecycle

1. `sendMessage(text)` called
2. `isRunning = true`; broadcast `{type:'status', state:'running'}`
3. Spawn subprocess (see flags below)
4. Write user message JSON to stdin; close stdin
5. Stream stdout line-by-line through `EventParser`
6. On `result` event: save `claude_session_id` to DB; broadcast `{type:'status', state:'idle'}`; set `isRunning = false`; `currentProc = null`
7. Process exits

### Spawn Flags

```
claude
  --print
  --output-format stream-json
  --verbose
  --include-partial-messages
  --input-format stream-json
  [--resume <claude_session_id>]           # omitted on first turn
  [--dangerously-skip-permissions]         # when bypassPermissions=true
  --cwd <workdir>                          # via spawn options, not flag
```

### EventParser — stream-json → WS frames

| stream-json event | Condition | WS frame(s) emitted |
|---|---|---|
| `system` / `init` | always | `type:'output'` — session-start banner line |
| `assistant` | content has `tool_use` block | `type:'output'` — `⚙ <ToolName>: <input summary>` |
| `user` | content has `tool_result` | `type:'output'` — `→ <result text, truncated to 500 chars>` |
| `assistant` | partial text (`stop_reason: null`) | `type:'output'` — raw text chunk |
| `assistant` | final text (`stop_reason: 'end_turn'`) | `type:'output'` — final text chunk + `type:'message'` — full content + DB insert |
| `result` | always | capture `session_id`; `UPDATE sessions SET claude_session_id` |

### kill()

Sends `SIGTERM` to `currentProc` if non-null. Sets `ended_at` in DB.

## DB Schema Change

One new column on the `sessions` table:

```sql
ALTER TABLE sessions ADD COLUMN claude_session_id TEXT;
```

Migration runs at startup in `db/schema.ts` using a `try/catch` around the `ALTER TABLE` (SQLite ignores duplicate column errors when caught explicitly, or use a `PRAGMA table_info` check).

## WS Protocol — No Changes

The backend→frontend frames keep identical shapes:

| Frame | Purpose |
|---|---|
| `{type:'output', data:string}` | Terminal log line (now formatted text, not raw ANSI) |
| `{type:'message', role:'assistant', content:string}` | Chat view — final response only |
| `{type:'status', state:'running'|'idle'|'error'|'disconnected'}` | Session state indicator |
| `{type:'history', messages:[...]}` | Sent on WS connect; unchanged |

Incoming frames `type:'input'` and `type:'resize'` are silently ignored (PTY-only; can be cleaned up from the frontend later).

## Frontend Changes

### Terminal Drawer

No component changes. `onOutput` callback in `DashboardView` continues to feed `type:'output'` data to xterm.js. Content is now formatted text lines instead of raw ANSI bytes — xterm renders them cleanly without control-character noise.

### Settings View

Add a **Bypass permissions** toggle:

- Label: "Bypass tool permissions"
- Description: "When enabled, Claude Code runs with `--dangerously-skip-permissions` (same trust as running `claude` directly). When disabled, writes outside the workspace are sandboxed."
- Default: enabled
- Backed by: `GET /api/settings` and `POST /api/settings` endpoints
- Storage: new `settings` table in SQLite (`key TEXT PRIMARY KEY, value TEXT`)

## Permissions — Rationale

Interactive per-tool permission approval is not available through the `--print` / `stream-json` API. Tool calls are executed automatically inside the subprocess; there are no `permission_request` events to intercept. The two available modes are:

- **Bypass** (`--dangerously-skip-permissions`): same trust as running `claude` in your terminal. All tools execute. Default for a local dashboard on your own machine.
- **Sandbox** (default Claude Code behavior): writes outside the workspace are auto-blocked; Claude is notified and adapts. Bash commands in the workspace directory still execute.

Both modes surface all tool calls in the terminal log so the user can see what is happening and stop the session if needed.

## Removed

- `node-pty` dependency (removed from `backend/package.json`)
- `ANSI_RE`, `messageBuffer`, `inAssistantBlock`, marker-parsing logic
- `TerminalDrawer` PTY resize wiring (`onTerminalMounted` call in `DashboardView` — was already `void`-ed)

## Files Changed

| File | Change |
|---|---|
| `backend/src/ws/session.ts` | Full rewrite — `ActiveSession` and `SessionManager` |
| `backend/src/routes/sessions.ts` | Add `GET/POST /api/settings` or new `settings.ts` route |
| `backend/src/db/schema.ts` | Add `claude_session_id` column + `settings` table |
| `backend/package.json` | Remove `node-pty`; no new deps needed (`child_process` is built-in) |
| `frontend/src/views/DashboardView.tsx` | Remove PTY resize wiring |
| `frontend/src/views/SettingsView.tsx` | Add bypass-permissions toggle |
