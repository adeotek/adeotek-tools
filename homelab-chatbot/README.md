# homelab-chatbot

Self-hosted RAG chatbot for a home-lab knowledge base. Ingests Markdown from two private GitHub repositories plus a multi-sheet Excel spreadsheet, then answers natural-language questions via **Ollama**, **Claude**, or **Gemini** (user-selectable per conversation).

- **Backend:** Python 3.13 + FastAPI + LlamaIndex + LanceDB + SQLite
- **Frontend:** Next.js 15 (static export) + Vercel AI SDK streaming UI
- **Deploy:** single Docker container on a LAN host

## Quick start

1. Copy configs and fill in secrets:
   ```bash
   cp config/config.example.yaml config/config.yaml
   cp .env.example .env
   # edit both
   ```
2. Build and run:
   ```bash
   make docker-build TAG=latest
   docker compose up -d
   ```
3. Browse to `http://<host>:8000`, log in, pick a provider + model, ask a question.

## Local development

**First-time setup:**
```bash
# Use the dev-specific config (local paths instead of /data/...)
cp config/config.dev.example.yaml config/config.dev.yaml
cp .env.example .env
# Edit .env: set HLCB_AUTH_PASSWORD_HASH, HLCB_SESSION_SECRET, API keys

mkdir -p data  # local data directory (gitignored)
```

```bash
# Backend (terminal 1) — run from homelab-chatbot/
cd backend && uv sync
HLCB_CONFIG_PATH=../config/config.dev.yaml \
  uv run --env-file ../.env \
  uvicorn app.main:create_app --factory --reload --port 8000

# Frontend (terminal 2) — run from homelab-chatbot/
cd frontend && npm ci
npm run dev
```

or

```bash
# Backend (terminal 1) — run from homelab-chatbot/
cd backend && ./run.sh

# Frontend (terminal 2) — run from homelab-chatbot/
cd frontend && ./run.sh
```

- **Frontend dev:** open `http://localhost:3000` — Next.js proxies `/api/*` to FastAPI on port 8000.
- **API only / no frontend:** open `http://localhost:8000` — FastAPI serves the last built static frontend (run `make build-frontend` first).

## Configuration

- `config/config.yaml` — non-secret knobs (providers, repos, retrieval settings). See `config.example.yaml`.
- `config/config.dev.yaml` — local development config (paths relative to backend/). See `config.dev.example.yaml`.
- `.env` — API keys, session secret, bcrypt-hashed user passwords. Never commit.

Generate a session secret:
```bash
python -c "import secrets; print(secrets.token_urlsafe(48))"
```
Hash a password:
```bash
python -c "import bcrypt; print(bcrypt.hashpw(b'pw', bcrypt.gensalt(12)).decode())"
```

## Providers

| Provider   | Source          | Tool-calling |
|------------|-----------------|--------------|
| Anthropic  | Cloud API       | yes          |
| Google     | Cloud API       | yes          |
| Ollama     | Local (host)    | model-dependent — see `config.yaml:providers.ollama.tool_capable` |

Non-tool-capable Ollama models degrade gracefully: the UI shows a banner and the agent skips retrieval. Keep the whitelist in sync with the models you have pulled.

## Ingestion

- Git repos are polled every `sync.interval_seconds` seconds. Change detection uses commit SHA + file-path diff.
- Excel sheets are loaded into an embedded SQLite (`kb.db`) on first run and on file-mtime change. All queries run read-only.
- Embeddings: `BAAI/bge-small-en-v1.5` on CPU (384-dim). Chunks carry `{repo, file_path, line_start, line_end, commit_sha, heading_path}` for future citation support.

## Testing

```bash
make test              # unit tests (fast, offline)
make test-integration  # opt-in; requires RUN_INTEGRATION_TESTS=1
```

Retrieval golden-set lives in `backend/tests/retrieval/` and asserts the expected chunk wins for ~12 curated queries.

## Deployment

Docker image is multi-stage (Node builder then slim Python runtime) and runs as `uid 1000`. Expose only on a trusted LAN; auth is intentionally minimal (session cookie + bcrypt).

## Spec & plan

- Design: [`docs/superpowers/specs/2026-04-16-homelab-chatbot-design.md`](../docs/superpowers/specs/2026-04-16-homelab-chatbot-design.md)
- Plan:   [`docs/superpowers/plans/2026-04-16-homelab-chatbot.md`](../docs/superpowers/plans/2026-04-16-homelab-chatbot.md)
