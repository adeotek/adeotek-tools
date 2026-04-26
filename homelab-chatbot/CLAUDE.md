# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository description

Self-hosted RAG chatbot for a home-lab knowledge base. Ingests Markdown from GitHub repos and Excel spreadsheets, then answers questions via Anthropic, Google, or Ollama (user-selectable per conversation).

- **Backend:** Python 3.13 + FastAPI + LlamaIndex + LanceDB + SQLite
- **Frontend:** Next.js 15 (static export) + Vercel AI SDK streaming UI
- **Deploy:** single Docker container (multi-stage build), runs as uid 10001

## Build / dev commands

```bash
# Install backend deps (run from homelab-chatbot/)
cd backend && uv sync

# Install frontend deps
cd frontend && npm ci

# Start backend (dev, with hot reload)
cd backend && ./run.sh          # or: HLCB_CONFIG_PATH=../config/config.dev.yaml uv run --env-file ../.env uvicorn app.main:create_app --factory --reload --port 8000

# Start frontend (dev, proxies /api/* to :8000)
cd frontend && ./run.sh         # or: npm run dev
```

### Tests

```bash
make test                       # backend (pytest) + frontend (vitest)
cd backend && uv run pytest -q  # backend unit tests only
cd backend && uv run pytest tests/unit/test_chat_route.py  # single test file
cd backend && RUN_INTEGRATION_TESTS=1 uv run pytest -v tests/integration  # integration tests
cd frontend && npm run test -- --run  # frontend tests once
```

### Lint / format

```bash
make lint                       # ruff check (backend) + eslint (frontend)
make fmt                        # ruff format (backend) + prettier (frontend)
cd frontend && npm run typecheck  # TypeScript type checking
```

### Docker

```bash
make docker-build TAG=latest    # builds ghcr.io/adeotek/homelab-chatbot:latest
make run                        # docker compose up --build (foreground)
```

## Configuration

- `config/config.yaml` — non-secret knobs (providers, repos, retrieval settings). Copy from `config.example.yaml`.
- `config/config.dev.yaml` — local paths instead of `/data/...`. Copy from `config.dev.example.yaml`.
- `.env` — API keys, bcrypt-hashed password, session secret. Never commit. Copy from `.env.example`.

Config path is controlled by `HLCB_CONFIG_PATH` env var (defaults to `config.yaml`).

## Architecture

### Request / response flow

```
POST /api/chat
  → chat.py route
  → build_llm() selects Anthropic / Google / Ollama LLM via LlamaIndex
  → supports_tools() checks provider + config.llm.ollama.tool_capable_models
  → run_agent() streams AgentEvents:
      ├─ tools enabled  → LlamaIndex ReActAgent with VectorSearchTool + SQLTool
      └─ tools disabled → direct astream_chat (no retrieval, shows banner in UI)
  → SSE stream (text-delta / tool-result / done) → StreamingResponse
  → frontend lib/stream.ts consumes events
```

### Dual-database retrieval

The agent has two tools:

| Tool | Storage | Use case |
|------|---------|----------|
| `VectorSearchTool` | LanceDB (`/data/lance`) | Semantic search over Markdown chunks |
| `SQLTool` | SQLite `kb.db` | Structured queries over Excel inventory sheets |

Both tools are registered as LlamaIndex `FunctionTool`s and presented to the ReActAgent. Non-tool-capable Ollama models only get `VectorSearchTool` (no SQL), and the UI shows a banner.

### Ingestion pipeline

`SyncScheduler` (APScheduler) → `IngestionOrchestrator.run_once()` → per repo:
1. `GitSync.sync()` — clones/pulls repo, returns `ChangedFile` list
2. For `.md` files: `chunk_markdown_file()` → `Embedder.embed_batch()` → `VectorStore.upsert()`
3. For `.xlsx` files: `ExcelLoader.load()` → SQLite `kb.db`

Change detection uses commit SHA diff; deleted files are removed from the vector store.

### App state

`create_app()` in `app/main.py` is the FastAPI application factory. All singletons (config, secrets, chat_db, vector_store, vector_tool, sql_tool, orchestrator) are attached to `app.state` and accessed in routes via `request.app.state.*`. The `Secrets` class reads from env vars at startup.

### Frontend

Next.js static export (`output: 'export'`). In production the built `out/` is served by FastAPI's `StaticFiles` mount. In dev, Next.js dev server runs on `:3000` and proxies `/api/*` to FastAPI on `:8000` (configured in `next.config.ts`).

Chat streaming is handled in `lib/stream.ts` consuming the SSE event stream. The `components/chat/` directory owns the chat UI; `ProviderPicker` in settings controls per-conversation provider+model selection persisted via `POST /api/settings`.

### Backend code style

- Line length: 100 (ruff)
- Python 3.13, strict mypy
- Imports: stdlib → third-party, grouped with blank lines
- PascalCase for exported classes, camelCase for internal variables
- `pytest-asyncio` with `asyncio_mode = "auto"` — no `@pytest.mark.asyncio` needed
