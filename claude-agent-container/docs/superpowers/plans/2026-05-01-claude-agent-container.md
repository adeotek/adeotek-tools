# Claude Agent Container Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a sandboxed Docker container that runs Claude Code CLI as a fully autonomous agent — receiving a task, cloning a GitHub repo, doing the work, and opening a PR.

**Architecture:** Single-use container (one `docker run` = one task). A bash `entrypoint.sh` orchestrates six phases: credential resolution, git identity, plugin installation (first run only), repo setup, Claude Code execution with timeout, and report + PR creation. The image is built on `debian:bookworm-slim` with Node.js 24, .NET SDK 9, Python 3 + uv, GitHub CLI, language servers, and the native Claude Code binary.

**Tech Stack:** Docker (debian:bookworm-slim), Bash, Claude Code CLI (native), Node.js 24, .NET SDK 9, Python 3 + uv, GitHub CLI, TypeScript/pyright/csharp-ls language servers.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `config/settings.json` | Create | Claude Code agent settings baked into image |
| `entrypoint.sh` | Create | 6-phase orchestration script |
| `Dockerfile` | Create | Image definition — 8 ordered layers |
| `Makefile` | Create | build / push / run / clean targets |
| `docker-compose.yml` | Create | Local dev/test helper with named volumes |
| `.dockerignore` | Create | Exclude non-essential files from build context |
| `CLAUDE.md` | Create | Project docs for developing this container |
| `README.md` | Create | Usage docs with concrete examples |
| `.claude/settings.local.json` | Modify | Add docker/make permissions for local dev |

---

## Task 1: Claude Code agent settings

**Files:**
- Create: `config/settings.json`

- [ ] **Step 1: Create config directory and settings file**

```bash
mkdir -p config
```

Create `config/settings.json`:

```json
{
  "autoUpdates": false,
  "permissions": {
    "allow": [
      "Bash(*)",
      "Read(*)",
      "Write(*)",
      "Edit(*)",
      "WebFetch(*)",
      "WebSearch(*)",
      "Agent(*)"
    ],
    "deny": []
  }
}
```

- [ ] **Step 2: Verify file is valid JSON**

```bash
python3 -c "import json; json.load(open('config/settings.json')); print('valid')"
```

Expected: `valid`

- [ ] **Step 3: Commit**

```bash
git add config/settings.json
git commit -m "feat(claude-agent-container): add Claude Code agent settings"
```

---

## Task 2: entrypoint.sh — six-phase orchestration script

**Files:**
- Create: `entrypoint.sh`

This is the heart of the container. Read the entire script carefully before modifying any phase.

- [ ] **Step 1: Create entrypoint.sh**

Create `entrypoint.sh` with the following content:

