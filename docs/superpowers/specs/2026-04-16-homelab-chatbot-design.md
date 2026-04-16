# homelab-chatbot — Design Document

**Date:** 2026-04-16
**Status:** Approved (pending final review)
**Location in monorepo:** `homelab-chatbot/`
**Version tag convention:** `homelab-chatbot/vMAJOR.MINOR.PATCH`

## 1. Purpose

A self-hosted chatbot for querying a personal home-lab knowledge base. It ingests content from:

- Two private GitHub repositories containing Markdown documentation
- An Excel spreadsheet (multiple sheets of small tabular data, kept inside one of those repositories)

Users ask natural-language questions through a web UI and receive answers grounded in the ingested content. The chatbot supports three LLM providers as true peers — **Ollama** (local), **Anthropic Claude** (API), and **Google Gemini** (API) — selectable per conversation.

## 2. Requirements

### Functional

- Ingest Markdown files from two configured GitHub repositories (private, token auth).
- Ingest a multi-sheet Excel spreadsheet; each sheet becomes a queryable SQLite table.
- Support natural-language questions against Markdown (semantic retrieval) and against Excel (aggregate-capable structured queries).
- Expose a chat web UI with conversation history sidebar (ChatGPT-style) and per-conversation provider/model selection.
- Persist conversations locally in SQLite.
- Poll the GitHub repositories every few minutes and reindex changed content automatically.
- Gate access with basic authentication (single shared password, LAN-only deployment).

### Non-functional

- **Deployment:** Single Docker container runnable via `docker-compose` on a Linux homelab host.
- **Privacy posture:** Data stays local except when the user explicitly selects a cloud LLM provider for a conversation.
- **Flexibility (top priority):** Swapping LLM providers, retrievers, chunking strategies, or embedding models must be low-friction.
- **Stability:** Rely on mature libraries (LlamaIndex, FastAPI, Next.js, Vercel AI SDK).
- **Performance:** Adequate for single-digit concurrent users; LLM inference dominates latency.

### Out of scope (v1)

- Citations / source attribution in responses (metadata will be preserved in the data model for future implementation).
- Real-time webhook-driven reindex (polling is sufficient for the stated update frequency).
- Internet-exposed deployment / OAuth / multi-tenant auth.
- Full-text search over chat history (basic list + filter-by-title only).
- E2E browser tests (component tests + integration tests cover the seams).

## 3. Technology Stack

| Layer | Choice | Rationale |
|-------|--------|-----------|
| Backend language | Python 3.13 | RAG ecosystem maturity; official Anthropic/Google/Ollama SDKs; excellent Excel tooling |
| Backend framework | FastAPI + uvicorn | Async, lightweight, well-suited to SSE streaming |
| RAG framework | LlamaIndex | Best-in-class provider/retriever abstractions; native LanceDB + NL-SQL integrations |
| Vector store | LanceDB (embedded) | Single-directory storage, zero ops, fast at this scale |
| Embedding model | `BAAI/bge-small-en-v1.5` via `sentence-transformers` | Deterministic, CPU-fine, offline, ~130 MB |
| Structured store | SQLite | One DB for chat history, a separate read-only DB for Excel sheets |
| Scheduler | APScheduler (`AsyncIOScheduler`) | In-process, matches FastAPI's event loop, no extra service |
| Package manager (Python) | `uv` | Fast, modern, lockfile-based |
| Frontend framework | Next.js 15 (App Router, static export) | Maximum UI flexibility; single-container deploy (no Node runtime in prod) |
| Chat UI primitives | Vercel AI SDK (`useChat`) | Reference React library for LLM streaming UIs |
| Component library | shadcn/ui | Copy-in components, easy to customize |
| Frontend test runner | Vitest + React Testing Library | Fast, idiomatic for Vite/Next.js |
| Container runtime | Docker + docker-compose | Matches existing monorepo pattern |

## 4. Architecture

### High-level diagram

