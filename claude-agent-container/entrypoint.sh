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

# Safe defaults so the trap function has valid variables even on early exit
CLAUDE_EXIT=0
STATUS="failure"
FINAL_EXIT=1
START_TIME=0
ELAPSED_FORMATTED="0m 0s"
TASK=""

# ─── Phase 8 function (registered via trap, always runs on EXIT) ─────────────
_run_phase8() {
  trap - EXIT
  echo "[claude-agent] Phase 8: Committing work, pushing branch, writing report..."

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
    PR_FLAGS=()
    if [ "$STATUS" != "success" ]; then
      TASK_SHORT="[Draft] ${TASK_SHORT}"
      PR_FLAGS=("--draft")
    fi
    PR_URL="$(gh pr create \
      --title "$TASK_SHORT" \
      --body "Automated task by Claude Agent.

**Status:** $STATUS
**Branch:** \`$BRANCH_NAME\`
**Elapsed:** $ELAPSED_FORMATTED" \
      ${PR_FLAGS[@]+"${PR_FLAGS[@]}"} 2>/dev/null)" \
      || PR_URL="none"
  fi

  # Write report
  TASK_SUMMARY="$(echo "$TASK" | head -c 200)"
  cat > "${OUTPUT_DIR}/report.md" <<__CLAUDE_AGENT_REPORT_END__
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
__CLAUDE_AGENT_REPORT_END__

  echo "[claude-agent] Report written to ${OUTPUT_DIR}/report.md"
  echo "[claude-agent] Done. Status: ${STATUS} | Exit: ${FINAL_EXIT}"
  exit "$FINAL_EXIT"
}

# ─── Guard Rails (Phase 1–4): all checks before any network calls ─────────────
# These run BEFORE registering the Phase 8 trap so that validation failures
# exit cleanly (exit 1) without triggering the commit/push/report logic.

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

# ─── Phase 3: Task resolution ────────────────────────────────────────────────
echo "[claude-agent] Phase 3: Resolving task..."

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

# ─── Phase 4: GitHub authentication ─────────────────────────────────────────
# All guard-rail checks above have passed. Now authenticate with GitHub and
# register the Phase 8 trap so the cleanup/report logic covers the remaining work.
echo "[claude-agent] Phase 4: Authenticating with GitHub..."

if ! echo "$GITHUB_TOKEN_VALUE" | gh auth login --with-token 2>&1; then
  echo "[claude-agent] ERROR: GitHub authentication failed. Check your token." >&2
  exit 1
fi
echo "[claude-agent] GitHub authentication successful."

# Register Phase 8 cleanup trap now that all guard-rail checks have passed and
# GitHub auth succeeded. The trap will handle commit/push/report on any exit.
trap _run_phase8 EXIT

# ─── Phase 5: Plugin installation (first run only) ───────────────────────────
MARKER_FILE="/root/.claude-agent-state/.plugins-installed"

if [ ! -f "$MARKER_FILE" ]; then
  echo "[claude-agent] Phase 5: Installing Claude Code plugins..."
  for plugin in "${PLUGINS[@]}"; do
    echo "[claude-agent] Installing plugin: $plugin"
    claude plugins install "$plugin" \
      || echo "[claude-agent] WARNING: Failed to install $plugin, continuing..."
  done
  touch "$MARKER_FILE"
  echo "[claude-agent] Plugins installed successfully."
else
  echo "[claude-agent] Phase 5: Plugins already installed (marker found), skipping."
fi

# ─── Phase 6: Repository setup ───────────────────────────────────────────────
echo "[claude-agent] Phase 6: Setting up repository..."

git clone "$REPO_URL" /workspace/repo
cd /workspace/repo
git checkout "$BASE_BRANCH" \
  || { echo "[claude-agent] ERROR: BASE_BRANCH '${BASE_BRANCH}' not found in repo." >&2; exit 1; }
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

# ─── Phase 7: Run Claude Code ────────────────────────────────────────────────
echo "[claude-agent] Phase 7: Running Claude Code agent (max-turns=$CLAUDE_MAX_TURNS, timeout=${TASK_TIMEOUT_MINUTES}m)..."
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

# ─── Phase 8: Report + push ──────────────────────────────────────────────────
# Handled by _run_phase8() via trap EXIT — always runs even on unexpected error.