```bash
#!/usr/bin/env bash
set -euo pipefail

# ─── Defaults ────────────────────────────────────────────────────────────────
TASK_TIMEOUT_MINUTES="${TASK_TIMEOUT_MINUTES:-60}"
CLAUDE_MAX_TURNS="${CLAUDE_MAX_TURNS:-50}"
BASE_BRANCH="${BASE_BRANCH:-main}"
OUTPUT_DIR="${OUTPUT_DIR:-/output}"
AGENT_INSTRUCTIONS_FILE="${AGENT_INSTRUCTIONS_FILE:-/agent/instructions.md}"
BRANCH_NAME="${BRANCH_NAME:-claude-agent/$(date +%Y%m%d-%H%M%S)}"

DEFAULT_PLUGINS=(
  "code-review@claude-plugins-official"
  "feature-dev@claude-plugins-official"
  "typescript-lsp@claude-plugins-official"
  "code-simplifier@claude-plugins-official"
  "security-guidance@claude-plugins-official"
  "pr-review-toolkit@claude-plugins-official"
  "superpowers@claude-plugins-official"
  "csharp-lsp@claude-plugins-official"
  "claude-md-management@claude-plugins-official"
  "context7@claude-plugins-official"
  "ralph-loop@claude-plugins-official"
  "pyright-lsp@claude-plugins-official"
)
# CLAUDE_PLUGINS env var overrides default list (space-separated)
IFS=' ' read -ra PLUGINS <<< "${CLAUDE_PLUGINS:-${DEFAULT_PLUGINS[*]}}"

mkdir -p "$OUTPUT_DIR"
mkdir -p /root/.claude-agent-state

# ─── Phase 1: Credentials ────────────────────────────────────────────────────
echo "[claude-agent] Phase 1: Resolving credentials..."

if [ -f /run/secrets/claude_credentials ]; then
  echo "[claude-agent] Using Claude credentials from /run/secrets/claude_credentials"
  cp /run/secrets/claude_credentials /root/.claude.json
elif [ -f /root/.claude.json ]; then
  echo "[claude-agent] Using mounted /root/.claude.json"
elif [ -n "${ANTHROPIC_API_KEY:-}" ]; then
  echo "[claude-agent] Using ANTHROPIC_API_KEY"
else
  echo "[claude-agent] ERROR: No Claude credentials found." >&2
  echo "[claude-agent] Provide one of: /run/secrets/claude_credentials, /root/.claude.json mount, or ANTHROPIC_API_KEY." >&2
  exit 1
fi

GITHUB_TOKEN_VALUE=""
if [ -f /run/secrets/github_token ]; then
  GITHUB_TOKEN_VALUE="$(cat /run/secrets/github_token)"
elif [ -n "${GITHUB_TOKEN:-}" ]; then
  GITHUB_TOKEN_VALUE="$GITHUB_TOKEN"
else
  echo "[claude-agent] ERROR: No GitHub token found." >&2
  echo "[claude-agent] Provide GITHUB_TOKEN env var or /run/secrets/github_token." >&2
  exit 1
fi

echo "$GITHUB_TOKEN_VALUE" | gh auth login --with-token
echo "[claude-agent] GitHub authentication successful."

# ─── Phase 2: Git identity ───────────────────────────────────────────────────
echo "[claude-agent] Phase 2: Configuring git identity..."

if [ -z "${GIT_AUTHOR_NAME:-}" ]; then
  echo "[claude-agent] ERROR: GIT_AUTHOR_NAME is required." >&2
  exit 1
fi
if [ -z "${GIT_AUTHOR_EMAIL:-}" ]; then
  echo "[claude-agent] ERROR: GIT_AUTHOR_EMAIL is required." >&2
  exit 1
fi

git config --global user.name "$GIT_AUTHOR_NAME"
git config --global user.email "$GIT_AUTHOR_EMAIL"
git config --global push.autoSetupRemote true

# ─── Phase 3: Plugin installation (first run only) ───────────────────────────
MARKER_FILE="/root/.claude-agent-state/.plugins-installed"

if [ ! -f "$MARKER_FILE" ]; then
  echo "[claude-agent] Phase 3: Installing Claude Code plugins..."
  for plugin in "${PLUGINS[@]}"; do
    echo "[claude-agent] Installing plugin: $plugin"
    claude plugins install "$plugin" \
      || echo "[claude-agent] WARNING: Failed to install $plugin, continuing..."
  done
  touch "$MARKER_FILE"
  echo "[claude-agent] Plugins installed successfully."
else
  echo "[claude-agent] Phase 3: Plugins already installed (marker found), skipping."
fi

# ─── Phase 4: Task + repo setup ──────────────────────────────────────────────
echo "[claude-agent] Phase 4: Resolving task and setting up repository..."

TASK=""
if [ -f /task/task.md ]; then
  TASK="$(cat /task/task.md)"
  echo "[claude-agent] Task loaded from /task/task.md"
elif [ -n "${1:-}" ]; then
  TASK="$1"
  echo "[claude-agent] Task loaded from argument"
else
  echo "[claude-agent] ERROR: No task provided." >&2
  echo "[claude-agent] Pass a task as a docker run argument or mount a file at /task/task.md." >&2
  exit 1
fi

if [ -z "${REPO_URL:-}" ]; then
  echo "[claude-agent] ERROR: REPO_URL is required." >&2
  exit 1
fi

git clone "$REPO_URL" /workspace/repo
cd /workspace/repo
git checkout "$BASE_BRANCH" 2>/dev/null || true
git checkout -b "$BRANCH_NAME"
echo "[claude-agent] Repository cloned. Working branch: $BRANCH_NAME"

FULL_PROMPT="$TASK"
if [ -f "$AGENT_INSTRUCTIONS_FILE" ]; then
  INSTRUCTIONS="$(cat "$AGENT_INSTRUCTIONS_FILE")"
  FULL_PROMPT="${INSTRUCTIONS}

---

${TASK}"
  echo "[claude-agent] Container-level instructions prepended from $AGENT_INSTRUCTIONS_FILE"
fi

# ─── Phase 5: Run Claude Code ────────────────────────────────────────────────
echo "[claude-agent] Phase 5: Running Claude Code agent (max-turns=$CLAUDE_MAX_TURNS, timeout=${TASK_TIMEOUT_MINUTES}m)..."
START_TIME=$(date +%s)

CLAUDE_EXIT=0
timeout "$((TASK_TIMEOUT_MINUTES * 60))" \
  claude \
    --dangerously-skip-permissions \
    --max-turns "$CLAUDE_MAX_TURNS" \
    -p "$FULL_PROMPT" \
  || CLAUDE_EXIT=$?

END_TIME=$(date +%s)
ELAPSED=$(( END_TIME - START_TIME ))
ELAPSED_FORMATTED="$(( ELAPSED / 60 ))m $(( ELAPSED % 60 ))s"

# ─── Phase 6: Report + push ──────────────────────────────────────────────────
echo "[claude-agent] Phase 6: Committing work, pushing branch, writing report..."

if [ "$CLAUDE_EXIT" -eq 0 ]; then
  STATUS="success"
  FINAL_EXIT=0
elif [ "$CLAUDE_EXIT" -eq 124 ]; then
  STATUS="timeout"
  FINAL_EXIT=2
else
  STATUS="failure"
  FINAL_EXIT=1
fi

# Commit any changes (best-effort)
git add -A
if ! git diff --staged --quiet; then
  git commit -m "chore: claude-agent task [status: $STATUS]"
fi

# Push branch (best-effort)
git push || echo "[claude-agent] WARNING: Failed to push branch."

# Collect git stats for report
COMMITS=""
DIFF_STAT=""
if git log "origin/${BASE_BRANCH}..HEAD" --oneline 2>/dev/null | grep -q .; then
  COMMITS="$(git log "origin/${BASE_BRANCH}..HEAD" --oneline)"
  DIFF_STAT="$(git diff "origin/${BASE_BRANCH}..HEAD" --stat 2>/dev/null || true)"
fi

# Open PR (best-effort)
PR_URL="none"
if [ -n "$COMMITS" ]; then
  TASK_SHORT="$(echo "$TASK" | head -c 72 | tr '\n' ' ')"
  PR_FLAGS=""
  if [ "$STATUS" != "success" ]; then
    TASK_SHORT="[Draft] ${TASK_SHORT}"
    PR_FLAGS="--draft"
  fi
  PR_URL="$(gh pr create \
    --title "$TASK_SHORT" \
    --body "Automated task by Claude Agent.

**Status:** $STATUS
**Branch:** \`$BRANCH_NAME\`
**Elapsed:** $ELAPSED_FORMATTED" \
    $PR_FLAGS 2>/dev/null)" \
    || PR_URL="none"
fi

# Write report
TASK_SUMMARY="$(echo "$TASK" | head -c 200)"
cat > "${OUTPUT_DIR}/report.md" <<REPORT
# Claude Agent Report

- **Status:** ${STATUS}
- **Task:** ${TASK_SUMMARY}
- **Repo:** ${REPO_URL}
- **Branch:** ${BRANCH_NAME}
- **PR:** ${PR_URL}
- **Elapsed:** ${ELAPSED_FORMATTED}

## Changes

${DIFF_STAT:-_No changes committed._}

### Commits

${COMMITS:-_None._}
REPORT

echo "[claude-agent] Report written to ${OUTPUT_DIR}/report.md"
echo "[claude-agent] Done. Status: ${STATUS} | Exit: ${FINAL_EXIT}"

exit "$FINAL_EXIT"
```

