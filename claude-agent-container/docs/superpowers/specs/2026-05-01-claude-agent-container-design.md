# Claude Agent Container — Design Spec & Implementation Plan

## Context

This project creates a sandboxed Docker container that runs Claude Code CLI as a fully
autonomous agent. The container receives a task, clones a GitHub repository, creates a
branch, works on the task, tests the results, commits, pushes, and opens a Pull Request —
all without human interaction. It is the self-hosted equivalent of Claude Code cloud
agents, designed for use in CI pipelines, scripting, or any automated workflow.

Project path: `claude-agent-container/` within the `adeotek-tools` monorepo.

---

## Design Spec

### Architecture

Single-use autonomous agent runner. One `docker run` = one task. Boots, works, reports,
exits.

```
┌─────────────────────────────────────────────────────────┐
│                  claude-agent container                  │
│                                                         │
│  entrypoint.sh $@                                       │
│       │                                                 │
│       ├─ 1. Resolve credentials  ◄── /run/secrets/      │
│       │                          ◄── /root/.claude.json  │
│       │                          ◄── ANTHROPIC_API_KEY   │
│       ├─ 2. Configure git identity (GIT_AUTHOR_*)       │
│       ├─ 3. Install plugins (first run only)            │
│       ├─ 4. Read task + clone repo + create branch      │
│       ├─ 5. Run: timeout + claude --max-turns ...       │
│       │         (with optional instructions prepended)  │
│       └─ 6. Write /output/report.md + push + open PR   │
│                                                         │
└─────────────────────────────────────────────────────────┘
         ▲                            ▼
    /task/task.md              /output/report.md
    /run/secrets/              git push → GitHub PR
    /agent/instructions.md
```

**Task input (first match wins):**
1. `/task/task.md` (mounted file, takes precedence)
2. `docker run claude-agent "inline prompt"` (positional argument `$1`)

**Authentication (first match wins):**
1. `/run/secrets/claude_credentials` → copied to `/root/.claude.json` (Pro/Max OAuth)
2. `/root/.claude.json` volume-mounted (Pro/Max OAuth)
3. `ANTHROPIC_API_KEY` env var (Anthropic API key)

**GitHub auth:** `GITHUB_TOKEN` env var or `/run/secrets/github_token`

**Container-level instructions:** Optional file at `$AGENT_INSTRUCTIONS_FILE`
(default: `/agent/instructions.md`). Contents are prepended to the task prompt at
runtime. Enables org-wide coding standards or policies without touching the cloned repo.

---

### Dockerfile Layers

Base image: `debian:bookworm-slim`

Ordered from most stable → most volatile for optimal layer caching:

| Layer | Contents |
|-------|----------|
| ① System packages | curl, wget, git, ca-certificates, gnupg, build-essential, unzip, jq, procps |
| ② Node.js 24 LTS | via nodesource setup script |
| ③ .NET SDK 9 | via `dotnet-install.sh` → `/usr/local/dotnet`; `ENV DOTNET_ROOT + PATH` |
| ④ Python 3 + uv | python3, python3-pip, python3-venv (apt); uv (astral.sh installer) |
| ⑤ GitHub CLI | via official GitHub apt repo |
| ⑥ Language servers | `npm install -g`: typescript, typescript-language-server, pyright, @angular/cli; `dotnet tool install -g`: csharp-ls |
| ⑦ Claude Code native | `curl -fsSL https://claude.ai/install.sh \| bash` |
| ⑧ Config + entrypoint | `COPY config/settings.json /root/.claude/settings.json`; `COPY entrypoint.sh /entrypoint.sh` |

Auto-updates disabled via `"autoUpdates": false` in `config/settings.json`.
Plugins are **not** installed at build time — deferred to first run in `entrypoint.sh`.

---

### entrypoint.sh — Six Phases

**Phase 1 — Credentials**
- Check `/run/secrets/claude_credentials` → copy to `/root/.claude.json`
- Else check `/root/.claude.json` already present → use as-is
- Else check `ANTHROPIC_API_KEY` → passed through to claude
- None found → print error, `exit 1`
- Resolve `GITHUB_TOKEN` from env var or `/run/secrets/github_token`
- Missing GitHub token → print error, `exit 1`
- Run `echo "$GITHUB_TOKEN" | gh auth login --with-token`