```
┌────────────────────────────────────────────────────────────┐
│  homelab-chatbot (single Docker container)                 │
│                                                            │
│  ┌──────────────────────────────┐    ┌──────────────────┐  │
│  │ Next.js (static export)      │◄──►│ FastAPI (uvicorn)│  │
│  │ served by FastAPI as static  │    │                  │  │
│  │ assets on `/`                │    │  /api/chat (SSE) │  │
│  └──────────────────────────────┘    │  /api/conv/*     │  │
│                                      │  /api/health     │  │
│                                      │  /api/settings   │  │
│                                      │                  │  │
│                                      │  ┌────────────┐  │  │
│                                      │  │ APScheduler│  │  │
│                                      │  │ (git poll) │  │  │
│                                      │  └────────────┘  │  │
│                                      └──────────────────┘  │
│                                                            │
│  Volumes:                                                  │
│    /data/repos/      — cloned GitHub repos                 │
│    /data/lance/      — LanceDB vector store                │
│    /data/chat.db     — SQLite (conversations, messages)    │
│    /data/kb.db       — SQLite (Excel sheets as tables)     │
│    /data/models/     — sentence-transformers cache         │
└────────────────────────────────────────────────────────────┘
```

### Process model

A single Python process runs:

- FastAPI HTTP server (API routes + static asset serving)
- APScheduler background jobs (git sync, embedding refresh)
- In-memory embedding model (loaded once at startup)

Single-container, single-process deployment. No inter-service coordination needed at this scale.

### Directory layout

```
homelab-chatbot/
├── backend/
│   ├── pyproject.toml
│   ├── uv.lock
│   ├── app/
│   │   ├── main.py               # FastAPI app factory, route registration
│   │   ├── config.py             # Pydantic Settings (env + yaml)
│   │   ├── auth.py               # basic-auth middleware, session cookie
│   │   ├── deps.py               # FastAPI dependencies
│   │   ├── routes/
│   │   │   ├── chat.py           # POST /api/chat (SSE stream)
│   │   │   ├── conversations.py  # CRUD /api/conv/*
│   │   │   ├── settings.py       # GET/PUT /api/settings
│   │   │   └── health.py
│   │   ├── llm/
│   │   │   ├── provider.py       # LLM factory
│   │   │   └── registry.py       # available models per provider
│   │   ├── ingestion/
│   │   │   ├── git_sync.py       # poll, pull, detect changes
│   │   │   ├── markdown.py       # chunker + metadata preservation
│   │   │   ├── excel.py          # xlsx → SQLite tables
│   │   │   ├── embed.py          # sentence-transformers wrapper
│   │   │   └── scheduler.py      # APScheduler bootstrap
│   │   ├── retrieval/
│   │   │   ├── vector_tool.py    # LlamaIndex vector query tool
│   │   │   ├── sql_tool.py       # NL→SQL over kb.db
│   │   │   └── router.py         # agent that selects tools
│   │   ├── storage/
│   │   │   ├── chat_db.py        # SQLAlchemy models: Conversation, Message
│   │   │   └── lance.py          # LanceDB client wrapper
│   │   └── models/               # Pydantic request/response schemas
│   ├── tests/
│   │   ├── unit/
│   │   ├── integration/          # opt-in, real LLM calls
│   │   └── retrieval/            # golden-set tests
│   └── Dockerfile.backend        # (optional — only for backend-only builds)
├── frontend/
│   ├── package.json
│   ├── next.config.ts            # output: 'export'
│   ├── app/
│   │   ├── layout.tsx
│   │   ├── page.tsx              # main chat view
│   │   ├── login/page.tsx
│   │   └── settings/page.tsx
│   ├── components/
│   │   ├── chat/                 # MessageList, Composer, StreamingMessage
│   │   ├── sidebar/              # ConversationList, NewChatButton
│   │   ├── settings/             # ProviderPicker, ModelPicker
│   │   └── ui/                   # shadcn/ui primitives
│   ├── lib/
│   │   ├── api.ts                # fetch wrappers
│   │   └── hooks.ts              # useChat, useConversations
│   └── tests/                    # vitest + RTL
├── docker-compose.yml
├── Dockerfile                    # multi-stage: node build → python runtime
├── Makefile
├── config.yaml.example
└── README.md
```

## 5. Data Model

### `chat.db` — SQLite via SQLAlchemy