- [ ] **Step 2: Make executable**

```bash
chmod +x entrypoint.sh
```

- [ ] **Step 3: Verify syntax**

```bash
bash -n entrypoint.sh && echo "syntax ok"
```

Expected: `syntax ok`

- [ ] **Step 4: Commit**

```bash
git add entrypoint.sh
git commit -m "feat(claude-agent-container): add 6-phase entrypoint script"
```

---

## Task 3: Dockerfile

**Files:**
- Create: `Dockerfile`

- [ ] **Step 1: Create Dockerfile**

Create `Dockerfile`:

```dockerfile
FROM debian:bookworm-slim

LABEL description="Claude Code autonomous agent container"

ENV DEBIAN_FRONTEND=noninteractive

# ─── Layer 1: System packages + Python 3 ─────────────────────────────────────
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    git \
    ca-certificates \
    gnupg \
    build-essential \
    libssl-dev \
    unzip \
    jq \
    procps \
    python3 \
    python3-pip \
    python3-venv \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# ─── Layer 2: Node.js 24 LTS ─────────────────────────────────────────────────
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# ─── Layer 3: .NET SDK 9 ─────────────────────────────────────────────────────
ENV DOTNET_ROOT=/usr/local/dotnet
ENV PATH="${PATH}:${DOTNET_ROOT}:${DOTNET_ROOT}/tools"

RUN curl -fsSL https://dot.net/v1/dotnet-install.sh \
    | bash -s -- --channel 9.0 --install-dir "$DOTNET_ROOT" --no-path

# ─── Layer 4: Python uv ──────────────────────────────────────────────────────
ENV PATH="${PATH}:/root/.local/bin"

RUN curl -fsSL https://astral.sh/uv/install.sh | bash

# ─── Layer 5: GitHub CLI ─────────────────────────────────────────────────────
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      | tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

# ─── Layer 6: Language servers & dev tooling ─────────────────────────────────
RUN npm install -g \
    typescript \
    typescript-language-server \
    pyright \
    @angular/cli

RUN dotnet tool install --global csharp-ls

# ─── Layer 7: Claude Code (native binary) ────────────────────────────────────
RUN curl -fsSL https://claude.ai/install.sh | bash

# ─── Layer 8: Config + entrypoint ────────────────────────────────────────────
WORKDIR /workspace

COPY config/settings.json /root/.claude/settings.json
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
```

