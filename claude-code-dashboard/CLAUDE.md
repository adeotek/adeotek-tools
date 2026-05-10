# claude-code-dashboard

Web dashboard for Claude Code — displays account info, usage graphs, and acts as a browser-based proxy to the local `claude` CLI.

## Stack

- **Backend**: Fastify + TypeScript (Node.js), `@fastify/websocket`, `@fastify/static`, `node-pty` for PTY process management, `better-sqlite3` for session/usage caching
- **Frontend**: Vite + React + TypeScript, Tailwind CSS (dark theme), `@xterm/xterm` (terminal drawer), `recharts` (usage chart)

In production (Docker or after `make build`), the backend serves the built frontend via `@fastify/static` — single port, no nginx. In development, Vite runs separately with its own dev server (HMR) and proxies API/WS to the backend.

## Prerequisites

- Node.js >= 20
- `claude` CLI installed and authenticated on the host

## Development

```bash
# Install deps in both packages
make install

# Start backend (:3001) and frontend (:5173) concurrently
make dev
```

Vite proxies `/api` and `/ws` to the backend, and binds to `0.0.0.0` — reachable from LAN at `http://<host-ip>:5173`.

## Environment

```bash
cp .env.example backend/.env
```

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | — | Optional; enables billing data in Usage view |
| `CLAUDE_BIN` | `claude` | Path to the `claude` binary |
| `PORT` | `3001` | Backend listen port |
| `FRONTEND_ORIGIN` | `http://localhost:5173` | CORS allowed origin (dev only) |

## Commands

```bash
make dev          # start both servers (backend + frontend)
make build        # compile backend TS + build frontend Vite bundle
make lint         # tsc --noEmit in both packages
make install      # npm install in both packages
make clean        # remove dist/ directories
```

## Project structure

```
backend/src/
  server.ts              # Fastify entry point (CORS, WebSocket, routes)
  routes/
    account.ts           # GET /api/account
    usage.ts             # GET /api/usage
    sessions.ts          # GET/POST /api/sessions, POST /api/sessions/:id/stop
  ws/
    session.ts           # WebSocket handler + node-pty lifecycle
  services/
    claudeAccount.ts     # Parse ~/.claude/settings.json + run claude --version
    localLogs.ts         # Parse ~/.claude/projects/**/*.jsonl for usage
    anthropicApi.ts      # Anthropic billing API client
    usageCache.ts        # SQLite read/write with 1hr TTL
  db/
    schema.ts            # better-sqlite3 init + table migrations

frontend/src/
  App.tsx                # HashRouter + layout shell (no sidebar — full width)
  components/
    StatsStrip.tsx       # Compact stat chips row + settings icon (top of dashboard)
    SessionList.tsx      # Session rows with resume/delete, active/ended badges
    SessionHeader.tsx    # Active workdir + model + New Session button
    InfoCard.tsx         # Account info card
    StatCard.tsx         # Usage stat card
    UsageChart.tsx       # recharts AreaChart
    MessageList.tsx      # Chat message history
    AssistantMessage.tsx # Markdown + syntax-highlighted response
    UserMessage.tsx      # User chat bubble
    TerminalDrawer.tsx   # xterm.js — CSS-toggled, never unmounted
    NewSessionModal.tsx  # Directory picker shown on session start
    ChatInput.tsx        # Textarea input (Enter=send, Shift+Enter=newline)
  views/
    DashboardView.tsx    # Main view: StatsStrip → collapsible UsageChart → session area
    SettingsView.tsx     # API key config + Back to Dashboard link
  hooks/
    useWebSocket.ts      # WS connection with exponential-backoff reconnect
    useAccount.ts        # Fetches GET /api/account
    useUsage.ts          # Fetches GET /api/usage
    useDashboard.ts      # Parallel fetch account+usage+sessions, 60s auto-refresh
  context/
    SessionContext.tsx   # Active session state + dispatch
```

## Docker

```bash
docker build -t claude-code-dashboard .
docker-compose up
# Dashboard at http://localhost:8080
```

`docker-compose.yml` mounts `~/.claude` (read-only) and `~/projects` (read-write) from the host.