**Phase 2 — Git identity**
- `GIT_AUTHOR_NAME` required → `exit 1` if absent
- `GIT_AUTHOR_EMAIL` required → `exit 1` if absent
- `git config --global user.name "$GIT_AUTHOR_NAME"`
- `git config --global user.email "$GIT_AUTHOR_EMAIL"`
- `git config --global push.autoSetupRemote true`

**Phase 3 — Plugin installation (first run only)**
- Marker file: `/root/.claude-agent-state/.plugins-installed`
  (separate directory to avoid conflicting with `/root/.claude/` image contents)
- If absent: install each plugin in `$CLAUDE_PLUGINS` (default: all 12)
- Write marker after success

**Phase 4 — Task + repo setup**
- Task resolution (first match wins):
  1. `/task/task.md` present → `TASK=$(cat /task/task.md)`
  2. `$1` (docker run arg) → `TASK="$1"`
  - Neither → print error, `exit 1`
- `REPO_URL` required → `exit 1` if absent
- `BASE_BRANCH` (default: `main`)
- `BRANCH_NAME` (default: `claude-agent/YYYYMMDD-HHmmss`)
- `git clone "$REPO_URL" /workspace/repo && cd /workspace/repo`
- `git checkout -b "$BRANCH_NAME"`
- Build full prompt: prepend `$AGENT_INSTRUCTIONS_FILE` contents if file exists

**Phase 5 — Run Claude Code**
```bash
timeout $((TASK_TIMEOUT_MINUTES * 60)) \
  claude \
    --dangerously-skip-permissions \
    --max-turns "$CLAUDE_MAX_TURNS" \
    -p "$FULL_PROMPT"
```
Exit codes captured: `0`=success, `1`=claude failure, `124`=timeout

**Phase 6 — Report + push (always runs)**
- `git add -A && git diff --staged --quiet || git commit -m "chore: claude-agent task"`
- `git push` (best-effort)
- If commits exist on branch:
  - success → `gh pr create --title "..." --body "..."`
  - failure/timeout → `gh pr create --draft --title "[Draft] ..." --body "..."`
- Write `/output/report.md` (see Report Format below)
- Exit `0` (success), `1` (failure), or `2` (timeout)

---

### Report Format (`/output/report.md`)

```markdown
# Claude Agent Report

- **Status:** success | failure | timeout
- **Task:** <first 200 chars of task>
- **Repo:** <REPO_URL>
- **Branch:** <branch name>
- **PR:** <URL or "none">
- **Turns used:** 14 / 50
- **Elapsed:** 3m 12s

## Changes

2 files changed, 47 insertions(+), 1 deletion(-)
 CONTRIBUTING.md | 46 ++++++++++++++++++++++++++++++
 README.md       |  2 +-

### Commits
- a3f1c2e Add CONTRIBUTING.md with contribution guidelines
- 9b2d441 Update README.md to link to CONTRIBUTING.md
```

---

### Environment Variables

| Variable | Default | Required |
|---|---|---|
| `ANTHROPIC_API_KEY` | — | If not using OAuth file |
| `GITHUB_TOKEN` | — | Yes (or via secrets) |
| `REPO_URL` | — | Yes |
| `GIT_AUTHOR_NAME` | — | Yes |
| `GIT_AUTHOR_EMAIL` | — | Yes |
| `BASE_BRANCH` | `main` | No |
| `BRANCH_NAME` | `claude-agent/YYYYMMDD-HHmmss` | No |
| `TASK_TIMEOUT_MINUTES` | `60` | No |
| `CLAUDE_MAX_TURNS` | `50` | No |
| `OUTPUT_DIR` | `/output` | No |
| `CLAUDE_PLUGINS` | *(all 12 plugins space-separated)* | No |
| `AGENT_INSTRUCTIONS_FILE` | `/agent/instructions.md` | No |

### Volume Mounts

| Path | Purpose |
|---|---|
| `/task/task.md` | Task markdown file (optional, takes precedence over arg) |
| `/run/secrets/claude_credentials` | OAuth credentials (optional) |
| `/run/secrets/github_token` | GitHub token file (optional) |
| `/root/.claude.json` | OAuth credentials alternative |
| `/agent/instructions.md` | Container-level instructions (optional) |
| `/output` | Report output directory |
| `/root/.claude-agent-state/` *(named volume)* | Plugin cache — persists marker file across runs so plugins install once per host, not once per container. Separate from `/root/.claude/` to avoid shadowing baked-in `settings.json`. |