- [ ] **Step 2: Verify Dockerfile parses correctly**

```bash
docker build --check . 2>&1 | head -20
```

Expected: no fatal errors (warnings about ARG before FROM are acceptable)

- [ ] **Step 3: Build the image (this will take 10–20 minutes on first run)**

```bash
docker build -t claude-agent:latest .
```

Expected: `Successfully built <id>` and `Successfully tagged claude-agent:latest`

- [ ] **Step 4: Level 1 sanity checks — all tools present**

```bash
docker run --rm claude-agent:latest which git gh node dotnet python3 uv claude
```

Expected: one path per line, no "not found" errors.

```bash
docker run --rm claude-agent:latest bash -c "
  echo Node: \$(node --version) &&
  echo .NET: \$(dotnet --version) &&
  echo Python: \$(python3 --version) &&
  echo uv: \$(uv --version) &&
  echo gh: \$(gh --version | head -1) &&
  echo claude: \$(claude --version)
"
```

Expected: version strings for all six tools.

- [ ] **Step 5: Verify settings baked in correctly**

```bash
docker run --rm claude-agent:latest cat /root/.claude/settings.json
```

Expected: JSON with `"autoUpdates": false`.

- [ ] **Step 6: Commit**

```bash
git add Dockerfile
git commit -m "feat(claude-agent-container): add Dockerfile with 8-layer build"
```

---

## Task 4: Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Create Makefile**

Create `Makefile` (use tabs, not spaces, for recipe indentation):

