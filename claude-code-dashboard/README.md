# Claude Code Dashboard

Web dashboard for [Claude Code](https://claude.ai/code) — view account info, usage graphs, and manage Claude sessions from the browser.

**Stack:** Fastify + TypeScript backend · Vite + React + Tailwind CSS frontend · SQLite session cache · xterm.js terminal · recharts usage graph

In production the backend serves the built frontend on a single port. In development Vite runs separately with HMR and proxies API/WS calls to the backend.

## Prerequisites

- Node.js ≥ 20
- `claude` CLI [installed](https://claude.ai/code) and authenticated on the host

## Running

### Option 1 — Production script (Linux / macOS)

```bash
./run.sh          # install deps, build, start
# or: make run
```

```powershell
# Windows (PowerShell 7)
.\run.ps1
```

Both scripts create `backend/.env` from `.env.example` on first run, load it, build both packages, and start the server.

Dashboard available at **http://localhost:9998** (or `PORT` from your `.env`).

### Option 2 — Docker

```bash
docker-compose up
```

Dashboard available at **http://localhost:8080**.

`docker-compose.yml` mounts `~/.claude` (read-only) and `~/projects` (read-write) from the host.

To build the image locally instead of pulling:

```bash
docker build -t claude-code-dashboard .
docker-compose up
```

### Option 3 — Systemd user service (Linux)

Installs the dashboard as a persistent user-level service that starts automatically at boot (with lingering):

```bash
make service-install
# or to skip the build step if already built:
make service-install ARGS=--skip-build
```

Reads `backend/.env` and generates `~/.config/systemd/user/claude-code-dashboard.env`. The service runs under your user account — no root required.

```bash
make service-uninstall          # stop, disable, remove unit file
systemctl --user status claude-code-dashboard
journalctl --user -u claude-code-dashboard -f
```

## Development

```bash
cp .env.example backend/.env    # configure environment
make install                    # npm install in both packages
make dev                        # backend :9998 + frontend :9999 (HMR)
```

Vite proxies `/api` and `/ws` to the backend and binds to `0.0.0.0`, so the dev server is reachable from the LAN at `http://<host-ip>:9999`.

## Configuration

```bash
cp .env.example backend/.env
```

| Variable | Default | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | — | Optional — enables billing data in the Usage view |
| `CLAUDE_BIN` | `claude` | Path to the `claude` binary |
| `PORT` | `9998` | Backend listen port |
| `FRONTEND_ORIGIN` | `http://localhost:9999` | Allowed CORS origin (dev only) |

## Commands

```bash
make dev              # start backend + frontend dev servers concurrently
make build            # compile backend TS + Vite production bundle
make install          # npm install in both packages
make lint             # tsc --noEmit in both packages
make clean            # remove dist/ directories
make run              # production build + start (wraps run.sh)

make service-install  # install as systemd user service
make service-uninstall
make service-test     # run service install test suite
```

## Project structure

```
backend/src/
  server.ts              # Fastify entry (CORS, WebSocket, static serving)
  routes/
    account.ts           # GET /api/account
    usage.ts             # GET /api/usage
    sessions.ts          # GET/POST /api/sessions, POST /:id/stop, PATCH /:id
    settings.ts          # GET/POST /api/settings
  ws/
    session.ts           # WebSocket handler — headless claude subprocess
  services/
    claudeAccount.ts     # Parse ~/.claude/settings.json + claude --version
    localLogs.ts         # Parse ~/.claude/projects/**/*.jsonl for usage
    anthropicApi.ts      # Anthropic billing API client
    oauthUsage.ts        # OAuth-based usage data
    usageCache.ts        # SQLite cache with 1 hr TTL
  db/
    schema.ts            # better-sqlite3 init + migrations

frontend/src/
  App.tsx                # HashRouter + layout shell
  components/
    StatsStrip.tsx       # Stat chips + settings icon
    SessionList.tsx      # Session rows (resume / delete / active badge)
    SessionHeader.tsx    # Workdir + model + New Session button
    UsageChart.tsx       # recharts AreaChart
    MessageList.tsx      # Chat message history
    AssistantMessage.tsx # Markdown + syntax-highlighted response
    TerminalDrawer.tsx   # xterm.js drawer (CSS-toggled, never unmounted)
    NewSessionModal.tsx  # Directory picker for new sessions
    ChatInput.tsx        # Textarea (Enter = send, Shift+Enter = newline)
  views/
    DashboardView.tsx    # StatsStrip → collapsible chart → session area
    SettingsView.tsx     # API key config
  hooks/
    useWebSocket.ts      # WS with exponential-backoff reconnect
    useDashboard.ts      # Parallel fetch account + usage + sessions, 60 s refresh
  context/
    SessionContext.tsx   # Active session state + dispatch

scripts/
  install-service.sh                      # systemd user service installer
  uninstall-service.sh                    # systemd user service remover
  claude-code-dashboard.service.template  # unit file template
  test-service-install.sh                 # installer test suite
```
