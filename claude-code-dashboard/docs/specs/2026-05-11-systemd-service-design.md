# Systemd User Service Support

**Date:** 2026-05-11
**Status:** Approved

## Overview

Add a `make service-install` target that builds the dashboard and installs it as a systemd user service, enabling the dashboard to start automatically on boot (with lingering) without Docker.

## Approach

Shell script + Makefile entry point (Approach B). Complex install logic lives in dedicated shell scripts; the Makefile provides the user-facing entry points.

## New Files

```
scripts/
  claude-code-dashboard.service.template   # committed unit file template (__INSTALL_DIR__ placeholder)
  install-service.sh                       # install logic
  uninstall-service.sh                     # uninstall/cleanup logic
```

## install-service.sh Flow

1. **Preflight checks** — verify `node`, `npm`, and `claude` are on PATH; abort with a clear message if any are missing
2. **Build** — run `make build`; skip if both `backend/dist` and `frontend/dist` already exist and `--skip-build` flag is passed
3. **Create config dir** — `mkdir -p ~/.config/systemd/user/`
4. **Generate env file** — write `~/.config/systemd/user/claude-code-dashboard.env` from `backend/.env`; copy only known vars (`ANTHROPIC_API_KEY`, `CLAUDE_BIN`, `PORT`), using defaults for any that are absent
5. **Stamp unit file** — substitute `__INSTALL_DIR__` (repo absolute path) in the template using `sed`; write result to `~/.config/systemd/user/claude-code-dashboard.service`
6. **Reload daemon** — `systemctl --user daemon-reload`
7. **Enable and start** — `systemctl --user enable --now claude-code-dashboard`
8. **Enable linger** — `loginctl enable-linger "$USER"` so the service survives logout; print an explanatory note to the user
9. **Print URL** — output the dashboard address (`http://localhost:<PORT>`)

## uninstall-service.sh Flow

1. Stop and disable the service — `systemctl --user disable --now claude-code-dashboard`
2. Remove the unit file — `~/.config/systemd/user/claude-code-dashboard.service`
3. Prompt before removing the env file — it may have been customised
4. `systemctl --user daemon-reload`

Lingering is **not** disabled on uninstall — it may be relied on by other user services.

## Makefile Targets

```makefile
service-install:
	@bash scripts/install-service.sh

service-uninstall:
	@bash scripts/uninstall-service.sh
```

Both scripts are called from the repo root so `$(CURDIR)` / `pwd` resolves the install directory correctly.

## Unit File Template

```ini
[Unit]
Description=Claude Code Dashboard
After=network.target

[Service]
Type=simple
WorkingDirectory=__INSTALL_DIR__
ExecStart=node __INSTALL_DIR__/backend/dist/server.js
EnvironmentFile=%h/.config/systemd/user/claude-code-dashboard.env
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
```

**Key decisions:**
- `%h` — systemd's home-dir specifier; avoids hardcoding `/home/<user>`
- `ExecStart` uses `node` directly (not `npm start`) so systemd tracks the real PID
- `WorkingDirectory` is the repo root so relative paths (SQLite DB, frontend dist) resolve correctly
- No `User=` directive — user service inherits the invoking user automatically

## Env File Format

Generated at `~/.config/systemd/user/claude-code-dashboard.env`:

```
ANTHROPIC_API_KEY=<value from backend/.env or empty>
CLAUDE_BIN=<value from backend/.env or claude>
PORT=<value from backend/.env or 9998>
```

Only these three vars are written. `FRONTEND_ORIGIN` is dev-only and intentionally excluded.

## Out of Scope

- System-level service (runs as root or a dedicated user) — user service covers the primary use case
- Automatic linger disable on uninstall — linger may be needed by other services
- Service update workflow (`make service-update`) — can be a follow-up
