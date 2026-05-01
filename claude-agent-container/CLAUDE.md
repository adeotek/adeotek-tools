# CLAUDE.md

## Project

`claude-agent-container` — a sandboxed Docker container that runs Claude Code CLI as a
fully autonomous agent. Receives a task, clones a GitHub repo, creates a branch, does
the work, and opens a Pull Request.

Part of the `adeotek-tools` monorepo. This project is a Docker image only — no runtime
dependencies on other projects in the monorepo.

## Build & run commands

```bash
make build                                       # build image locally
make push                                        # tag + push to ghcr.io/adeotek/claude-agent
make run REPO_URL=https://github.com/org/repo \
         TASK="add a CONTRIBUTING.md"            # run a task (uses ~/.claude.json for auth)
make clean                                       # remove local image
docker compose up                                # alternative: run via compose with task.md mount
```

## Project structure

| Path | Purpose |
|------|---------|
| `Dockerfile` | Image — 8 ordered layers (system → Node.js → .NET → Python → gh → LSPs → claude → config) |
| `entrypoint.sh` | 6-phase orchestration: credentials, git, plugins, task setup, claude, report+PR |
| `config/settings.json` | Claude Code settings baked into image (`autoUpdates: false`, full permissions) |
| `docker-compose.yml` | Local dev/test helper; defines `claude-agent-state` named volume |
| `Makefile` | Shortcuts for build, push, run, clean |

## Key design decisions

- **Single-use**: one `docker run` = one task; container exits when done
- **Task input**: `docker run claude-agent "prompt"` (arg) OR mount `/task/task.md` (file takes precedence)
- **Claude auth**: Pro/Max OAuth via `/root/.claude.json` or `/run/secrets/claude_credentials`;
  API key via `ANTHROPIC_API_KEY`
- **GitHub auth**: `GITHUB_TOKEN` env var or `/run/secrets/github_token`
- **Plugins**: installed on first run via `entrypoint.sh`; cached with marker file in
  named volume `/root/.claude-agent-state/` (avoids shadowing `/root/.claude/settings.json`)
- **Auto-updates**: disabled in `config/settings.json` for deterministic behavior
- **Container instructions**: mount `/agent/instructions.md` to inject org-wide context
  into every task prompt

## Environment variable reference

See `README.md` for the full table of env vars, volume mounts, and usage examples.