```
conversations
  id            TEXT PRIMARY KEY    -- uuid4
  title         TEXT                -- auto-generated from first user message
  provider      TEXT NOT NULL       -- 'anthropic' | 'google' | 'ollama'
  model         TEXT NOT NULL
  created_at    TIMESTAMP
  updated_at    TIMESTAMP

messages
  id            INTEGER PRIMARY KEY AUTOINCREMENT
  conv_id       TEXT NOT NULL  FOREIGN KEY → conversations.id  ON DELETE CASCADE
  role          TEXT NOT NULL       -- 'user' | 'assistant' | 'tool'
  content       TEXT NOT NULL
  tool_calls    TEXT                -- JSON array (when role='assistant' and tools were called)
  tool_name     TEXT                -- (when role='tool')
  partial       BOOLEAN NOT NULL DEFAULT 0  -- true if assistant response was interrupted mid-stream
  created_at    TIMESTAMP
  INDEX(conv_id, created_at)

settings
  key           TEXT PRIMARY KEY
  value         TEXT                -- JSON-encoded scalar or object
```

### `kb.db` — SQLite, one table per Excel sheet

- Sheet names and column names are normalized to `snake_case`.
- Opened with `PRAGMA query_only = ON` for all query paths (defense-in-depth against prompt-injected SQL).
- Separate `_kb_meta` table stores `source_hash`, `last_rebuilt_at`, sheet→table name map, and per-column descriptions (used to build the NL-SQL prompt).

### LanceDB — `markdown_chunks` table

```
id            string    -- stable hash of (repo, file_path, line_start, chunk_text)
vector        vector    -- 384-dim (bge-small-en-v1.5)
text          string
repo          string    -- 'homelab-docs' | 'homelab-inventory' (configurable)
file_path     string    -- relative to repo root
line_start    int
line_end      int
commit_sha    string    -- provenance, set at ingest time
heading_path  string    -- e.g., 'Networking > VLANs > Management'
```

Per-chunk metadata is preserved end-to-end so citations can be added later without re-architecting.

## 6. Ingestion Pipeline

### Scheduler behaviour

```
every N seconds (default 180, configurable via config.yaml):
  for each configured repo:
    git fetch origin <branch>
    if local HEAD == remote HEAD:
      skip
    else:
      git pull --ff-only
      changes = git diff --name-status <old_sha>..<new_sha>
      for each changed markdown file:
        reindex_markdown(file, status)
      for each changed xlsx file:
        rebuild_excel_kb(file)
      persist new commit_sha to state file

on boot:
  if index is empty or stale (state file missing / commit_sha mismatch):
    full reindex
```

State is persisted in `/data/sync_state.json`:

```json
{
  "homelab-docs": { "commit_sha": "abc123", "last_synced_at": "2026-04-16T10:30:00Z" },
  "homelab-inventory": { "commit_sha": "def456", "last_synced_at": "2026-04-16T10:30:00Z" }
}
```

### Markdown chunking

- Library: LlamaIndex `MarkdownNodeParser`
- Parameters: `chunk_size=512 tokens`, `chunk_overlap=50 tokens`
- Respects header boundaries (splits on `##` / `###`) — preserves semantic coherence
- Each chunk carries `{repo, file_path, line_start, line_end, commit_sha, heading_path}` metadata

### Excel → SQLite conversion

```
for each sheet in xlsx:
  df = pandas.read_excel(path, sheet_name=sheet)
  df.columns = [normalize_snake_case(c) for c in df.columns]
  table_name = normalize_snake_case(sheet)
  df.to_sql(table_name, kb_db, if_exists='replace', index=False)

persist to _kb_meta:
  source_hash, last_rebuilt_at, sheet→table map, column descriptions
```

Full rebuild on any `.xlsx` change — at <100 rows per sheet it's effectively instant and avoids reconciliation bugs.

### Embedding strategy

- Model loaded once at startup; held in memory
- Batch-embedding used during bulk reindex (batch size 32)
- Replacement strategy for changed files: delete all existing chunks for that `(repo, file_path)` → insert new chunks. Per-chunk diffing would be over-engineering at this scale.

## 7. Query Pipeline

### Per-message flow