```makefile
IMAGE_NAME ?= claude-agent
IMAGE_TAG  ?= latest
REGISTRY   ?= ghcr.io/adeotek

.PHONY: build push run clean

build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

push: build
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

run:
	@[ -n "$(REPO_URL)" ] || (echo "Usage: make run REPO_URL=https://github.com/org/repo TASK=\"your task\""; exit 1)
	docker run --rm \
	  -v ~/.claude.json:/root/.claude.json:ro \
	  -e GITHUB_TOKEN \
	  -e GIT_AUTHOR_NAME \
	  -e GIT_AUTHOR_EMAIL \
	  -e REPO_URL="$(REPO_URL)" \
	  -v "$(PWD)/output:/output" \
	  $(IMAGE_NAME):$(IMAGE_TAG) "$(TASK)"

clean:
	docker rmi $(IMAGE_NAME):$(IMAGE_TAG) 2>/dev/null || true
```

- [ ] **Step 2: Verify make help works**

```bash
make build --dry-run
```

Expected: prints `docker build -t claude-agent:latest .` without executing.

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "feat(claude-agent-container): add Makefile with build/push/run/clean targets"
```

---

## Task 5: docker-compose.yml

**Files:**
- Create: `docker-compose.yml`

- [ ] **Step 1: Create docker-compose.yml**

Create `docker-compose.yml`:

```yaml
services:
  claude-agent:
    build: .
    image: claude-agent:latest
    volumes:
      - ~/.claude.json:/root/.claude.json:ro
      - ./task.md:/task/task.md:ro
      - ./output:/output
      - claude-agent-state:/root/.claude-agent-state
    environment:
      - GITHUB_TOKEN
      - GIT_AUTHOR_NAME
      - GIT_AUTHOR_EMAIL
      - REPO_URL
      - BASE_BRANCH=main
      - TASK_TIMEOUT_MINUTES=60
      - CLAUDE_MAX_TURNS=50
    # command: "your task here"   # alternative to task.md mount

volumes:
  claude-agent-state:
    # Persists the plugin installation marker across container runs.
    # Plugins install once per host rather than on every docker run.
```

- [ ] **Step 2: Verify compose config is valid**

```bash
docker compose config --quiet && echo "valid"
```

Expected: `valid`

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml
git commit -m "feat(claude-agent-container): add docker-compose with named plugin cache volume"
```

---

## Task 6: .dockerignore

**Files:**
- Create: `.dockerignore`

- [ ] **Step 1: Create .dockerignore**

Create `.dockerignore`:

```
.git
.claude/
output/
docker-compose.yml
Makefile
docs/
*.md
```

- [ ] **Step 2: Verify build context is smaller**

```bash
docker build --no-cache -t claude-agent:latest . 2>&1 | grep "Sending build context"
```

The build context should be small (kilobytes, not megabytes).

- [ ] **Step 3: Commit**

```bash
git add .dockerignore
git commit -m "feat(claude-agent-container): add .dockerignore to minimize build context"
```

---

## Task 7: CLAUDE.md

**Files:**
- Create: `CLAUDE.md`

- [ ] **Step 1: Create CLAUDE.md**

Create `CLAUDE.md`:

```markdown
# CLAUDE.md

## Project

`claude-agent-container` — a sandboxed Docker container that runs Claude Code CLI as a
fully autonomous agent. Receives a task, clones a GitHub repo, creates a branch, does
the work, and opens a Pull Request.

Part of the `adeotek-tools` monorepo. This project is a Docker image only — no runtime
dependencies on other projects in the monorepo.

## Build & run commands

\`\`\`bash
make build                                       # build image locally
make push                                        # tag + push to ghcr.io/adeotek/claude-agent
make run REPO_URL=https://github.com/org/repo \
         TASK="add a CONTRIBUTING.md"            # run a task (uses ~/.claude.json for auth)
make clean                                       # remove local image
docker compose up                                # alternative: run via compose with task.md mount
\`\`\`

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
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "feat(claude-agent-container): add CLAUDE.md project documentation"
```

---

## Task 8: README.md

**Files:**
- Modify: `README.md` (currently empty)

- [ ] **Step 1: Write README.md**

Replace the contents of `README.md`:

```markdown
# claude-agent-container

A sandboxed Docker container that runs [Claude Code CLI](https://claude.ai/code) as a
fully autonomous agent. Give it a task and a GitHub repo — it clones the repo, creates
a branch, does the work, and opens a Pull Request.

## Quick start

### Build

\`\`\`bash
make build
\`\`\`

### Run with Pro/Max account (OAuth)

Authenticate once on your host machine:

\`\`\`bash
claude /login
\`\`\`

Then run a task:

\`\`\`bash
docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./output:/output \
  claude-agent "Add a CONTRIBUTING.md with basic contribution guidelines"
\`\`\`

### Run with API key

\`\`\`bash
docker run --rm \
  -e ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./output:/output \
  claude-agent "Fix the null pointer exception in UserService"
\`\`\`

### Run with a task file

\`\`\`bash
docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./my-task.md:/task/task.md:ro \
  -v ./output:/output \
  claude-agent
\`\`\`

### Run with org-wide instructions

Create an `instructions.md` with your coding standards or policies:

\`\`\`bash
docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -v ./org-standards.md:/agent/instructions.md:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./output:/output \
  claude-agent "Add unit tests for the payment module"
\`\`\`

## Output

After each run, `/output/report.md` is written:

\`\`\`markdown
# Claude Agent Report

- **Status:** success
- **Task:** Add a CONTRIBUTING.md with basic contribution guidelines
- **Repo:** https://github.com/your-org/your-repo
- **Branch:** claude-agent/20260501-143022
- **PR:** https://github.com/your-org/your-repo/pull/42
- **Elapsed:** 3m 12s

## Changes

1 file changed, 46 insertions(+)
 CONTRIBUTING.md | 46 ++++++++++++++++++++++++++++++++++++++++++++++

### Commits
- a3f1c2e Add CONTRIBUTING.md with contribution guidelines
\`\`\`

Exit codes: `0` = success, `1` = failure, `2` = timeout.

## Environment variables

| Variable | Default | Required |
|---|---|---|
| `ANTHROPIC_API_KEY` | — | If not using OAuth file |
| `GITHUB_TOKEN` | — | Yes (or `/run/secrets/github_token`) |
| `REPO_URL` | — | Yes |
| `GIT_AUTHOR_NAME` | — | Yes |
| `GIT_AUTHOR_EMAIL` | — | Yes |
| `BASE_BRANCH` | `main` | No |
| `BRANCH_NAME` | `claude-agent/YYYYMMDD-HHmmss` | No |
| `TASK_TIMEOUT_MINUTES` | `60` | No |
| `CLAUDE_MAX_TURNS` | `50` | No |
| `OUTPUT_DIR` | `/output` | No |
| `CLAUDE_PLUGINS` | *(all 12 plugins)* | No |
| `AGENT_INSTRUCTIONS_FILE` | `/agent/instructions.md` | No |

## Volume mounts

| Path | Purpose |
|---|---|
| `/task/task.md` | Task file (takes precedence over CLI arg) |
| `/root/.claude.json` | Claude Pro/Max OAuth credentials |
| `/run/secrets/claude_credentials` | Claude OAuth credentials (Docker secrets) |
| `/run/secrets/github_token` | GitHub token (Docker secrets) |
| `/agent/instructions.md` | Container-level instructions prepended to every task |
| `/output` | Report output directory |

## Included tools

| Tool | Version |
|---|---|
| Claude Code CLI | native (latest at build time) |
| Node.js | 24 LTS |
| .NET SDK | 9 |
| Python | 3 (system) |
| uv | latest |
| GitHub CLI | latest |
| TypeScript LSP | `typescript-language-server` |
| Python LSP | `pyright` |
| C# LSP | `csharp-ls` |
| Angular CLI | `@angular/cli` |
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "feat(claude-agent-container): add README with usage examples"
```

---

## Task 9: Update .claude/settings.local.json

**Files:**
- Modify: `.claude/settings.local.json`

- [ ] **Step 1: Read current settings**

```bash
cat .claude/settings.local.json
```

- [ ] **Step 2: Add docker and make permissions**

Update `.claude/settings.local.json` to add permissions for building and testing the container locally. Merge these entries into the existing `permissions.allow` array:

```json
{
  "permissions": {
    "allow": [
      "Bash(grep -v \"^\\.$\")",
      "Bash(git -C /home/georg/projects/adeotek-tools log --all --format=\"%H %s\")",
      "Bash(docker build*)",
      "Bash(docker run*)",
      "Bash(docker rmi*)",
      "Bash(docker compose*)",
      "Bash(make*)",
      "Bash(bash -n entrypoint.sh)",
      "Bash(python3 -c*)"
    ]
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add .claude/settings.local.json
git commit -m "chore(claude-agent-container): add docker/make permissions to local settings"
```

---

## Task 10: Level 2 verification — entrypoint guard rails

No files created — this task verifies failure paths work correctly.

- [ ] **Step 1: Verify missing task exits with code 1**

```bash
docker run --rm \
  -e GITHUB_TOKEN=dummy \
  -e REPO_URL=dummy \
  -e GIT_AUTHOR_NAME=dummy \
  -e GIT_AUTHOR_EMAIL=dummy \
  claude-agent:latest
echo "Exit code: $?"
```

Expected output contains: `ERROR: No task provided`
Expected exit code: `1`

- [ ] **Step 2: Verify missing credentials exits with code 1**

```bash
docker run --rm \
  -e GITHUB_TOKEN=dummy \
  -e REPO_URL=dummy \
  -e GIT_AUTHOR_NAME=dummy \
  -e GIT_AUTHOR_EMAIL=dummy \
  claude-agent:latest "some task"
echo "Exit code: $?"
```

Expected output contains: `ERROR: No Claude credentials found`
Expected exit code: `1`

- [ ] **Step 3: Verify missing GitHub token exits with code 1**

```bash
docker run --rm \
  -e ANTHROPIC_API_KEY=dummy \
  -e REPO_URL=dummy \
  -e GIT_AUTHOR_NAME=dummy \
  -e GIT_AUTHOR_EMAIL=dummy \
  claude-agent:latest "some task"
echo "Exit code: $?"
```

Expected output contains: `ERROR: No GitHub token found`
Expected exit code: `1`

- [ ] **Step 4: Verify task file takes precedence over arg**

```bash
echo "Task from FILE" > /tmp/test-task.md
docker run --rm \
  -v /tmp/test-task.md:/task/task.md:ro \
  -e ANTHROPIC_API_KEY=dummy \
  -e GITHUB_TOKEN=dummy \
  -e REPO_URL=dummy \
  -e GIT_AUTHOR_NAME=dummy \
  -e GIT_AUTHOR_EMAIL=dummy \
  claude-agent:latest "Task from ARG" 2>&1 | head -20
```

Expected: output contains `Task loaded from /task/task.md` (not "from argument")

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "test(claude-agent-container): verify Level 1 and Level 2 entrypoint guard rails"
```

---

## Task 11: Save design spec

**Files:**
- Create: `docs/superpowers/specs/2026-05-01-claude-agent-container-design.md`

The design spec was written to the plan file during brainstorming. Copy it to the canonical location inside the project.

- [ ] **Step 1: Create specs directory and copy spec**

```bash
mkdir -p docs/superpowers/specs
cp /home/georg/.claude/plans/the-goal-of-this-snuggly-dahl.md \
   docs/superpowers/specs/2026-05-01-claude-agent-container-design.md
```

- [ ] **Step 2: Commit**

```bash
git add docs/
git commit -m "docs(claude-agent-container): add design spec and implementation plan"
```

---

## Level 3 end-to-end test (manual, requires real credentials)

Run this after all tasks are complete to confirm the full flow works.

```bash
# Prerequisites:
# - claude /login already run on host (creates ~/.claude.json)
# - GITHUB_TOKEN set with repo read/write and PR permissions
# - A scratch GitHub repo you own (replace below)

mkdir -p output

docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/YOUR_ORG/YOUR_TEST_REPO" \
  -v "$(pwd)/output:/output" \
  claude-agent:latest "Add a CONTRIBUTING.md with basic contribution guidelines"

echo "Exit code: $?"
cat output/report.md
```

Success criteria:
- Exit code `0`
- `output/report.md` contains `Status: success` and a non-empty `## Changes` section
- A new branch `claude-agent/YYYYMMDD-HHmmss` is visible on GitHub
- A PR is open with the correct commit author
