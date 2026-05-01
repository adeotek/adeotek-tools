# claude-agent-container

A sandboxed Docker container that runs [Claude Code CLI](https://claude.ai/code) as a
fully autonomous agent. Give it a task and a GitHub repo — it clones the repo, creates
a branch, does the work, and opens a Pull Request.

## Quick start

### Build

```bash
make build
```

### Run with Pro/Max account (OAuth)

Authenticate once on your host machine:

```bash
claude /login
```

Then run a task:

```bash
docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./output:/output \
  claude-agent "Add a CONTRIBUTING.md with basic contribution guidelines"
```

### Run with API key

```bash
docker run --rm \
  -e ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./output:/output \
  claude-agent "Fix the null pointer exception in UserService"
```

### Run with a task file

```bash
docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./my-task.md:/task/task.md:ro \
  -v ./output:/output \
  claude-agent
```

### Run with org-wide instructions

Create an `instructions.md` with your coding standards or policies:

```bash
docker run --rm \
  -v ~/.claude.json:/root/.claude.json:ro \
  -v ./org-standards.md:/agent/instructions.md:ro \
  -e GITHUB_TOKEN="$GITHUB_TOKEN" \
  -e GIT_AUTHOR_NAME="Claude Agent" \
  -e GIT_AUTHOR_EMAIL="claude-agent@noreply" \
  -e REPO_URL="https://github.com/your-org/your-repo" \
  -v ./output:/output \
  claude-agent "Add unit tests for the payment module"
```

## Output

After each run, `/output/report.md` is written:

```markdown
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
```

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
| uv | 0.11.8 |
| GitHub CLI | latest |
| TypeScript LSP | `typescript-language-server` |
| Python LSP | `pyright` |
| C# LSP | `csharp-ls` 0.20.0 |
| Angular CLI | `@angular/cli` |