```
user message arrives (POST /api/chat)
  ↓
load conversation from chat.db (last N turns, default 10)
  ↓
build LlamaIndex AgentRunner:
  - LLM = build_llm(conversation.provider, conversation.model)
  - tools = [search_homelab_docs, query_homelab_inventory]
  - memory = recent turns
  - system prompt (describes available tools + data sources)
  ↓
agent.astream_chat(user_message):
  - may answer from memory alone → stream answer
  - may call search_homelab_docs → retrieved chunks → synthesize answer
  - may call query_homelab_inventory → rows → synthesize answer
  - may call both tools, then synthesize
  ↓
tokens stream to frontend via SSE (Vercel AI SDK data-stream protocol)
  ↓
on completion:
  persist user message + assistant message + tool calls to chat.db
  update conversations.updated_at
```

### Tools exposed to the agent

**`search_homelab_docs(query: str, repo: str | None = None) -> list[Chunk]`**
- Wraps `VectorIndexRetriever` over LanceDB
- Returns top-K=5 chunks (configurable) with text + metadata
- Optional `repo` filter narrows retrieval when the user specifies a source

**`query_homelab_inventory(question: str) -> {sql: str, rows: list[dict]}`**
- Wraps LlamaIndex `NLSQLTableQueryEngine`
- Uses a dedicated internal LLM call to generate SQL from the natural-language question
- Executes generated SQL against `kb.db` (read-only connection)
- Returns generated SQL alongside results for debugging / future UI

### LLM abstraction

```python
def build_llm(provider: str, model: str, **opts) -> LLM:
    match provider:
        case "anthropic": return Anthropic(model=model, api_key=..., **opts)
        case "google":    return GoogleGenAI(model=model, api_key=..., **opts)
        case "ollama":    return Ollama(model=model, base_url=..., request_timeout=120, **opts)
```

`llm/registry.py` exposes per-provider model lists. Ollama's list is fetched live from `<ollama_host>/api/tags`; Claude/Gemini lists are static and bumped on provider releases.

### Ollama tool-capability handling

Not all Ollama models support native tool-calling. A hardcoded whitelist (configurable in `config.yaml`) declares which models support tools:

```yaml
llm:
  ollama:
    tool_capable_models: [llama3.1, qwen2.5, mistral-nemo]
```

If a user selects a non-whitelisted Ollama model, the router downgrades: vector RAG only, SQL tool disabled, a banner in the UI explains the limitation. No prompt-based tool-invocation fallback — honest failure mode over fragile emulation.

### Streaming protocol

FastAPI emits SSE events compatible with Vercel AI SDK's `data-stream` protocol:

```
event: text-delta
data: {"text": "the "}

event: tool-call
data: {"name": "search_homelab_docs", "args": {"query": "..."}}

event: tool-result
data: {"name": "search_homelab_docs", "summary": "5 chunks"}

event: done
data: {}
```

~30 lines of adapter translate LlamaIndex agent events into this envelope.

### Error handling

| Failure | Behaviour |
|---------|-----------|
| `git pull` fails | Log, skip cycle, retry next interval. Surface `last_sync_at` via `/api/health`. |
| Embedding of a single file fails | Log, skip file, continue. Don't block the whole ingestion. |
| LLM call fails mid-stream | Emit SSE `error` event; persist partial response with `partial=true`. |
| SQL tool errors | Return `{error: "..."}` to agent; agent can retry with reformulated query or apologize. |
| Ollama unreachable | `/api/settings` surfaces availability; chat request fails with clear banner. |

## 8. Frontend

### Pages

| Route | Purpose |
|-------|---------|
| `/login` | Basic-auth form; POSTs to `/api/auth/login`; sets session cookie |
| `/` | Main chat (sidebar + thread) |
| `/settings` | API keys (write-only masked), default provider/model, Ollama host |

### Key components

| Component | Role |
|-----------|------|
| `<ChatThread>` | Renders message list, auto-scrolls, markdown-renders assistant messages |
| `<Composer>` | Textarea + send; Enter sends, Shift+Enter newline; disables during streaming |
| `<StreamingMessage>` | Renders in-flight assistant message; shows tool-call badges |
| `<ConversationSidebar>` | Lists conversations by `updated_at` desc; New/Rename/Delete |
| `<ProviderPicker>` | Provider dropdown + cascading model dropdown; per-conversation |
| `<ToolCallBadge>` | Pill: `🔍 searching docs` / `📊 querying inventory` |
| `<ErrorBanner>` | Dismissible banner for LLM errors, stale-sync warnings, tool-capability warnings |

### Data flow