### Default Plugin List (`CLAUDE_PLUGINS`)

```
code-review@claude-plugins-official
feature-dev@claude-plugins-official
typescript-lsp@claude-plugins-official
code-simplifier@claude-plugins-official
security-guidance@claude-plugins-official
pr-review-toolkit@claude-plugins-official
superpowers@claude-plugins-official
csharp-lsp@claude-plugins-official
claude-md-management@claude-plugins-official
context7@claude-plugins-official
ralph-loop@claude-plugins-official
pyright-lsp@claude-plugins-official
```

---

## Project File Structure

```
claude-agent-container/
├── Dockerfile
├── entrypoint.sh               ← chmod +x, 6-phase orchestration
├── Makefile
├── docker-compose.yml
├── .dockerignore
├── README.md
├── CLAUDE.md
├── config/
│   └── settings.json           ← baked into image at /root/.claude/settings.json
└── .claude/
    └── settings.local.json     ← existing, for developing this project
```

---

## Implementation Plan

### Files to create (in this order)

1. **`config/settings.json`** — agent Claude Code settings:
   `autoUpdates: false`, full `permissions.allow` for all tools

2. **`entrypoint.sh`** — 6-phase bash script as designed above; `chmod +x`

3. **`Dockerfile`** — 8 layers as described; references `config/settings.json`
   and `entrypoint.sh` via COPY

4. **`Makefile`** — targets: `build`, `push`, `run`, `clean`
   - `build`: `docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .`
   - `push`: tag + push to `ghcr.io/adeotek/$(IMAGE_NAME)`
   - `run`: convenience target with all required `-e` and `-v` flags
   - `clean`: remove local image

5. **`docker-compose.yml`** — local dev/test helper with all volume mounts
   and env var pass-throughs; define a named volume `claude-agent-state` mounted
   at `/root/.claude-agent-state/` so the plugin marker file persists across runs
   on the same host (plugins install once, not on every `docker run`)

6. **`.dockerignore`** — exclude: `.git`, `.claude/`, `output/`, `Makefile`,
   `docker-compose.yml`, `*.md` (keep nothing unnecessary in build context)

7. **`CLAUDE.md`** — project docs: build commands, architecture summary,
   key design decisions, env var reference

8. **`README.md`** — usage docs with concrete `docker run` examples for both
   auth methods, task file usage, and instructions file usage

### Files to modify

- **`.claude/settings.local.json`** — add permissions for `docker build`,
  `make`, and bash commands needed to develop/test this project

---

## Verification

### Level 1 — Image sanity (no credentials needed)
```bash
docker run --rm claude-agent which git gh node dotnet python3 uv claude
docker run --rm claude-agent bash -c "node --version && dotnet --version && python3 --version && gh --version && claude --version"
docker run --rm claude-agent cat /root/.claude/settings.json
# expect: autoUpdates: false
```

### Level 2 — Entrypoint guard rails
```bash
# Missing task → exit 1
docker run --rm -e GITHUB_TOKEN=x -e REPO_URL=x -e GIT_AUTHOR_NAME=x -e GIT_AUTHOR_EMAIL=x claude-agent
# expect: "No task provided", exit 1

# Missing credentials → exit 1
docker run --rm -e REPO_URL=x -e GIT_AUTHOR_NAME=x -e GIT_AUTHOR_EMAIL=x claude-agent "task"
# expect: "No credentials found", exit 1

# File takes precedence over arg
docker run --rm -v ./test-task.md:/task/task.md:ro ... claude-agent "this arg is ignored"
# expect: task content read from file

# Timeout fires correctly
docker run --rm -e TASK_TIMEOUT_MINUTES=1 ... claude-agent "run a very long task"
# expect: exit 2, report status=timeout
```

### Level 3 — Full end-to-end
```bash
docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/yourorg/test-repo" \
  -v ./output:/output \
  claude-agent "Add a CONTRIBUTING.md with basic guidelines"
```

Success criteria:
- Exit code `0`
- `/output/report.md` exists with `status=success` and a changes summary
- Branch `claude-agent/YYYYMMDD-HHmmss` visible on GitHub
- PR created with correct commit author
- Commit diff appears in report under `## Changes`