- Vercel AI SDK `useChat({ api: '/api/chat' })` hook handles streaming, cancellation, retries
- Conversation list: SWR against `/api/conv` (revalidate on focus + after mutations)
- Settings form: React Hook Form + Zod; PUT to `/api/settings`
- Local UI state: `useState` / `useReducer`. No global state library in v1.

## 9. Deployment

### Dockerfile (multi-stage)

```dockerfile
# Stage 1 — Next.js static export
FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build   # produces ./out

# Stage 2 — Python runtime
FROM python:3.13-slim AS runtime
RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY backend/pyproject.toml backend/uv.lock ./
RUN pip install --no-cache-dir uv && uv sync --frozen --no-dev
COPY backend/app ./app
COPY --from=frontend-builder /app/out ./static
EXPOSE 8000
CMD ["uv", "run", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
```

### docker-compose.yml

```yaml
services:
  chatbot:
    build: .
    ports: ["8000:8000"]
    environment:
      - HOMELAB_CHATBOT_CONFIG=/config/config.yaml
      - AUTH_PASSWORD_HASH=${AUTH_PASSWORD_HASH}
      - SESSION_SECRET=${SESSION_SECRET}
      - GITHUB_TOKEN_DOCS=${GITHUB_TOKEN_DOCS}
      - GITHUB_TOKEN_INV=${GITHUB_TOKEN_INV}
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - GOOGLE_API_KEY=${GOOGLE_API_KEY}
      - OLLAMA_HOST=http://host.docker.internal:11434
    volumes:
      - ./config:/config:ro
      - chatbot-data:/data
    restart: unless-stopped
volumes:
  chatbot-data:
```

FastAPI serves the static frontend from `/app/static` at `/`, and API routes at `/api/*`. Single port, single container.

## 10. Config & Secrets

### `config.yaml.example`

```yaml
sync:
  interval_seconds: 180
  state_file: /data/sync_state.json

repos:
  - name: homelab-docs
    url: https://github.com/user/homelab-docs.git
    branch: main
    token_env: GITHUB_TOKEN_DOCS
    include_globs: ["**/*.md"]
  - name: homelab-inventory
    url: https://github.com/user/homelab-inventory.git
    branch: main
    token_env: GITHUB_TOKEN_INV
    include_globs: ["**/*.md", "**/inventory.xlsx"]

embeddings:
  model: BAAI/bge-small-en-v1.5
  cache_dir: /data/models

vector_store:
  path: /data/lance

chat_db:
  path: /data/chat.db

kb_db:
  path: /data/kb.db

llm:
  default_provider: anthropic
  default_model: claude-sonnet-4-6
  ollama:
    host: http://host.docker.internal:11434
    tool_capable_models: [llama3.1, qwen2.5, mistral-nemo]

retrieval:
  top_k: 5
  memory_turns: 10
```

### Secrets — env vars only

- `AUTH_PASSWORD_HASH` — bcrypt hash of the shared password
- `SESSION_SECRET` — random 32-byte string for cookie signing
- `GITHUB_TOKEN_DOCS`, `GITHUB_TOKEN_INV` — PATs with read access to the private repos
- `ANTHROPIC_API_KEY`, `GOOGLE_API_KEY` — optional; required only if user picks that provider

Pydantic `Settings` validates required secrets on startup; app fails fast if any are missing.

### Auth

- Bcrypt password hash generated locally (`make hash-password` target wraps `python -c 'import bcrypt...'`)
- FastAPI middleware verifies `Authorization` header (for API) or session cookie (for browser)
- Login endpoint issues a signed session cookie (itsdangerous or `fastapi-sessions`)
- LAN-only deployment assumed; no rate limiting or brute-force protection in v1 (can add a simple counter later if needed)

## 11. Testing Strategy

### Backend

**Unit tests (`backend/tests/unit/`) — fast, offline, no LLM, no network**

- `git_sync`: change detection, diff parsing (mock `subprocess` calls)
- `markdown` chunker: fixture MD files → asserted chunk count and metadata
- `excel` loader: fixture xlsx → asserted SQLite schema and row counts
- `config` loading: happy path + missing required fields
- `auth`: password verification, session cookie round-trip
- `llm/registry`: provider→models mapping, Ollama list fetch (mock HTTP)
- Chat persistence: conversation CRUD, message append/retrieve
- SSE envelope formatting

Target: ~30–40 tests. `pytest` + `pytest-asyncio`. Total runtime <5s.

**Retrieval golden-set tests (`backend/tests/retrieval/`) — offline, deterministic, ~10–20s total**

- ~15 hand-written `(query, expected_file_path_substring)` pairs
- Real embedding model, real LanceDB (built in `tmp` dir from fixture Markdown)
- Assertion: expected file path appears in top-K retrieval results
- Catches regressions when chunking or retrievers are tuned

**Integration tests (`backend/tests/integration/`) — opt-in, network-required**

- Gated by `RUN_INTEGRATION_TESTS=1` (matches monorepo convention)
- One test per provider: trivial factual question with both tools enabled; assert non-empty response containing an expected keyword (not exact match)
- Uses cheapest/fastest model per provider (Haiku, Gemini Flash, small Ollama model)
- One git-sync test against a scratch public repo

### Frontend

**Component tests (Vitest + React Testing Library) — fast, offline**

- `<ChatThread>`: renders messages, handles streaming state
- `<Composer>`: Enter-to-send, Shift+Enter-for-newline, disabled during streaming
- `<ConversationSidebar>`: sort, rename, delete
- `<ProviderPicker>`: cascading dropdown, per-conversation persistence

Target: ~15 tests.

**E2E:** None in v1. Integration tests cover the backend streaming seam; component tests cover UI logic. Playwright can be added later if regressions warrant.

### Shared conventions

- `MockLLM` helper returns canned streaming responses + tool calls; used across unit tests for route handlers
- CI (GitHub Actions): unit + retrieval tests on every PR; integration tests nightly or on-demand (gated by API key availability)
- No numeric coverage target. Rule: "every non-trivial branch in ingestion, retrieval routing, and chat persistence is tested."

## 12. Makefile

```
build              # multi-stage Docker build
build-dev          # backend + frontend dev servers with hot reload
dev                # docker-compose up with live code mounts
test               # backend unit + retrieval + frontend component tests
test-integration   # RUN_INTEGRATION_TESTS=1, requires API keys
lint               # ruff + mypy + eslint + tsc --noEmit
fmt                # ruff format + prettier
hash-password      # generate bcrypt hash for AUTH_PASSWORD_HASH
clean
```

## 13. Build Sequence (High-Level)

The implementation plan (separate document) will break this into discrete tasks, but the recommended build order is:

1. Backend scaffold: FastAPI app, config loading, health endpoint
2. Storage layer: `chat.db` schema + SQLAlchemy models + tests
3. Ingestion — Markdown: git sync + chunker + LanceDB write + unit tests
4. Ingestion — Excel: xlsx → SQLite + unit tests
5. Retrieval — vector tool + SQL tool + golden-set tests
6. LLM abstraction: provider factory + registry + tool-capability whitelist
7. Agent/router + SSE streaming route
8. Auth middleware + session cookie
9. Conversations/messages CRUD routes
10. Frontend scaffold: Next.js app, shadcn setup, auth/login page
11. Frontend chat UI: ChatThread, Composer, StreamingMessage, `useChat` integration
12. Frontend sidebar + conversation management
13. Frontend settings page
14. Docker multi-stage build + compose + README
15. Integration tests + CI workflow

## 14. Open Questions / Future Work

- **Citations in responses** — metadata is preserved; UI wiring is a future feature
- **Full-text search over chat history** — trivial SQLite FTS5 addition when needed
- **Webhook-based reindex** — requires public endpoint or tunnel; not needed at current update frequency
- **Per-repo branch override per conversation** ("show me what's on the `draft` branch of `homelab-docs`") — speculative; not in v1
- **Observability** — structured logging is in scope; metrics/tracing deferred
- **Cost tracking for API usage** — deferred; can be added by intercepting at the LLM wrapper layer

## 15. References

- LlamaIndex: https://docs.llamaindex.ai/
- Vercel AI SDK (React): https://sdk.vercel.ai/docs
- LanceDB: https://lancedb.github.io/lancedb/
- shadcn/ui: https://ui.shadcn.com/
- Existing repo convention: `/home/georg/projects/adeotek-tools/AGENTS.md`
