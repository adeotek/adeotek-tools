# homelab-chatbot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a self-hosted RAG chatbot for a home-lab knowledge base that ingests two private GitHub repos (Markdown + a multi-sheet Excel spreadsheet) and answers natural-language questions via Ollama, Claude, or Gemini.

**Architecture:** Python 3.13 + FastAPI backend with LlamaIndex for RAG, LanceDB for vectors, SQLite for chat history and Excel-as-SQL. Next.js 15 (static export) frontend served by FastAPI, using Vercel AI SDK `useChat` for streaming. Router agent dispatches between a vector retrieval tool (Markdown) and an NL→SQL tool (Excel). Single Docker container deployed via docker-compose on a LAN homelab host.

**Tech Stack:** Python 3.13, `uv`, FastAPI, uvicorn, SQLAlchemy 2.0 (async + aiosqlite), LlamaIndex, LanceDB, sentence-transformers (`bge-small-en-v1.5`), APScheduler, bcrypt + itsdangerous; Next.js 15, React 19, TypeScript, Vercel AI SDK (`ai`), shadcn/ui, Tailwind, SWR, Vitest + React Testing Library; pytest + pytest-asyncio; Docker multi-stage build.

**Spec:** `docs/superpowers/specs/2026-04-16-homelab-chatbot-design.md`

**Working directory for all paths:** `homelab-chatbot/` under the monorepo root.

---

## Conventions

- All file paths are relative to the monorepo root (`/home/georg/projects/adeotek-tools/`).
- Commit message prefix for this project: `[feat:homelab-chatbot]` (or `[fix:...]`, `[test:...]`, `[docs:...]`).
- Run all backend commands from `homelab-chatbot/backend/`; frontend from `homelab-chatbot/frontend/`.
- After every task, run `make lint` at the appropriate level before committing.

## Phase Map

- **Phase 0 (Tasks 1–3):** Scaffolding — directories, backend package, frontend package
- **Phase 1 (Tasks 4–7):** Backend foundation — config, chat DB, auth, health endpoint
- **Phase 2 (Tasks 8–13):** Ingestion — git sync, markdown chunker, embeddings, LanceDB, Excel loader, scheduler
- **Phase 3 (Tasks 14–16):** Retrieval — vector tool, SQL tool, golden-set tests
- **Phase 4 (Tasks 17–20):** LLM & agent — provider factory, registry, tool-capability, agent + SSE
- **Phase 5 (Tasks 21–24):** API routes — chat, conversations CRUD, settings, error envelope
- **Phase 6 (Tasks 25–28):** Frontend — login, chat thread + streaming, conversation sidebar, settings + provider picker
- **Phase 7 (Tasks 29–33):** Deployment & finishing — Dockerfile, docker-compose + config samples, Makefile, integration tests, README + CI

---

## Phase 0 — Scaffolding

### Task 1: Create project directory structure

**Files:**
- Create: `homelab-chatbot/` (directory)
- Create: `homelab-chatbot/.gitignore`
- Create: `homelab-chatbot/README.md` (stub)

- [ ] **Step 1: Create directories**

Run:
```bash
mkdir -p homelab-chatbot/backend/app/{routes,llm,ingestion,retrieval,storage,models}
mkdir -p homelab-chatbot/backend/tests/{unit,integration,retrieval,fixtures}
mkdir -p homelab-chatbot/frontend
mkdir -p homelab-chatbot/config
```

- [ ] **Step 2: Write `homelab-chatbot/.gitignore`**

```gitignore
# Python
__pycache__/
*.py[cod]
*.egg-info/
.venv/
.pytest_cache/
.mypy_cache/
.ruff_cache/
.coverage
htmlcov/

# Node
node_modules/
.next/
out/

# App data
/data/
/config/config.yaml
/config/.env
*.lance/

# Editor
.vscode/
.idea/
*.swp
.DS_Store
```

- [ ] **Step 3: Write stub `homelab-chatbot/README.md`**

```markdown
# homelab-chatbot

Self-hosted RAG chatbot for a home-lab knowledge base (Markdown + Excel), supporting Ollama, Claude, and Gemini as peer LLM providers.

See `docs/superpowers/specs/2026-04-16-homelab-chatbot-design.md` for the full design.

## Quick start

Stub — replaced with full runtime docs in Task 33.
```

- [ ] **Step 4: Verify structure**

Run: `find homelab-chatbot -type d | sort`
Expected: lists all created directories including `backend/app/routes`, `backend/tests/unit`, `frontend`, `config`.

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/.gitignore homelab-chatbot/README.md
git commit -m "[feat:homelab-chatbot] scaffold project directory structure"
```

---

### Task 2: Backend Python package bootstrap

**Files:**
- Create: `homelab-chatbot/backend/pyproject.toml`
- Create: `homelab-chatbot/backend/app/__init__.py`
- Create: `homelab-chatbot/backend/app/main.py`
- Create: `homelab-chatbot/backend/tests/__init__.py`
- Create: `homelab-chatbot/backend/tests/unit/__init__.py`
- Create: `homelab-chatbot/backend/tests/unit/test_main.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/pyproject.toml`**

```toml
[project]
name = "homelab-chatbot"
version = "0.1.0"
description = "Self-hosted RAG chatbot for a home-lab knowledge base"
requires-python = ">=3.13"
dependencies = [
    "fastapi>=0.115",
    "uvicorn[standard]>=0.32",
    "pydantic>=2.9",
    "pydantic-settings>=2.6",
    "pyyaml>=6.0",
    "sqlalchemy[asyncio]>=2.0",
    "aiosqlite>=0.20",
    "bcrypt>=4.2",
    "itsdangerous>=2.2",
    "httpx>=0.27",
    "apscheduler>=3.10",
    "gitpython>=3.1",
    "pandas>=2.2",
    "openpyxl>=3.1",
    "sentence-transformers>=3.3",
    "lancedb>=0.15",
    "llama-index-core>=0.12",
    "llama-index-vector-stores-lancedb>=0.3",
    "llama-index-llms-anthropic>=0.6",
    "llama-index-llms-google-genai>=0.1",
    "llama-index-llms-ollama>=0.5",
    "llama-index-embeddings-huggingface>=0.4",
]

[dependency-groups]
dev = [
    "pytest>=8.3",
    "pytest-asyncio>=0.24",
    "pytest-cov>=6.0",
    "ruff>=0.8",
    "mypy>=1.13",
    "httpx>=0.27",
]

[tool.pytest.ini_options]
asyncio_mode = "auto"
testpaths = ["tests"]
addopts = "-v"

[tool.ruff]
line-length = 100
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "B", "UP", "SIM", "C4"]

[tool.mypy]
python_version = "3.13"
strict = true
ignore_missing_imports = true
```

- [ ] **Step 2: Write empty package markers**

Create:
- `homelab-chatbot/backend/app/__init__.py` — empty file
- `homelab-chatbot/backend/tests/__init__.py` — empty file
- `homelab-chatbot/backend/tests/unit/__init__.py` — empty file

- [ ] **Step 3: Write failing test `homelab-chatbot/backend/tests/unit/test_main.py`**

```python
from fastapi.testclient import TestClient

from app.main import create_app


def test_app_factory_returns_fastapi_instance():
    app = create_app()
    assert app.title == "homelab-chatbot"


def test_health_endpoint_exists():
    app = create_app()
    client = TestClient(app)
    response = client.get("/api/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}
```

- [ ] **Step 4: Install dependencies and run failing test**

Run:
```bash
cd homelab-chatbot/backend
uv sync
uv run pytest tests/unit/test_main.py -v
```
Expected: FAIL with `ModuleNotFoundError: No module named 'app.main'`

- [ ] **Step 5: Write minimal `homelab-chatbot/backend/app/main.py`**

```python
"""FastAPI application factory."""

from fastapi import FastAPI


def create_app() -> FastAPI:
    """Build and return the FastAPI application."""
    app = FastAPI(title="homelab-chatbot")

    @app.get("/api/health")
    def health() -> dict[str, str]:
        return {"status": "ok"}

    return app


app = create_app()
```

- [ ] **Step 6: Run test — verify pass**

Run: `uv run pytest tests/unit/test_main.py -v`
Expected: PASS (2 tests).

- [ ] **Step 7: Commit**

```bash
git add homelab-chatbot/backend/
git commit -m "[feat:homelab-chatbot] bootstrap backend Python package with health endpoint"
```

---

### Task 3: Frontend Next.js bootstrap with static export

**Files:**
- Create: `homelab-chatbot/frontend/package.json`
- Create: `homelab-chatbot/frontend/next.config.ts`
- Create: `homelab-chatbot/frontend/tsconfig.json`
- Create: `homelab-chatbot/frontend/app/layout.tsx`
- Create: `homelab-chatbot/frontend/app/page.tsx`
- Create: `homelab-chatbot/frontend/app/globals.css`
- Create: `homelab-chatbot/frontend/postcss.config.mjs`
- Create: `homelab-chatbot/frontend/tailwind.config.ts`
- Create: `homelab-chatbot/frontend/.eslintrc.json`
- Create: `homelab-chatbot/frontend/vitest.config.ts`
- Create: `homelab-chatbot/frontend/tests/setup.ts`

- [ ] **Step 1: Write `homelab-chatbot/frontend/package.json`**

```json
{
  "name": "homelab-chatbot-frontend",
  "version": "0.1.0",
  "private": true,
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "next lint",
    "typecheck": "tsc --noEmit",
    "test": "vitest run",
    "test:watch": "vitest"
  },
  "dependencies": {
    "next": "15.1.0",
    "react": "19.0.0",
    "react-dom": "19.0.0",
    "ai": "^4.0.0",
    "@ai-sdk/react": "^1.0.0",
    "swr": "^2.2.5",
    "react-hook-form": "^7.54.0",
    "zod": "^3.24.0",
    "@hookform/resolvers": "^3.9.0",
    "clsx": "^2.1.1",
    "tailwind-merge": "^2.5.0",
    "react-markdown": "^9.0.1",
    "lucide-react": "^0.460.0"
  },
  "devDependencies": {
    "@types/node": "^22",
    "@types/react": "^19",
    "@types/react-dom": "^19",
    "typescript": "^5.7",
    "tailwindcss": "^3.4",
    "postcss": "^8.4",
    "autoprefixer": "^10.4",
    "eslint": "^9",
    "eslint-config-next": "15.1.0",
    "vitest": "^2.1",
    "@vitejs/plugin-react": "^4.3",
    "@testing-library/react": "^16.1",
    "@testing-library/jest-dom": "^6.6",
    "@testing-library/user-event": "^14.5",
    "jsdom": "^25.0"
  }
}
```

- [ ] **Step 2: Write `homelab-chatbot/frontend/next.config.ts`**

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: "export",
  images: { unoptimized: true },
  trailingSlash: true,
};

export default nextConfig;
```

- [ ] **Step 3: Write `homelab-chatbot/frontend/tsconfig.json`**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": false,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": { "@/*": ["./*"] }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules", "out"]
}
```

- [ ] **Step 4: Write base layout and page**

`homelab-chatbot/frontend/app/layout.tsx`:
```tsx
import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "homelab-chatbot",
  description: "Home-lab knowledge base chatbot",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-neutral-950 text-neutral-100 antialiased">{children}</body>
    </html>
  );
}
```

`homelab-chatbot/frontend/app/page.tsx`:
```tsx
export default function Home() {
  return (
    <main className="flex min-h-screen items-center justify-center">
      <p className="text-xl">homelab-chatbot</p>
    </main>
  );
}
```

`homelab-chatbot/frontend/app/globals.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;
```

- [ ] **Step 5: Write Tailwind + PostCSS config**

`homelab-chatbot/frontend/tailwind.config.ts`:
```typescript
import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: { extend: {} },
  plugins: [],
};

export default config;
```

`homelab-chatbot/frontend/postcss.config.mjs`:
```javascript
export default {
  plugins: { tailwindcss: {}, autoprefixer: {} },
};
```

- [ ] **Step 6: Write Vitest config and setup**

`homelab-chatbot/frontend/vitest.config.ts`:
```typescript
import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
    setupFiles: ["./tests/setup.ts"],
    globals: true,
  },
  resolve: { alias: { "@": path.resolve(__dirname, ".") } },
});
```

`homelab-chatbot/frontend/tests/setup.ts`:
```typescript
import "@testing-library/jest-dom/vitest";
```

- [ ] **Step 7: Install and verify build**

Run:
```bash
cd homelab-chatbot/frontend
npm install
npm run build
```
Expected: build succeeds; `out/` directory contains `index.html`.

- [ ] **Step 8: Commit**

```bash
git add homelab-chatbot/frontend/
git commit -m "[feat:homelab-chatbot] bootstrap Next.js frontend with static export"
```

---

## Phase 1 — Backend Foundation

### Task 4: Configuration loading (Pydantic Settings + YAML)

**Files:**
- Create: `homelab-chatbot/backend/app/config.py`
- Create: `homelab-chatbot/backend/tests/unit/test_config.py`
- Create: `homelab-chatbot/backend/tests/fixtures/config_valid.yaml`
- Create: `homelab-chatbot/config/config.yaml.example`

- [ ] **Step 1: Write fixture `homelab-chatbot/backend/tests/fixtures/config_valid.yaml`**

```yaml
sync:
  interval_seconds: 180
  state_file: /tmp/sync_state.json

repos:
  - name: homelab-docs
    url: https://github.com/user/homelab-docs.git
    branch: main
    token_env: GITHUB_TOKEN_DOCS
    include_globs: ["**/*.md"]

embeddings:
  model: BAAI/bge-small-en-v1.5
  cache_dir: /tmp/models

vector_store:
  path: /tmp/lance

chat_db:
  path: /tmp/chat.db

kb_db:
  path: /tmp/kb.db

llm:
  default_provider: anthropic
  default_model: claude-sonnet-4-6
  ollama:
    host: http://localhost:11434
    tool_capable_models: [llama3.1, qwen2.5]

retrieval:
  top_k: 5
  memory_turns: 10
```

- [ ] **Step 2: Write failing test `homelab-chatbot/backend/tests/unit/test_config.py`**

```python
from pathlib import Path

import pytest

from app.config import AppConfig, load_config

FIXTURE_DIR = Path(__file__).parent.parent / "fixtures"


def test_load_config_from_yaml():
    cfg = load_config(FIXTURE_DIR / "config_valid.yaml")
    assert isinstance(cfg, AppConfig)
    assert cfg.sync.interval_seconds == 180
    assert len(cfg.repos) == 1
    assert cfg.repos[0].name == "homelab-docs"
    assert cfg.llm.default_provider == "anthropic"
    assert "llama3.1" in cfg.llm.ollama.tool_capable_models


def test_load_config_missing_file_raises():
    with pytest.raises(FileNotFoundError):
        load_config(Path("/nonexistent/config.yaml"))


def test_invalid_provider_rejected(tmp_path: Path):
    bad = tmp_path / "bad.yaml"
    bad.write_text(
        "sync: {interval_seconds: 60, state_file: /tmp/s.json}\n"
        "repos: []\n"
        "embeddings: {model: m, cache_dir: /tmp}\n"
        "vector_store: {path: /tmp}\n"
        "chat_db: {path: /tmp/c.db}\n"
        "kb_db: {path: /tmp/k.db}\n"
        "llm:\n"
        "  default_provider: invalid_provider\n"
        "  default_model: x\n"
        "  ollama: {host: http://localhost:11434, tool_capable_models: []}\n"
        "retrieval: {top_k: 5, memory_turns: 10}\n"
    )
    with pytest.raises(ValueError):
        load_config(bad)
```

- [ ] **Step 3: Run test — verify fails**

Run: `cd homelab-chatbot/backend && uv run pytest tests/unit/test_config.py -v`
Expected: FAIL — `ModuleNotFoundError: No module named 'app.config'`

- [ ] **Step 4: Write `homelab-chatbot/backend/app/config.py`**

```python
"""Application configuration loaded from YAML + environment variables."""

from pathlib import Path
from typing import Literal

import yaml
from pydantic import BaseModel, Field, field_validator

Provider = Literal["anthropic", "google", "ollama"]


class SyncConfig(BaseModel):
    interval_seconds: int = Field(ge=30, le=3600)
    state_file: str


class RepoConfig(BaseModel):
    name: str
    url: str
    branch: str = "main"
    token_env: str
    include_globs: list[str] = Field(default_factory=lambda: ["**/*.md"])


class EmbeddingsConfig(BaseModel):
    model: str
    cache_dir: str


class PathConfig(BaseModel):
    path: str


class OllamaConfig(BaseModel):
    host: str
    tool_capable_models: list[str] = Field(default_factory=list)


class LLMConfig(BaseModel):
    default_provider: Provider
    default_model: str
    ollama: OllamaConfig


class RetrievalConfig(BaseModel):
    top_k: int = Field(ge=1, le=50, default=5)
    memory_turns: int = Field(ge=0, le=100, default=10)


class AppConfig(BaseModel):
    sync: SyncConfig
    repos: list[RepoConfig]
    embeddings: EmbeddingsConfig
    vector_store: PathConfig
    chat_db: PathConfig
    kb_db: PathConfig
    llm: LLMConfig
    retrieval: RetrievalConfig

    @field_validator("repos")
    @classmethod
    def unique_repo_names(cls, v: list[RepoConfig]) -> list[RepoConfig]:
        names = [r.name for r in v]
        if len(names) != len(set(names)):
            raise ValueError("repo names must be unique")
        return v


def load_config(path: Path) -> AppConfig:
    """Load configuration from a YAML file and validate it."""
    if not path.exists():
        raise FileNotFoundError(f"config file not found: {path}")
    with path.open() as f:
        raw = yaml.safe_load(f)
    return AppConfig.model_validate(raw)
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_config.py -v`
Expected: PASS (3 tests).

- [ ] **Step 6: Copy fixture to `homelab-chatbot/config/config.yaml.example`**

Same content as the fixture, with `/data/...` paths instead of `/tmp/...`, and both repos listed:

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

- [ ] **Step 7: Commit**

```bash
git add homelab-chatbot/backend/app/config.py homelab-chatbot/backend/tests/ homelab-chatbot/config/config.yaml.example
git commit -m "[feat:homelab-chatbot] add YAML config loader with Pydantic validation"
```

---

### Task 5: Secrets loader (env vars via Pydantic Settings)

**Files:**
- Create: `homelab-chatbot/backend/app/secrets.py`
- Create: `homelab-chatbot/backend/tests/unit/test_secrets.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_secrets.py`**

```python
import pytest

from app.secrets import Secrets


def test_required_secrets_present(monkeypatch):
    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", "$2b$12$abc")
    monkeypatch.setenv("HLCB_SESSION_SECRET", "s" * 32)
    s = Secrets()
    assert s.auth_password_hash == "$2b$12$abc"
    assert s.session_secret == "s" * 32


def test_missing_required_secret_raises(monkeypatch):
    monkeypatch.delenv("HLCB_AUTH_PASSWORD_HASH", raising=False)
    monkeypatch.delenv("HLCB_SESSION_SECRET", raising=False)
    with pytest.raises(Exception):
        Secrets()


def test_optional_api_keys_default_none(monkeypatch):
    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", "x")
    monkeypatch.setenv("HLCB_SESSION_SECRET", "s" * 32)
    monkeypatch.delenv("HLCB_ANTHROPIC_API_KEY", raising=False)
    s = Secrets()
    assert s.anthropic_api_key is None


def test_github_token_lookup(monkeypatch):
    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", "x")
    monkeypatch.setenv("HLCB_SESSION_SECRET", "s" * 32)
    monkeypatch.setenv("HLCB_GIT_TOKEN_DOCS", "ghp_abc")
    s = Secrets()
    assert s.github_token("HLCB_GIT_TOKEN_DOCS") == "ghp_abc"
    assert s.github_token("HLCB_GIT_TOKEN_MISSING") is None
```

- [ ] **Step 2: Run test — verify fails**

Run: `uv run pytest tests/unit/test_secrets.py -v`
Expected: FAIL — `ModuleNotFoundError`.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/secrets.py`**

```python
"""Environment-variable-backed secrets."""

import os

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Secrets(BaseSettings):
    """Secrets loaded from environment variables."""

    model_config = SettingsConfigDict(case_sensitive=True, extra="ignore")

    auth_password_hash: str = Field(alias="HLCB_AUTH_PASSWORD_HASH")
    session_secret: str = Field(alias="HLCB_SESSION_SECRET", min_length=16)

    anthropic_api_key: str | None = Field(default=None, alias="HLCB_ANTHROPIC_API_KEY")
    google_api_key: str | None = Field(default=None, alias="HLCB_GOOGLE_API_KEY")

    def github_token(self, env_var: str) -> str | None:
        """Look up a GitHub token by the env-var name referenced in config.yaml."""
        return os.environ.get(env_var)
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_secrets.py -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/secrets.py homelab-chatbot/backend/tests/unit/test_secrets.py
git commit -m "[feat:homelab-chatbot] add env-var secrets loader with Pydantic Settings"
```

---

### Task 6: Chat database models (SQLAlchemy async)

**Files:**
- Create: `homelab-chatbot/backend/app/storage/__init__.py`
- Create: `homelab-chatbot/backend/app/storage/chat_db.py`
- Create: `homelab-chatbot/backend/tests/unit/test_chat_db.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/app/storage/__init__.py`** (empty).

- [ ] **Step 2: Write failing test `homelab-chatbot/backend/tests/unit/test_chat_db.py`**

```python
import pytest

from app.storage.chat_db import (
    ChatDB,
    Conversation,
    Message,
)


@pytest.fixture
async def db(tmp_path):
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/test.db")
    await db.init_schema()
    yield db
    await db.close()


async def test_create_conversation(db: ChatDB):
    conv = await db.create_conversation(
        title="my first chat", provider="anthropic", model="claude-sonnet-4-6"
    )
    assert conv.id
    assert conv.title == "my first chat"
    assert conv.provider == "anthropic"


async def test_append_and_list_messages(db: ChatDB):
    conv = await db.create_conversation(title="t", provider="ollama", model="llama3.1")
    await db.append_message(conv.id, role="user", content="hello")
    await db.append_message(conv.id, role="assistant", content="hi there")
    msgs = await db.list_messages(conv.id)
    assert [m.role for m in msgs] == ["user", "assistant"]
    assert msgs[0].content == "hello"


async def test_list_conversations_orders_by_updated_desc(db: ChatDB):
    c1 = await db.create_conversation(title="first", provider="anthropic", model="x")
    c2 = await db.create_conversation(title="second", provider="anthropic", model="x")
    await db.append_message(c1.id, role="user", content="new msg")
    convs = await db.list_conversations()
    assert convs[0].id == c1.id  # c1 updated most recently


async def test_delete_conversation_cascades_messages(db: ChatDB):
    conv = await db.create_conversation(title="t", provider="anthropic", model="x")
    await db.append_message(conv.id, role="user", content="bye")
    await db.delete_conversation(conv.id)
    msgs = await db.list_messages(conv.id)
    assert msgs == []


async def test_partial_flag_defaults_false(db: ChatDB):
    conv = await db.create_conversation(title="t", provider="anthropic", model="x")
    msg = await db.append_message(conv.id, role="assistant", content="x")
    assert msg.partial is False
```

- [ ] **Step 3: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 4: Write `homelab-chatbot/backend/app/storage/chat_db.py`**

```python
"""Async SQLAlchemy models and CRUD for conversations and messages."""

import uuid
from datetime import datetime, timezone

from sqlalchemy import ForeignKey, Index, String, delete, select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship


def _utcnow() -> datetime:
    return datetime.now(timezone.utc)


class Base(DeclarativeBase):
    pass


class Conversation(Base):
    __tablename__ = "conversations"

    id: Mapped[str] = mapped_column(String, primary_key=True)
    title: Mapped[str] = mapped_column(String, nullable=False)
    provider: Mapped[str] = mapped_column(String, nullable=False)
    model: Mapped[str] = mapped_column(String, nullable=False)
    created_at: Mapped[datetime] = mapped_column(default=_utcnow)
    updated_at: Mapped[datetime] = mapped_column(default=_utcnow)

    messages: Mapped[list["Message"]] = relationship(
        back_populates="conversation",
        cascade="all, delete-orphan",
        order_by="Message.created_at",
    )


class Message(Base):
    __tablename__ = "messages"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    conv_id: Mapped[str] = mapped_column(
        ForeignKey("conversations.id", ondelete="CASCADE"), nullable=False
    )
    role: Mapped[str] = mapped_column(String, nullable=False)
    content: Mapped[str] = mapped_column(String, nullable=False)
    tool_calls: Mapped[str | None] = mapped_column(String, nullable=True)
    tool_name: Mapped[str | None] = mapped_column(String, nullable=True)
    partial: Mapped[bool] = mapped_column(default=False, nullable=False)
    created_at: Mapped[datetime] = mapped_column(default=_utcnow)

    conversation: Mapped[Conversation] = relationship(back_populates="messages")

    __table_args__ = (Index("ix_messages_conv_created", "conv_id", "created_at"),)


class ChatDB:
    """Async wrapper around chat.db for conversations and messages."""

    def __init__(self, url: str) -> None:
        self._engine = create_async_engine(url, echo=False, future=True)
        self._session_factory = async_sessionmaker(self._engine, expire_on_commit=False)

    async def init_schema(self) -> None:
        async with self._engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)

    async def close(self) -> None:
        await self._engine.dispose()

    def session(self) -> AsyncSession:
        return self._session_factory()

    async def create_conversation(
        self, title: str, provider: str, model: str
    ) -> Conversation:
        conv = Conversation(
            id=str(uuid.uuid4()), title=title, provider=provider, model=model
        )
        async with self.session() as s:
            s.add(conv)
            await s.commit()
            await s.refresh(conv)
        return conv

    async def append_message(
        self,
        conv_id: str,
        role: str,
        content: str,
        tool_calls: str | None = None,
        tool_name: str | None = None,
        partial: bool = False,
    ) -> Message:
        async with self.session() as s:
            msg = Message(
                conv_id=conv_id,
                role=role,
                content=content,
                tool_calls=tool_calls,
                tool_name=tool_name,
                partial=partial,
            )
            s.add(msg)
            conv = await s.get(Conversation, conv_id)
            if conv:
                conv.updated_at = _utcnow()
            await s.commit()
            await s.refresh(msg)
        return msg

    async def list_messages(self, conv_id: str) -> list[Message]:
        async with self.session() as s:
            result = await s.execute(
                select(Message).where(Message.conv_id == conv_id).order_by(Message.created_at)
            )
            return list(result.scalars().all())

    async def list_conversations(self) -> list[Conversation]:
        async with self.session() as s:
            result = await s.execute(
                select(Conversation).order_by(Conversation.updated_at.desc())
            )
            return list(result.scalars().all())

    async def get_conversation(self, conv_id: str) -> Conversation | None:
        async with self.session() as s:
            return await s.get(Conversation, conv_id)

    async def delete_conversation(self, conv_id: str) -> None:
        async with self.session() as s:
            await s.execute(delete(Conversation).where(Conversation.id == conv_id))
            await s.commit()

    async def rename_conversation(self, conv_id: str, title: str) -> None:
        async with self.session() as s:
            conv = await s.get(Conversation, conv_id)
            if conv:
                conv.title = title
                conv.updated_at = _utcnow()
                await s.commit()
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_chat_db.py -v`
Expected: PASS (5 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/storage/ homelab-chatbot/backend/tests/unit/test_chat_db.py
git commit -m "[feat:homelab-chatbot] add async SQLAlchemy chat database with CRUD"
```

---

### Task 7: Authentication (bcrypt + signed session cookie)

**Files:**
- Create: `homelab-chatbot/backend/app/auth.py`
- Create: `homelab-chatbot/backend/tests/unit/test_auth.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_auth.py`**

```python
import pytest

from app.auth import (
    AuthService,
    hash_password,
    verify_password,
)


def test_hash_and_verify_password():
    h = hash_password("correct horse battery staple")
    assert verify_password("correct horse battery staple", h) is True
    assert verify_password("wrong", h) is False


@pytest.fixture
def auth():
    pwd_hash = hash_password("secret")
    return AuthService(password_hash=pwd_hash, session_secret="x" * 32)


def test_session_roundtrip(auth: AuthService):
    token = auth.issue_session_token()
    assert auth.verify_session_token(token) is True


def test_tampered_session_rejected(auth: AuthService):
    token = auth.issue_session_token()
    tampered = token[:-1] + ("A" if token[-1] != "A" else "B")
    assert auth.verify_session_token(tampered) is False


def test_expired_session_rejected(auth: AuthService):
    token = auth.issue_session_token(max_age_seconds=0)
    import time

    time.sleep(0.1)
    assert auth.verify_session_token(token, max_age_seconds=0) is False


def test_check_password_success(auth: AuthService):
    assert auth.check_password("secret") is True
    assert auth.check_password("wrong") is False
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/auth.py`**

```python
"""Password hashing and signed session cookies."""

import bcrypt
from itsdangerous import BadSignature, SignatureExpired, TimestampSigner

SESSION_TOKEN_PAYLOAD = "authenticated"
DEFAULT_SESSION_MAX_AGE = 60 * 60 * 24 * 7  # 7 days


def hash_password(plain: str) -> str:
    """Hash a plaintext password with bcrypt and return the encoded hash."""
    return bcrypt.hashpw(plain.encode("utf-8"), bcrypt.gensalt()).decode("utf-8")


def verify_password(plain: str, password_hash: str) -> bool:
    """Check a plaintext password against a bcrypt hash."""
    try:
        return bcrypt.checkpw(plain.encode("utf-8"), password_hash.encode("utf-8"))
    except ValueError:
        return False


class AuthService:
    """Issues and validates signed session tokens."""

    def __init__(self, password_hash: str, session_secret: str) -> None:
        self._password_hash = password_hash
        self._signer = TimestampSigner(session_secret)

    def check_password(self, plain: str) -> bool:
        return verify_password(plain, self._password_hash)

    def issue_session_token(self, max_age_seconds: int | None = None) -> str:
        return self._signer.sign(SESSION_TOKEN_PAYLOAD).decode("utf-8")

    def verify_session_token(
        self, token: str, max_age_seconds: int = DEFAULT_SESSION_MAX_AGE
    ) -> bool:
        try:
            payload = self._signer.unsign(token, max_age=max_age_seconds).decode("utf-8")
            return payload == SESSION_TOKEN_PAYLOAD
        except (BadSignature, SignatureExpired):
            return False
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_auth.py -v`
Expected: PASS (5 tests). Note: the `expired` test uses `max_age_seconds=0` and a sleep — if flaky, bump sleep to 0.2.

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/auth.py homelab-chatbot/backend/tests/unit/test_auth.py
git commit -m "[feat:homelab-chatbot] add bcrypt password hashing and signed session tokens"
```

---

## Phase 2 — Ingestion

### Task 8: Git sync (clone, pull, diff detection)

**Files:**
- Create: `homelab-chatbot/backend/app/ingestion/__init__.py`
- Create: `homelab-chatbot/backend/app/ingestion/git_sync.py`
- Create: `homelab-chatbot/backend/tests/unit/test_git_sync.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/app/ingestion/__init__.py`** (empty).

- [ ] **Step 2: Write failing test `homelab-chatbot/backend/tests/unit/test_git_sync.py`**

```python
import subprocess
from pathlib import Path

import pytest

from app.config import RepoConfig
from app.ingestion.git_sync import GitSync, SyncResult


def _init_remote_repo(tmp: Path) -> Path:
    remote = tmp / "remote.git"
    work = tmp / "work"
    work.mkdir()
    subprocess.run(["git", "init", "--bare", str(remote)], check=True)
    subprocess.run(["git", "init", str(work)], check=True)
    subprocess.run(["git", "-C", str(work), "config", "user.email", "t@t"], check=True)
    subprocess.run(["git", "-C", str(work), "config", "user.name", "t"], check=True)
    subprocess.run(["git", "-C", str(work), "config", "commit.gpgsign", "false"], check=True)
    (work / "README.md").write_text("# hello\n")
    subprocess.run(["git", "-C", str(work), "add", "."], check=True)
    subprocess.run(["git", "-C", str(work), "commit", "-m", "init"], check=True)
    subprocess.run(["git", "-C", str(work), "branch", "-M", "main"], check=True)
    subprocess.run(
        ["git", "-C", str(work), "remote", "add", "origin", str(remote)], check=True
    )
    subprocess.run(["git", "-C", str(work), "push", "-u", "origin", "main"], check=True)
    return remote


@pytest.fixture
def local_repo(tmp_path: Path) -> Path:
    return _init_remote_repo(tmp_path)


def test_initial_clone_when_missing(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r", url=f"file://{local_repo}", branch="main", token_env="UNUSED"
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    result = sync.sync(cfg)
    assert result.cloned is True
    assert result.changed_files == []
    assert (tmp_path / "repos" / "r" / "README.md").exists()


def test_pull_detects_changed_files(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r", url=f"file://{local_repo}", branch="main", token_env="UNUSED"
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    sync.sync(cfg)  # initial clone

    work_dir = tmp_path / "upstream_work"
    subprocess.run(
        ["git", "clone", f"file://{local_repo}", str(work_dir)], check=True
    )
    subprocess.run(["git", "-C", str(work_dir), "config", "user.email", "t@t"], check=True)
    subprocess.run(["git", "-C", str(work_dir), "config", "user.name", "t"], check=True)
    subprocess.run(["git", "-C", str(work_dir), "config", "commit.gpgsign", "false"], check=True)
    (work_dir / "docs.md").write_text("# docs\n")
    subprocess.run(["git", "-C", str(work_dir), "add", "."], check=True)
    subprocess.run(["git", "-C", str(work_dir), "commit", "-m", "add"], check=True)
    subprocess.run(["git", "-C", str(work_dir), "push"], check=True)

    result = sync.sync(cfg)
    assert result.cloned is False
    names = [c.path for c in result.changed_files]
    assert "docs.md" in names


def test_no_change_returns_empty(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r", url=f"file://{local_repo}", branch="main", token_env="UNUSED"
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    sync.sync(cfg)
    result = sync.sync(cfg)
    assert result.cloned is False
    assert result.changed_files == []


def test_include_globs_filter(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r",
        url=f"file://{local_repo}",
        branch="main",
        token_env="UNUSED",
        include_globs=["**/*.xlsx"],
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    result = sync.sync(cfg)
    assert result.matched_files == []  # no .xlsx in repo


def test_token_injected_into_https_url(tmp_path: Path):
    cfg = RepoConfig(
        name="r",
        url="https://github.com/user/repo.git",
        branch="main",
        token_env="T",
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda v: "tok123" if v == "T" else None)
    assert (
        sync._auth_url(cfg)
        == "https://x-access-token:tok123@github.com/user/repo.git"
    )
```

- [ ] **Step 3: Run test — verify fails**

Run: `uv run pytest tests/unit/test_git_sync.py -v`
Expected: FAIL — module not found.

- [ ] **Step 4: Write `homelab-chatbot/backend/app/ingestion/git_sync.py`**

```python
"""Clone/pull configured repos and report changed files."""

import fnmatch
import subprocess
from dataclasses import dataclass, field
from pathlib import Path
from typing import Callable, Literal
from urllib.parse import urlparse, urlunparse

from app.config import RepoConfig

ChangeStatus = Literal["A", "M", "D", "R"]


@dataclass
class ChangedFile:
    path: str
    status: ChangeStatus


@dataclass
class SyncResult:
    repo: str
    cloned: bool
    old_sha: str | None
    new_sha: str
    changed_files: list[ChangedFile] = field(default_factory=list)
    matched_files: list[ChangedFile] = field(default_factory=list)


class GitSync:
    """Performs clone/pull operations and reports changes matching configured globs."""

    def __init__(
        self,
        clone_root: Path,
        get_token: Callable[[str], str | None],
    ) -> None:
        self._root = Path(clone_root)
        self._get_token = get_token
        self._root.mkdir(parents=True, exist_ok=True)

    def repo_path(self, cfg: RepoConfig) -> Path:
        return self._root / cfg.name

    def _auth_url(self, cfg: RepoConfig) -> str:
        token = self._get_token(cfg.token_env)
        if not token:
            return cfg.url
        parsed = urlparse(cfg.url)
        if parsed.scheme not in ("http", "https"):
            return cfg.url
        netloc = f"x-access-token:{token}@{parsed.hostname}"
        if parsed.port:
            netloc += f":{parsed.port}"
        return urlunparse(parsed._replace(netloc=netloc))

    def _git(self, *args: str, cwd: Path | None = None) -> str:
        result = subprocess.run(
            ["git", *args],
            cwd=str(cwd) if cwd else None,
            check=True,
            capture_output=True,
            text=True,
        )
        return result.stdout

    def _head_sha(self, path: Path) -> str:
        return self._git("rev-parse", "HEAD", cwd=path).strip()

    def _match(self, paths: list[ChangedFile], globs: list[str]) -> list[ChangedFile]:
        out = []
        for cf in paths:
            if any(fnmatch.fnmatch(cf.path, g) for g in globs):
                out.append(cf)
        return out

    def sync(self, cfg: RepoConfig) -> SyncResult:
        path = self.repo_path(cfg)
        url = self._auth_url(cfg)

        if not (path / ".git").exists():
            if path.exists():
                raise RuntimeError(f"{path} exists but is not a git repo")
            self._git("clone", "--branch", cfg.branch, url, str(path))
            new_sha = self._head_sha(path)
            matched = self._list_matching(path, cfg.include_globs)
            return SyncResult(
                repo=cfg.name,
                cloned=True,
                old_sha=None,
                new_sha=new_sha,
                changed_files=[],
                matched_files=matched,
            )

        old_sha = self._head_sha(path)
        self._git("fetch", "origin", cfg.branch, cwd=path)
        self._git("reset", "--hard", f"origin/{cfg.branch}", cwd=path)
        new_sha = self._head_sha(path)

        changed: list[ChangedFile] = []
        if old_sha != new_sha:
            diff = self._git("diff", "--name-status", f"{old_sha}..{new_sha}", cwd=path)
            for line in diff.strip().splitlines():
                parts = line.split("\t")
                if len(parts) >= 2:
                    status_char = parts[0][0]
                    if status_char in ("A", "M", "D", "R"):
                        changed.append(ChangedFile(path=parts[-1], status=status_char))

        matched = self._match(changed, cfg.include_globs)
        return SyncResult(
            repo=cfg.name,
            cloned=False,
            old_sha=old_sha,
            new_sha=new_sha,
            changed_files=changed,
            matched_files=matched,
        )

    def _list_matching(self, repo_path: Path, globs: list[str]) -> list[ChangedFile]:
        out = []
        for p in repo_path.rglob("*"):
            if not p.is_file():
                continue
            rel = p.relative_to(repo_path).as_posix()
            if ".git/" in rel or rel.startswith(".git/"):
                continue
            if any(fnmatch.fnmatch(rel, g) for g in globs):
                out.append(ChangedFile(path=rel, status="A"))
        return out
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_git_sync.py -v`
Expected: PASS (5 tests). Note: requires `git` installed on test host.

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/ingestion/ homelab-chatbot/backend/tests/unit/test_git_sync.py
git commit -m "[feat:homelab-chatbot] add git sync with clone/pull and change detection"
```

---

### Task 9: Markdown chunker with metadata preservation

**Files:**
- Create: `homelab-chatbot/backend/app/ingestion/markdown.py`
- Create: `homelab-chatbot/backend/tests/unit/test_markdown.py`
- Create: `homelab-chatbot/backend/tests/fixtures/sample.md`

- [ ] **Step 1: Write fixture `homelab-chatbot/backend/tests/fixtures/sample.md`**

```markdown
# Home Lab Documentation

## Networking

Our network uses VLANs for segmentation.

### VLAN 10 - Management

Management VLAN hosts switches and routers.

### VLAN 20 - Services

Services VLAN hosts application servers.

## Storage

TrueNAS Scale runs on the main storage node.
```

- [ ] **Step 2: Write failing test `homelab-chatbot/backend/tests/unit/test_markdown.py`**

```python
from pathlib import Path

from app.ingestion.markdown import Chunk, chunk_markdown_file

FIX = Path(__file__).parent.parent / "fixtures"


def test_chunk_produces_chunks_with_metadata():
    chunks = chunk_markdown_file(
        FIX / "sample.md",
        repo="homelab-docs",
        file_path="docs/sample.md",
        commit_sha="abc123",
    )
    assert len(chunks) > 0
    assert all(isinstance(c, Chunk) for c in chunks)
    for c in chunks:
        assert c.repo == "homelab-docs"
        assert c.file_path == "docs/sample.md"
        assert c.commit_sha == "abc123"
        assert c.line_start >= 1
        assert c.line_end >= c.line_start


def test_chunk_preserves_heading_path():
    chunks = chunk_markdown_file(
        FIX / "sample.md",
        repo="r",
        file_path="s.md",
        commit_sha="x",
    )
    headings = [c.heading_path for c in chunks]
    assert any("Networking" in h for h in headings)
    assert any("VLAN" in h for h in headings)


def test_chunk_text_non_empty():
    chunks = chunk_markdown_file(
        FIX / "sample.md",
        repo="r",
        file_path="s.md",
        commit_sha="x",
    )
    assert all(c.text.strip() for c in chunks)


def test_chunk_id_stable_across_runs(tmp_path: Path):
    (tmp_path / "a.md").write_text("# T\n\nHello world.\n")
    c1 = chunk_markdown_file(tmp_path / "a.md", repo="r", file_path="a.md", commit_sha="x")
    c2 = chunk_markdown_file(tmp_path / "a.md", repo="r", file_path="a.md", commit_sha="x")
    assert [c.id for c in c1] == [c.id for c in c2]
```

- [ ] **Step 3: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 4: Write `homelab-chatbot/backend/app/ingestion/markdown.py`**

```python
"""Chunk Markdown files into embedding-ready pieces with metadata."""

import hashlib
import re
from dataclasses import dataclass
from pathlib import Path

HEADING_RE = re.compile(r"^(#{1,6})\s+(.*?)\s*$")
MAX_CHUNK_CHARS = 2000  # approx 500 tokens
OVERLAP_CHARS = 200


@dataclass
class Chunk:
    id: str
    text: str
    repo: str
    file_path: str
    line_start: int
    line_end: int
    commit_sha: str
    heading_path: str


def _chunk_id(repo: str, file_path: str, line_start: int, text: str) -> str:
    h = hashlib.sha256()
    h.update(repo.encode())
    h.update(b"\0")
    h.update(file_path.encode())
    h.update(b"\0")
    h.update(str(line_start).encode())
    h.update(b"\0")
    h.update(text.encode())
    return h.hexdigest()[:24]


def _walk_sections(lines: list[str]) -> list[tuple[int, int, list[str], str]]:
    """Split lines into (start_line, end_line, body_lines, heading_path) sections.

    Each section is the content under a heading up to the next same-or-higher heading.
    """
    sections: list[tuple[int, int, list[str], str]] = []
    stack: list[tuple[int, str]] = []  # (level, title)
    current_start = 1
    current_lines: list[str] = []

    def heading_path() -> str:
        return " > ".join(title for _, title in stack)

    for idx, line in enumerate(lines, start=1):
        m = HEADING_RE.match(line)
        if m:
            if current_lines:
                sections.append(
                    (current_start, idx - 1, current_lines, heading_path())
                )
            level = len(m.group(1))
            title = m.group(2).strip()
            while stack and stack[-1][0] >= level:
                stack.pop()
            stack.append((level, title))
            current_start = idx
            current_lines = [line]
        else:
            current_lines.append(line)
    if current_lines:
        sections.append(
            (current_start, len(lines), current_lines, heading_path())
        )
    return sections


def _split_long(text: str, max_chars: int, overlap: int) -> list[str]:
    if len(text) <= max_chars:
        return [text]
    parts: list[str] = []
    start = 0
    while start < len(text):
        end = min(len(text), start + max_chars)
        parts.append(text[start:end])
        if end == len(text):
            break
        start = end - overlap
    return parts


def chunk_markdown_file(
    path: Path,
    *,
    repo: str,
    file_path: str,
    commit_sha: str,
) -> list[Chunk]:
    """Return a list of Chunks for the given Markdown file."""
    text = path.read_text(encoding="utf-8")
    lines = text.splitlines(keepends=False)

    chunks: list[Chunk] = []
    for start, end, body_lines, heading_path in _walk_sections(lines):
        body = "\n".join(body_lines).strip()
        if not body:
            continue
        pieces = _split_long(body, MAX_CHUNK_CHARS, OVERLAP_CHARS)
        for piece in pieces:
            chunk_text = piece.strip()
            if not chunk_text:
                continue
            chunks.append(
                Chunk(
                    id=_chunk_id(repo, file_path, start, chunk_text),
                    text=chunk_text,
                    repo=repo,
                    file_path=file_path,
                    line_start=start,
                    line_end=end,
                    commit_sha=commit_sha,
                    heading_path=heading_path,
                )
            )
    return chunks
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_markdown.py -v`
Expected: PASS (4 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/ingestion/markdown.py homelab-chatbot/backend/tests/unit/test_markdown.py homelab-chatbot/backend/tests/fixtures/sample.md
git commit -m "[feat:homelab-chatbot] add markdown chunker with heading-path metadata"
```

---

### Task 10: Embeddings wrapper (sentence-transformers)

**Files:**
- Create: `homelab-chatbot/backend/app/ingestion/embed.py`
- Create: `homelab-chatbot/backend/tests/unit/test_embed.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_embed.py`**

```python
from app.ingestion.embed import Embedder


def test_embedder_returns_384_dims():
    e = Embedder(model_name="BAAI/bge-small-en-v1.5")
    vecs = e.embed_batch(["hello world", "another sentence"])
    assert len(vecs) == 2
    assert len(vecs[0]) == 384
    assert len(vecs[1]) == 384


def test_embedder_deterministic():
    e = Embedder(model_name="BAAI/bge-small-en-v1.5")
    v1 = e.embed_batch(["home lab"])[0]
    v2 = e.embed_batch(["home lab"])[0]
    assert v1 == v2
```

- [ ] **Step 2: Run test — verify fails (model not downloaded yet is fine — should fail on import)**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/ingestion/embed.py`**

```python
"""Wrapper around sentence-transformers for deterministic embeddings."""

from functools import lru_cache

from sentence_transformers import SentenceTransformer

EMBED_DIM = 384


class Embedder:
    """Thin wrapper around SentenceTransformer with batch-embedding convenience."""

    def __init__(self, model_name: str, cache_dir: str | None = None) -> None:
        self._model = _load_model(model_name, cache_dir)

    def embed_batch(self, texts: list[str]) -> list[list[float]]:
        result = self._model.encode(
            texts,
            batch_size=32,
            convert_to_numpy=True,
            normalize_embeddings=True,
            show_progress_bar=False,
        )
        return [vec.tolist() for vec in result]


@lru_cache(maxsize=4)
def _load_model(model_name: str, cache_dir: str | None) -> SentenceTransformer:
    return SentenceTransformer(model_name, cache_folder=cache_dir)
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_embed.py -v`
Expected: PASS (2 tests). First run downloads the model (~130 MB); subsequent runs use the cache.

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/ingestion/embed.py homelab-chatbot/backend/tests/unit/test_embed.py
git commit -m "[feat:homelab-chatbot] add sentence-transformers embedding wrapper"
```

---

### Task 11: LanceDB vector store wrapper

**Files:**
- Create: `homelab-chatbot/backend/app/storage/lance.py`
- Create: `homelab-chatbot/backend/tests/unit/test_lance.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_lance.py`**

```python
from pathlib import Path

import pytest

from app.ingestion.markdown import Chunk
from app.storage.lance import VectorStore


def _chunk(n: int, repo: str = "r", file_path: str = "a.md") -> Chunk:
    return Chunk(
        id=f"id{n}",
        text=f"sample text number {n}",
        repo=repo,
        file_path=file_path,
        line_start=1,
        line_end=1,
        commit_sha="sha",
        heading_path="top",
    )


@pytest.fixture
def store(tmp_path: Path) -> VectorStore:
    return VectorStore(path=tmp_path / "lance")


def test_upsert_and_count(store: VectorStore):
    chunks = [_chunk(i) for i in range(3)]
    vectors = [[float(i)] * 384 for i in range(3)]
    store.upsert(chunks, vectors)
    assert store.count() == 3


def test_delete_by_file_removes_chunks(store: VectorStore):
    a = _chunk(1, file_path="a.md")
    b = _chunk(2, file_path="b.md")
    store.upsert([a, b], [[0.1] * 384, [0.2] * 384])
    store.delete_by_file(repo="r", file_path="a.md")
    assert store.count() == 1


def test_search_returns_top_k(store: VectorStore):
    chunks = [_chunk(i) for i in range(5)]
    vectors = [[float(i)] * 384 for i in range(5)]
    store.upsert(chunks, vectors)
    results = store.search([4.0] * 384, top_k=2)
    assert len(results) == 2
    assert results[0].id == "id4"


def test_search_with_repo_filter(store: VectorStore):
    a = _chunk(1, repo="repo-a")
    b = _chunk(2, repo="repo-b")
    store.upsert([a, b], [[0.1] * 384, [0.9] * 384])
    results = store.search([0.1] * 384, top_k=5, repo="repo-a")
    assert all(r.repo == "repo-a" for r in results)
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/storage/lance.py`**

```python
"""Thin LanceDB wrapper for the markdown_chunks table."""

from dataclasses import dataclass
from pathlib import Path

import lancedb
import pyarrow as pa

from app.ingestion.markdown import Chunk

TABLE_NAME = "markdown_chunks"
EMBED_DIM = 384


@dataclass
class SearchHit:
    id: str
    text: str
    repo: str
    file_path: str
    line_start: int
    line_end: int
    commit_sha: str
    heading_path: str
    score: float


class VectorStore:
    """Wraps a single LanceDB table for markdown chunks."""

    def __init__(self, path: Path) -> None:
        self._db = lancedb.connect(str(path))
        self._schema = pa.schema(
            [
                pa.field("id", pa.string()),
                pa.field("vector", pa.list_(pa.float32(), EMBED_DIM)),
                pa.field("text", pa.string()),
                pa.field("repo", pa.string()),
                pa.field("file_path", pa.string()),
                pa.field("line_start", pa.int32()),
                pa.field("line_end", pa.int32()),
                pa.field("commit_sha", pa.string()),
                pa.field("heading_path", pa.string()),
            ]
        )
        if TABLE_NAME not in self._db.table_names():
            self._db.create_table(TABLE_NAME, schema=self._schema)

    def _table(self) -> lancedb.table.Table:
        return self._db.open_table(TABLE_NAME)

    def upsert(self, chunks: list[Chunk], vectors: list[list[float]]) -> None:
        if len(chunks) != len(vectors):
            raise ValueError("chunks and vectors must have equal length")
        if not chunks:
            return
        rows = [
            {
                "id": c.id,
                "vector": v,
                "text": c.text,
                "repo": c.repo,
                "file_path": c.file_path,
                "line_start": c.line_start,
                "line_end": c.line_end,
                "commit_sha": c.commit_sha,
                "heading_path": c.heading_path,
            }
            for c, v in zip(chunks, vectors, strict=True)
        ]
        ids = [c.id for c in chunks]
        tbl = self._table()
        id_list = ", ".join(f"'{i}'" for i in ids)
        tbl.delete(f"id IN ({id_list})")
        tbl.add(rows)

    def delete_by_file(self, repo: str, file_path: str) -> None:
        tbl = self._table()
        tbl.delete(f"repo = '{repo}' AND file_path = '{file_path}'")

    def count(self) -> int:
        return self._table().count_rows()

    def search(
        self, query_vector: list[float], top_k: int = 5, repo: str | None = None
    ) -> list[SearchHit]:
        tbl = self._table()
        q = tbl.search(query_vector).limit(top_k)
        if repo:
            q = q.where(f"repo = '{repo}'")
        results = q.to_list()
        return [
            SearchHit(
                id=r["id"],
                text=r["text"],
                repo=r["repo"],
                file_path=r["file_path"],
                line_start=r["line_start"],
                line_end=r["line_end"],
                commit_sha=r["commit_sha"],
                heading_path=r["heading_path"],
                score=float(r.get("_distance", 0.0)),
            )
            for r in results
        ]
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_lance.py -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/storage/lance.py homelab-chatbot/backend/tests/unit/test_lance.py
git commit -m "[feat:homelab-chatbot] add LanceDB vector store wrapper"
```

---

### Task 12: Excel → SQLite loader

**Files:**
- Create: `homelab-chatbot/backend/app/ingestion/excel.py`
- Create: `homelab-chatbot/backend/tests/unit/test_excel.py`
- Create: `homelab-chatbot/backend/tests/fixtures/inventory.xlsx` (generated by a helper script during test)

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_excel.py`**

```python
import sqlite3
from pathlib import Path

import pandas as pd
import pytest

from app.ingestion.excel import ExcelLoader


@pytest.fixture
def xlsx(tmp_path: Path) -> Path:
    path = tmp_path / "inventory.xlsx"
    with pd.ExcelWriter(path) as w:
        pd.DataFrame(
            [
                {"Device Name": "nas-01", "OS Name": "Debian", "RAM GB": 32},
                {"Device Name": "router-01", "OS Name": "OpenWrt", "RAM GB": 1},
            ]
        ).to_excel(w, sheet_name="Devices", index=False)
        pd.DataFrame(
            [{"Service": "plex", "Port": 32400}]
        ).to_excel(w, sheet_name="Services", index=False)
    return path


def test_loader_creates_snake_case_tables(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    conn = sqlite3.connect(db_path)
    tables = [
        r[0]
        for r in conn.execute(
            "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
        )
    ]
    assert "devices" in tables
    assert "services" in tables


def test_loader_normalizes_column_names(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    conn = sqlite3.connect(db_path)
    cols = [r[1] for r in conn.execute("PRAGMA table_info(devices)")]
    assert "device_name" in cols
    assert "os_name" in cols
    assert "ram_gb" in cols


def test_loader_rewrites_on_reload(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    loader.load(xlsx)  # second call must not duplicate
    conn = sqlite3.connect(db_path)
    count = conn.execute("SELECT COUNT(*) FROM devices").fetchone()[0]
    assert count == 2


def test_loader_records_meta(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    conn = sqlite3.connect(db_path)
    rows = conn.execute(
        "SELECT sheet, table_name FROM _kb_meta ORDER BY sheet"
    ).fetchall()
    assert ("Devices", "devices") in rows
    assert ("Services", "services") in rows
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/ingestion/excel.py`**

```python
"""Load Excel sheets into a SQLite database as per-sheet tables."""

import hashlib
import re
import sqlite3
from datetime import datetime, timezone
from pathlib import Path

import pandas as pd

META_TABLE = "_kb_meta"


def normalize_snake_case(name: str) -> str:
    s = name.strip().lower()
    s = re.sub(r"[^\w]+", "_", s)
    s = re.sub(r"_+", "_", s).strip("_")
    if not s:
        s = "col"
    if s[0].isdigit():
        s = "_" + s
    return s


class ExcelLoader:
    """Rebuilds SQLite tables from an Excel workbook."""

    def __init__(self, db_path: Path) -> None:
        self._db_path = Path(db_path)

    def _source_hash(self, path: Path) -> str:
        h = hashlib.sha256()
        h.update(path.read_bytes())
        return h.hexdigest()

    def load(self, xlsx_path: Path) -> dict[str, str]:
        """Replace tables and return a {sheet -> table_name} map."""
        xlsx = pd.ExcelFile(xlsx_path)
        sheet_map: dict[str, str] = {}

        with sqlite3.connect(self._db_path) as conn:
            conn.execute(
                f"CREATE TABLE IF NOT EXISTS {META_TABLE} "
                "(sheet TEXT PRIMARY KEY, table_name TEXT, columns_json TEXT, "
                "source_hash TEXT, last_rebuilt_at TEXT)"
            )

            for sheet in xlsx.sheet_names:
                df = xlsx.parse(sheet)
                df.columns = [normalize_snake_case(c) for c in df.columns]
                table = normalize_snake_case(sheet)
                df.to_sql(table, conn, if_exists="replace", index=False)
                sheet_map[sheet] = table
                conn.execute(
                    f"INSERT OR REPLACE INTO {META_TABLE} "
                    "(sheet, table_name, columns_json, source_hash, last_rebuilt_at) "
                    "VALUES (?, ?, ?, ?, ?)",
                    (
                        sheet,
                        table,
                        ",".join(df.columns),
                        self._source_hash(xlsx_path),
                        datetime.now(timezone.utc).isoformat(),
                    ),
                )
            conn.commit()
        return sheet_map
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_excel.py -v`
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/ingestion/excel.py homelab-chatbot/backend/tests/unit/test_excel.py
git commit -m "[feat:homelab-chatbot] add Excel to SQLite loader with column normalization"
```

---

### Task 13: Ingestion orchestrator + scheduler

**Files:**
- Create: `homelab-chatbot/backend/app/ingestion/orchestrator.py`
- Create: `homelab-chatbot/backend/app/ingestion/scheduler.py`
- Create: `homelab-chatbot/backend/tests/unit/test_orchestrator.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_orchestrator.py`**

```python
import subprocess
from pathlib import Path

import pytest

from app.config import AppConfig, EmbeddingsConfig, LLMConfig, OllamaConfig, PathConfig, RepoConfig, RetrievalConfig, SyncConfig
from app.ingestion.orchestrator import IngestionOrchestrator


def _remote(tmp: Path) -> Path:
    remote = tmp / "remote.git"
    work = tmp / "work"
    work.mkdir()
    subprocess.run(["git", "init", "--bare", str(remote)], check=True)
    subprocess.run(["git", "init", str(work)], check=True)
    subprocess.run(["git", "-C", str(work), "config", "user.email", "t@t"], check=True)
    subprocess.run(["git", "-C", str(work), "config", "user.name", "t"], check=True)
    subprocess.run(["git", "-C", str(work), "config", "commit.gpgsign", "false"], check=True)
    (work / "notes.md").write_text("# Notes\n\nHome lab has a NAS.\n")
    subprocess.run(["git", "-C", str(work), "add", "."], check=True)
    subprocess.run(["git", "-C", str(work), "commit", "-m", "i"], check=True)
    subprocess.run(["git", "-C", str(work), "branch", "-M", "main"], check=True)
    subprocess.run(["git", "-C", str(work), "remote", "add", "origin", str(remote)], check=True)
    subprocess.run(["git", "-C", str(work), "push", "-u", "origin", "main"], check=True)
    return remote


@pytest.fixture
def cfg(tmp_path: Path) -> AppConfig:
    remote = _remote(tmp_path)
    return AppConfig(
        sync=SyncConfig(interval_seconds=60, state_file=str(tmp_path / "state.json")),
        repos=[
            RepoConfig(
                name="r",
                url=f"file://{remote}",
                branch="main",
                token_env="UNUSED",
                include_globs=["**/*.md"],
            )
        ],
        embeddings=EmbeddingsConfig(
            model="BAAI/bge-small-en-v1.5", cache_dir=str(tmp_path / "models")
        ),
        vector_store=PathConfig(path=str(tmp_path / "lance")),
        chat_db=PathConfig(path=str(tmp_path / "chat.db")),
        kb_db=PathConfig(path=str(tmp_path / "kb.db")),
        llm=LLMConfig(
            default_provider="anthropic",
            default_model="m",
            ollama=OllamaConfig(host="http://x", tool_capable_models=[]),
        ),
        retrieval=RetrievalConfig(top_k=5, memory_turns=10),
    )


def test_full_ingest_populates_vector_store(tmp_path: Path, cfg: AppConfig):
    orch = IngestionOrchestrator(
        config=cfg,
        clone_root=tmp_path / "repos",
        get_token=lambda _: None,
    )
    orch.run_once()
    assert orch.vector_store.count() > 0


def test_second_run_no_changes(tmp_path: Path, cfg: AppConfig):
    orch = IngestionOrchestrator(
        config=cfg,
        clone_root=tmp_path / "repos",
        get_token=lambda _: None,
    )
    orch.run_once()
    count_1 = orch.vector_store.count()
    orch.run_once()
    count_2 = orch.vector_store.count()
    assert count_1 == count_2
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/ingestion/orchestrator.py`**

```python
"""Coordinate git sync + markdown chunking + Excel loading + vector index updates."""

from pathlib import Path
from typing import Callable

from app.config import AppConfig, RepoConfig
from app.ingestion.embed import Embedder
from app.ingestion.excel import ExcelLoader
from app.ingestion.git_sync import ChangedFile, GitSync, SyncResult
from app.ingestion.markdown import chunk_markdown_file
from app.storage.lance import VectorStore


class IngestionOrchestrator:
    """Runs one full ingestion cycle across all configured repos."""

    def __init__(
        self,
        config: AppConfig,
        clone_root: Path,
        get_token: Callable[[str], str | None],
    ) -> None:
        self._config = config
        self._git = GitSync(clone_root=clone_root, get_token=get_token)
        self._embedder = Embedder(
            model_name=config.embeddings.model, cache_dir=config.embeddings.cache_dir
        )
        self.vector_store = VectorStore(Path(config.vector_store.path))
        self._excel = ExcelLoader(Path(config.kb_db.path))

    def run_once(self) -> list[SyncResult]:
        results = []
        for repo_cfg in self._config.repos:
            result = self._git.sync(repo_cfg)
            self._apply(repo_cfg, result)
            results.append(result)
        return results

    def _apply(self, repo_cfg: RepoConfig, result: SyncResult) -> None:
        repo_path = self._git.repo_path(repo_cfg)
        if result.cloned:
            files = result.matched_files
        else:
            files = result.matched_files or []

        for cf in files:
            self._apply_file(repo_cfg, repo_path, cf, result.new_sha)

    def _apply_file(
        self,
        repo_cfg: RepoConfig,
        repo_path: Path,
        cf: ChangedFile,
        new_sha: str,
    ) -> None:
        file_path = repo_path / cf.path

        if cf.path.endswith(".md"):
            if cf.status == "D" or not file_path.exists():
                self.vector_store.delete_by_file(repo_cfg.name, cf.path)
                return
            chunks = chunk_markdown_file(
                file_path,
                repo=repo_cfg.name,
                file_path=cf.path,
                commit_sha=new_sha,
            )
            self.vector_store.delete_by_file(repo_cfg.name, cf.path)
            if chunks:
                vectors = self._embedder.embed_batch([c.text for c in chunks])
                self.vector_store.upsert(chunks, vectors)

        elif cf.path.endswith(".xlsx"):
            if cf.status == "D" or not file_path.exists():
                return
            self._excel.load(file_path)
```

- [ ] **Step 4: Write `homelab-chatbot/backend/app/ingestion/scheduler.py`**

```python
"""Background scheduler that runs the orchestrator at a fixed interval."""

import logging
from pathlib import Path
from typing import Callable

from apscheduler.schedulers.asyncio import AsyncIOScheduler

from app.config import AppConfig
from app.ingestion.orchestrator import IngestionOrchestrator

logger = logging.getLogger(__name__)


class SyncScheduler:
    """Schedules periodic runs of the IngestionOrchestrator."""

    def __init__(
        self,
        config: AppConfig,
        clone_root: Path,
        get_token: Callable[[str], str | None],
    ) -> None:
        self.orchestrator = IngestionOrchestrator(config, clone_root, get_token)
        self._interval = config.sync.interval_seconds
        self._scheduler = AsyncIOScheduler()

    def start(self) -> None:
        self._scheduler.add_job(
            self._safe_run,
            trigger="interval",
            seconds=self._interval,
            id="ingest",
            max_instances=1,
            coalesce=True,
            next_run_time=None,
        )
        self._scheduler.start()
        try:
            self.orchestrator.run_once()
        except Exception as e:  # noqa: BLE001
            logger.exception("initial ingest failed: %s", e)

    def _safe_run(self) -> None:
        try:
            self.orchestrator.run_once()
        except Exception as e:  # noqa: BLE001
            logger.exception("scheduled ingest failed: %s", e)

    def shutdown(self) -> None:
        self._scheduler.shutdown(wait=False)
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_orchestrator.py -v`
Expected: PASS (2 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/ingestion/orchestrator.py homelab-chatbot/backend/app/ingestion/scheduler.py homelab-chatbot/backend/tests/unit/test_orchestrator.py
git commit -m "[feat:homelab-chatbot] add ingestion orchestrator and APScheduler wiring"
```

---

## Phase 3 — Retrieval

### Task 14: Vector retrieval tool

**Files:**
- Create: `homelab-chatbot/backend/app/retrieval/__init__.py`
- Create: `homelab-chatbot/backend/app/retrieval/vector_tool.py`
- Create: `homelab-chatbot/backend/tests/unit/test_vector_tool.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/app/retrieval/__init__.py`** (empty).

- [ ] **Step 2: Write failing test `homelab-chatbot/backend/tests/unit/test_vector_tool.py`**

```python
from pathlib import Path

import pytest

from app.ingestion.embed import Embedder
from app.ingestion.markdown import Chunk
from app.retrieval.vector_tool import VectorSearchTool
from app.storage.lance import VectorStore


@pytest.fixture
def tool(tmp_path: Path) -> VectorSearchTool:
    store = VectorStore(tmp_path / "lance")
    embedder = Embedder(model_name="BAAI/bge-small-en-v1.5")
    chunks = [
        Chunk(id="c1", text="NAS runs TrueNAS Scale on Debian",
              repo="docs", file_path="storage.md",
              line_start=1, line_end=2, commit_sha="s", heading_path="Storage"),
        Chunk(id="c2", text="VLAN 10 is the management network",
              repo="docs", file_path="net.md",
              line_start=1, line_end=2, commit_sha="s", heading_path="Net > VLAN"),
    ]
    store.upsert(chunks, embedder.embed_batch([c.text for c in chunks]))
    return VectorSearchTool(store=store, embedder=embedder, top_k=2)


def test_search_returns_relevant_hits(tool: VectorSearchTool):
    hits = tool.search("tell me about the storage NAS")
    assert len(hits) > 0
    assert hits[0].file_path == "storage.md"


def test_search_respects_top_k(tool: VectorSearchTool):
    hits = tool.search("anything", top_k=1)
    assert len(hits) == 1


def test_search_with_repo_filter(tool: VectorSearchTool):
    hits = tool.search("vlan", repo="nonexistent")
    assert hits == []
```

- [ ] **Step 3: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 4: Write `homelab-chatbot/backend/app/retrieval/vector_tool.py`**

```python
"""Vector-search tool exposed to the LLM agent."""

from app.ingestion.embed import Embedder
from app.storage.lance import SearchHit, VectorStore


class VectorSearchTool:
    """Embed a query and return top-K matching markdown chunks."""

    TOOL_NAME = "search_homelab_docs"

    def __init__(
        self,
        store: VectorStore,
        embedder: Embedder,
        top_k: int = 5,
    ) -> None:
        self._store = store
        self._embedder = embedder
        self._default_top_k = top_k

    def search(
        self,
        query: str,
        top_k: int | None = None,
        repo: str | None = None,
    ) -> list[SearchHit]:
        vec = self._embedder.embed_batch([query])[0]
        return self._store.search(
            query_vector=vec, top_k=top_k or self._default_top_k, repo=repo
        )

    def as_llama_tool(self):
        """Return a LlamaIndex FunctionTool wrapping this tool."""
        from llama_index.core.tools import FunctionTool

        def _run(query: str, repo: str | None = None) -> list[dict]:
            """Search the home lab documentation for information."""
            return [
                {
                    "text": hit.text,
                    "repo": hit.repo,
                    "file_path": hit.file_path,
                    "heading_path": hit.heading_path,
                    "line_start": hit.line_start,
                    "line_end": hit.line_end,
                }
                for hit in self.search(query, repo=repo)
            ]

        return FunctionTool.from_defaults(
            fn=_run,
            name=self.TOOL_NAME,
            description=(
                "Search the home lab documentation by semantic similarity. "
                "Use for prose/conceptual questions. Optional `repo` filter narrows to a specific repo."
            ),
        )
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_vector_tool.py -v`
Expected: PASS (3 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/retrieval/ homelab-chatbot/backend/tests/unit/test_vector_tool.py
git commit -m "[feat:homelab-chatbot] add vector search tool wrapping LanceDB"
```

---

### Task 15: NL→SQL tool over kb.db (read-only)

**Files:**
- Create: `homelab-chatbot/backend/app/retrieval/sql_tool.py`
- Create: `homelab-chatbot/backend/tests/unit/test_sql_tool.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_sql_tool.py`**

```python
import sqlite3
from pathlib import Path

import pandas as pd
import pytest

from app.retrieval.sql_tool import SQLTool


@pytest.fixture
def kb_db(tmp_path: Path) -> Path:
    path = tmp_path / "kb.db"
    with sqlite3.connect(path) as conn:
        pd.DataFrame(
            [
                {"device_name": "nas-01", "os_name": "Debian", "ram_gb": 32},
                {"device_name": "router-01", "os_name": "OpenWrt", "ram_gb": 1},
                {"device_name": "server-01", "os_name": "Debian", "ram_gb": 64},
            ]
        ).to_sql("devices", conn, index=False)
    return path


def test_run_select_returns_rows(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    result = tool.run_select("SELECT device_name FROM devices WHERE os_name = 'Debian'")
    names = [r["device_name"] for r in result.rows]
    assert set(names) == {"nas-01", "server-01"}


def test_run_select_blocks_writes(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    with pytest.raises(Exception):
        tool.run_select("DELETE FROM devices")


def test_aggregate_count(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    result = tool.run_select(
        "SELECT COUNT(*) AS n FROM devices WHERE os_name = 'Debian'"
    )
    assert result.rows[0]["n"] == 2


def test_schema_summary_lists_tables(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    summary = tool.schema_summary()
    assert "devices" in summary
    assert "device_name" in summary


def test_reject_non_select_statement(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    with pytest.raises(ValueError):
        tool.run_select("UPDATE devices SET os_name = 'Arch'")
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/retrieval/sql_tool.py`**

```python
"""Read-only SQL tool over kb.db (Excel-sourced tables)."""

import re
import sqlite3
from dataclasses import dataclass
from pathlib import Path


@dataclass
class QueryResult:
    sql: str
    rows: list[dict]


class SQLTool:
    """Executes read-only SELECT statements against kb.db."""

    TOOL_NAME = "query_homelab_inventory"

    def __init__(self, db_path: Path) -> None:
        self._db_path = Path(db_path)

    def _open(self) -> sqlite3.Connection:
        conn = sqlite3.connect(f"file:{self._db_path}?mode=ro", uri=True)
        conn.row_factory = sqlite3.Row
        conn.execute("PRAGMA query_only = ON")
        return conn

    def run_select(self, sql: str) -> QueryResult:
        if not self._is_select(sql):
            raise ValueError(f"only SELECT statements allowed; got: {sql[:80]}")
        with self._open() as conn:
            cur = conn.execute(sql)
            rows = [dict(r) for r in cur.fetchall()]
        return QueryResult(sql=sql, rows=rows)

    def _is_select(self, sql: str) -> bool:
        cleaned = re.sub(r"--[^\n]*", "", sql).strip().rstrip(";").strip()
        if ";" in cleaned:
            return False
        first = cleaned.split(None, 1)[0].lower() if cleaned else ""
        return first in ("select", "with")

    def schema_summary(self) -> str:
        with self._open() as conn:
            tables = [
                r[0]
                for r in conn.execute(
                    "SELECT name FROM sqlite_master WHERE type='table' "
                    "AND name NOT LIKE 'sqlite_%' AND name != '_kb_meta'"
                )
            ]
            lines = []
            for t in tables:
                cols = [r[1] for r in conn.execute(f"PRAGMA table_info({t})")]
                lines.append(f"{t}({', '.join(cols)})")
            return "\n".join(lines)

    def as_llama_tool(self, llm):
        """Return a LlamaIndex FunctionTool that uses `llm` to generate SQL."""
        from llama_index.core.tools import FunctionTool

        schema = self.schema_summary()

        async def _run(question: str) -> dict:
            """Answer an inventory question by generating and executing a SELECT."""
            prompt = (
                "You are a SQLite expert. Produce a single SELECT statement "
                "answering the user's question. Use only these tables:\n\n"
                f"{schema}\n\n"
                "Return ONLY the SQL, no prose, no fenced block.\n\n"
                f"Question: {question}"
            )
            response = await llm.acomplete(prompt)
            sql = str(response).strip().strip("`")
            if sql.lower().startswith("sql"):
                sql = sql[3:].strip()
            try:
                result = self.run_select(sql)
            except Exception as e:  # noqa: BLE001
                return {"sql": sql, "error": str(e), "rows": []}
            return {"sql": result.sql, "rows": result.rows}

        return FunctionTool.from_defaults(
            async_fn=_run,
            name=self.TOOL_NAME,
            description=(
                "Answer questions about the homelab inventory (devices, services, etc.) "
                "using structured SQL queries. Use for counting, filtering, aggregating "
                "or looking up specific items in tabular data."
            ),
        )
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_sql_tool.py -v`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/retrieval/sql_tool.py homelab-chatbot/backend/tests/unit/test_sql_tool.py
git commit -m "[feat:homelab-chatbot] add read-only NL-to-SQL tool with SELECT-only guard"
```

---

### Task 16: Retrieval golden-set tests

**Files:**
- Create: `homelab-chatbot/backend/tests/retrieval/__init__.py`
- Create: `homelab-chatbot/backend/tests/retrieval/test_golden.py`
- Create: `homelab-chatbot/backend/tests/retrieval/fixtures/docs/networking.md`
- Create: `homelab-chatbot/backend/tests/retrieval/fixtures/docs/storage.md`
- Create: `homelab-chatbot/backend/tests/retrieval/fixtures/docs/services.md`

- [ ] **Step 1: Write fixture markdown files**

`tests/retrieval/fixtures/docs/networking.md`:
```markdown
# Networking

## VLAN Layout

- VLAN 10: Management (switches, routers, IPMI)
- VLAN 20: Services (app servers, containers)
- VLAN 30: IoT (sensors, cameras)

## Firewall

OPNsense runs on a mini PC. Rules block inter-VLAN traffic by default.
```

`tests/retrieval/fixtures/docs/storage.md`:
```markdown
# Storage

## TrueNAS

Primary storage is a TrueNAS Scale box with 6 x 8TB drives in RAIDZ2.
SMB shares are exposed to the LAN; NFS to the services VLAN.

## Backup

Restic backs up to Backblaze B2 nightly.
```

`tests/retrieval/fixtures/docs/services.md`:
```markdown
# Services

## Plex

Plex runs on the media server, port 32400. Transcoding uses a GTX 1650.

## Home Assistant

Home Assistant supervises 30+ ESPHome devices on the IoT VLAN.
```

- [ ] **Step 2: Write `homelab-chatbot/backend/tests/retrieval/__init__.py`** (empty).

- [ ] **Step 3: Write golden-set test `homelab-chatbot/backend/tests/retrieval/test_golden.py`**

```python
from pathlib import Path

import pytest

from app.ingestion.embed import Embedder
from app.ingestion.markdown import chunk_markdown_file
from app.retrieval.vector_tool import VectorSearchTool
from app.storage.lance import VectorStore

FIX = Path(__file__).parent / "fixtures" / "docs"

GOLDEN_SET = [
    ("how is storage organized?", "storage.md"),
    ("what does RAIDZ2 provide?", "storage.md"),
    ("what firewall do we run?", "networking.md"),
    ("what's on VLAN 10?", "networking.md"),
    ("where does Plex run?", "services.md"),
    ("what port is Plex?", "services.md"),
    ("how many drives in the NAS?", "storage.md"),
    ("does Home Assistant talk to ESPHome?", "services.md"),
    ("where do backups go?", "storage.md"),
    ("what is the IoT VLAN used for?", "networking.md"),
    ("does the firewall block inter-VLAN traffic?", "networking.md"),
    ("is there a GPU for transcoding?", "services.md"),
]


@pytest.fixture(scope="module")
def tool(tmp_path_factory) -> VectorSearchTool:
    tmp = tmp_path_factory.mktemp("retrieval-golden")
    store = VectorStore(tmp / "lance")
    embedder = Embedder(model_name="BAAI/bge-small-en-v1.5")
    all_chunks = []
    for md in FIX.glob("*.md"):
        chunks = chunk_markdown_file(
            md,
            repo="docs",
            file_path=md.name,
            commit_sha="golden",
        )
        all_chunks.extend(chunks)
    vectors = embedder.embed_batch([c.text for c in all_chunks])
    store.upsert(all_chunks, vectors)
    return VectorSearchTool(store=store, embedder=embedder, top_k=3)


@pytest.mark.parametrize("query,expected_file", GOLDEN_SET)
def test_retrieval_hits_expected_file(
    tool: VectorSearchTool, query: str, expected_file: str
):
    hits = tool.search(query)
    assert hits, f"no hits for: {query}"
    top_files = [h.file_path for h in hits]
    assert expected_file in top_files, (
        f"expected {expected_file} in top-K for {query!r}, got {top_files}"
    )
```

- [ ] **Step 4: Run golden-set tests**

Run: `uv run pytest tests/retrieval -v`
Expected: PASS (12 parametrized tests). Total runtime 10–30s (includes loading the embedding model once).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/tests/retrieval/
git commit -m "[test:homelab-chatbot] add retrieval golden-set tests"
```

---

## Phase 4 — LLM & Agent

### Task 17: LLM provider factory

**Files:**
- Create: `homelab-chatbot/backend/app/llm/__init__.py`
- Create: `homelab-chatbot/backend/app/llm/provider.py`
- Create: `homelab-chatbot/backend/tests/unit/test_provider.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/app/llm/__init__.py`** (empty).

- [ ] **Step 2: Write failing test `homelab-chatbot/backend/tests/unit/test_provider.py`**

```python
import pytest

from app.llm.provider import build_llm, UnknownProviderError


def test_build_ollama(monkeypatch):
    llm = build_llm(
        provider="ollama",
        model="llama3.1",
        ollama_host="http://localhost:11434",
    )
    from llama_index.llms.ollama import Ollama

    assert isinstance(llm, Ollama)


def test_build_anthropic(monkeypatch):
    llm = build_llm(provider="anthropic", model="claude-sonnet-4-6", api_key="k")
    from llama_index.llms.anthropic import Anthropic

    assert isinstance(llm, Anthropic)


def test_build_google(monkeypatch):
    llm = build_llm(provider="google", model="gemini-2.5-flash", api_key="k")
    from llama_index.llms.google_genai import GoogleGenAI

    assert isinstance(llm, GoogleGenAI)


def test_unknown_provider_raises():
    with pytest.raises(UnknownProviderError):
        build_llm(provider="cohere", model="x")


def test_anthropic_missing_key_raises():
    with pytest.raises(ValueError):
        build_llm(provider="anthropic", model="m", api_key=None)
```

- [ ] **Step 3: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 4: Write `homelab-chatbot/backend/app/llm/provider.py`**

```python
"""Factory for LlamaIndex LLM instances across Ollama / Anthropic / Google."""

from typing import Literal

from llama_index.core.llms import LLM

Provider = Literal["anthropic", "google", "ollama"]


class UnknownProviderError(ValueError):
    pass


def build_llm(
    *,
    provider: Provider | str,
    model: str,
    api_key: str | None = None,
    ollama_host: str | None = None,
    request_timeout: int = 120,
) -> LLM:
    """Return a LlamaIndex LLM instance for the named provider/model."""
    if provider == "ollama":
        from llama_index.llms.ollama import Ollama

        return Ollama(
            model=model,
            base_url=ollama_host or "http://localhost:11434",
            request_timeout=request_timeout,
        )
    if provider == "anthropic":
        if not api_key:
            raise ValueError("ANTHROPIC_API_KEY required for Anthropic provider")
        from llama_index.llms.anthropic import Anthropic

        return Anthropic(model=model, api_key=api_key)
    if provider == "google":
        if not api_key:
            raise ValueError("GOOGLE_API_KEY required for Google provider")
        from llama_index.llms.google_genai import GoogleGenAI

        return GoogleGenAI(model=model, api_key=api_key)
    raise UnknownProviderError(f"unknown provider: {provider}")
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_provider.py -v`
Expected: PASS (5 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/llm/ homelab-chatbot/backend/tests/unit/test_provider.py
git commit -m "[feat:homelab-chatbot] add LLM provider factory for three backends"
```

---

### Task 18: Model registry with Ollama live fetch

**Files:**
- Create: `homelab-chatbot/backend/app/llm/registry.py`
- Create: `homelab-chatbot/backend/tests/unit/test_registry.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_registry.py`**

```python
import httpx
import pytest

from app.llm.registry import ModelRegistry


def test_static_models_for_anthropic_and_google():
    registry = ModelRegistry(ollama_host="http://localhost:11434")
    anth = registry.static_models("anthropic")
    goog = registry.static_models("google")
    assert any(m.startswith("claude") for m in anth)
    assert any(m.startswith("gemini") for m in goog)


async def test_ollama_models_fetched_from_api(respx_mock):
    respx_mock.get("http://mock-ollama:11434/api/tags").mock(
        return_value=httpx.Response(
            200,
            json={"models": [{"name": "llama3.1:latest"}, {"name": "qwen2.5:7b"}]},
        )
    )
    registry = ModelRegistry(ollama_host="http://mock-ollama:11434")
    models = await registry.ollama_models()
    assert "llama3.1:latest" in models
    assert "qwen2.5:7b" in models


async def test_ollama_unreachable_returns_empty(respx_mock):
    respx_mock.get("http://mock-ollama:11434/api/tags").mock(
        side_effect=httpx.ConnectError("boom")
    )
    registry = ModelRegistry(ollama_host="http://mock-ollama:11434")
    models = await registry.ollama_models()
    assert models == []


def test_static_models_unknown_provider_returns_empty():
    registry = ModelRegistry(ollama_host="http://x")
    assert registry.static_models("cohere") == []
```

- [ ] **Step 2: Add `respx` to dev dependencies**

In `homelab-chatbot/backend/pyproject.toml`, under `[dependency-groups] dev`, add `respx>=0.22`. Then run `uv sync`.

- [ ] **Step 3: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 4: Write `homelab-chatbot/backend/app/llm/registry.py`**

```python
"""Per-provider model inventory."""

import httpx

ANTHROPIC_MODELS = [
    "claude-opus-4-7",
    "claude-sonnet-4-6",
    "claude-haiku-4-5",
]
GOOGLE_MODELS = [
    "gemini-2.5-pro",
    "gemini-2.5-flash",
]


class ModelRegistry:
    """Returns available models for each provider."""

    def __init__(self, ollama_host: str) -> None:
        self._ollama_host = ollama_host.rstrip("/")

    def static_models(self, provider: str) -> list[str]:
        if provider == "anthropic":
            return list(ANTHROPIC_MODELS)
        if provider == "google":
            return list(GOOGLE_MODELS)
        return []

    async def ollama_models(self) -> list[str]:
        url = f"{self._ollama_host}/api/tags"
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                resp = await client.get(url)
                resp.raise_for_status()
                data = resp.json()
        except (httpx.HTTPError, ValueError):
            return []
        return [m["name"] for m in data.get("models", [])]
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_registry.py -v`
Expected: PASS (4 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/llm/registry.py homelab-chatbot/backend/tests/unit/test_registry.py homelab-chatbot/backend/pyproject.toml homelab-chatbot/backend/uv.lock
git commit -m "[feat:homelab-chatbot] add model registry with live Ollama model discovery"
```

---

### Task 19: Tool-capability gate

**Files:**
- Create: `homelab-chatbot/backend/app/llm/capability.py`
- Create: `homelab-chatbot/backend/tests/unit/test_capability.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_capability.py`**

```python
from app.llm.capability import supports_tools


def test_anthropic_always_supports():
    assert supports_tools(provider="anthropic", model="claude-sonnet-4-6", ollama_whitelist=[])


def test_google_always_supports():
    assert supports_tools(provider="google", model="gemini-2.5-pro", ollama_whitelist=[])


def test_ollama_in_whitelist_matches_exact():
    assert supports_tools(
        provider="ollama", model="llama3.1", ollama_whitelist=["llama3.1", "qwen2.5"]
    )


def test_ollama_in_whitelist_matches_tagged_variant():
    assert supports_tools(
        provider="ollama",
        model="llama3.1:8b-instruct",
        ollama_whitelist=["llama3.1", "qwen2.5"],
    )


def test_ollama_not_in_whitelist():
    assert not supports_tools(
        provider="ollama", model="phi3", ollama_whitelist=["llama3.1"]
    )
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/llm/capability.py`**

```python
"""Determine whether a given provider/model supports native tool-calling."""


def _base_name(model: str) -> str:
    return model.split(":", 1)[0]


def supports_tools(
    *, provider: str, model: str, ollama_whitelist: list[str]
) -> bool:
    """Return True if the model can reliably use LlamaIndex tool-calling."""
    if provider in ("anthropic", "google"):
        return True
    if provider == "ollama":
        base = _base_name(model)
        return base in ollama_whitelist
    return False
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_capability.py -v`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/llm/capability.py homelab-chatbot/backend/tests/unit/test_capability.py
git commit -m "[feat:homelab-chatbot] add tool-capability gate for Ollama models"
```

---

### Task 20: Agent + SSE streaming adapter

**Files:**
- Create: `homelab-chatbot/backend/app/models/__init__.py`
- Create: `homelab-chatbot/backend/app/models/chat.py`
- Create: `homelab-chatbot/backend/app/llm/agent.py`
- Create: `homelab-chatbot/backend/app/llm/sse.py`
- Create: `homelab-chatbot/backend/tests/unit/test_agent.py`
- Create: `homelab-chatbot/backend/tests/unit/test_sse.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/app/models/__init__.py`** (empty).

- [ ] **Step 2: Write `homelab-chatbot/backend/app/models/chat.py`**

```python
"""Pydantic request/response schemas for the chat API."""

from typing import Literal

from pydantic import BaseModel


class ChatRequest(BaseModel):
    conv_id: str | None = None
    message: str
    provider: str
    model: str


class ConversationOut(BaseModel):
    id: str
    title: str
    provider: str
    model: str
    created_at: str
    updated_at: str


class MessageOut(BaseModel):
    id: int
    role: Literal["user", "assistant", "tool"]
    content: str
    tool_name: str | None = None
    tool_calls: str | None = None
    partial: bool = False
    created_at: str
```

- [ ] **Step 3: Write `homelab-chatbot/backend/app/llm/sse.py`**

```python
"""Translate LlamaIndex agent events into Vercel AI SDK `data-stream` SSE frames."""

import json
from dataclasses import dataclass
from typing import AsyncIterator


@dataclass
class AgentEvent:
    kind: str  # 'text-delta' | 'tool-call' | 'tool-result' | 'error' | 'done'
    data: dict


def format_sse(event: AgentEvent) -> bytes:
    lines = [f"event: {event.kind}", f"data: {json.dumps(event.data)}", "", ""]
    return ("\n".join(lines)).encode("utf-8")


async def to_sse_stream(events: AsyncIterator[AgentEvent]) -> AsyncIterator[bytes]:
    async for event in events:
        yield format_sse(event)
```

- [ ] **Step 4: Write `homelab-chatbot/backend/app/llm/agent.py`**

```python
"""Build a LlamaIndex agent and stream events out as AgentEvents."""

import logging
from typing import AsyncIterator

from llama_index.core.agent import ReActAgent
from llama_index.core.llms import LLM, ChatMessage
from llama_index.core.tools import FunctionTool

from app.llm.sse import AgentEvent

logger = logging.getLogger(__name__)

SYSTEM_PROMPT = (
    "You are a helpful assistant for a home lab knowledge base. "
    "You have two tools available:\n"
    "- search_homelab_docs: for prose/conceptual questions about the docs\n"
    "- query_homelab_inventory: for structured questions about devices, services, and inventory\n\n"
    "Call tools when they help answer the user's question. Answer concisely. "
    "If the tools return no useful information, say so."
)


def build_messages(history: list[tuple[str, str]], user_message: str) -> list[ChatMessage]:
    messages: list[ChatMessage] = [ChatMessage(role="system", content=SYSTEM_PROMPT)]
    for role, content in history:
        messages.append(ChatMessage(role=role, content=content))
    messages.append(ChatMessage(role="user", content=user_message))
    return messages


async def run_agent(
    llm: LLM,
    tools: list[FunctionTool],
    history: list[tuple[str, str]],
    user_message: str,
    tools_enabled: bool,
) -> AsyncIterator[AgentEvent]:
    """Stream AgentEvents while the agent produces a response."""
    try:
        if not tools_enabled or not tools:
            async for delta in _stream_chat(llm, history, user_message):
                yield AgentEvent("text-delta", {"text": delta})
            yield AgentEvent("done", {})
            return

        agent = ReActAgent.from_tools(tools=tools, llm=llm, verbose=False)
        response = await agent.astream_chat(user_message)
        async for token in response.async_response_gen():
            yield AgentEvent("text-delta", {"text": token})

        for source in getattr(response, "sources", []) or []:
            yield AgentEvent(
                "tool-result",
                {
                    "name": getattr(source, "tool_name", "?"),
                    "summary": str(source.content)[:500] if source.content else "",
                },
            )
        yield AgentEvent("done", {})
    except Exception as e:  # noqa: BLE001
        logger.exception("agent error")
        yield AgentEvent("error", {"message": str(e)})


async def _stream_chat(
    llm: LLM, history: list[tuple[str, str]], user_message: str
) -> AsyncIterator[str]:
    messages = build_messages(history, user_message)
    stream = await llm.astream_chat(messages)
    async for chunk in stream:
        delta = chunk.delta if hasattr(chunk, "delta") else ""
        if delta:
            yield delta
```

- [ ] **Step 5: Write `homelab-chatbot/backend/tests/unit/test_sse.py`**

```python
import json

from app.llm.sse import AgentEvent, format_sse


def test_format_sse_contains_event_and_data():
    out = format_sse(AgentEvent("text-delta", {"text": "hi"}))
    decoded = out.decode("utf-8")
    assert "event: text-delta" in decoded
    assert "data:" in decoded
    payload_line = [l for l in decoded.splitlines() if l.startswith("data:")][0]
    assert json.loads(payload_line[len("data: "):]) == {"text": "hi"}


def test_done_event_serialization():
    out = format_sse(AgentEvent("done", {}))
    assert b"event: done" in out
    assert b"data: {}" in out
```

- [ ] **Step 6: Write `homelab-chatbot/backend/tests/unit/test_agent.py`**

```python
from typing import AsyncIterator

from llama_index.core.base.llms.types import ChatMessage, ChatResponse

from app.llm.agent import run_agent
from app.llm.sse import AgentEvent


class FakeLLM:
    """Minimal stand-in for LlamaIndex LLM's astream_chat."""

    metadata = None

    async def astream_chat(self, messages) -> AsyncIterator[ChatResponse]:
        text = "Hello world"
        acc = ""
        for ch in text:
            acc += ch
            yield ChatResponse(
                message=ChatMessage(role="assistant", content=acc),
                delta=ch,
                raw={},
            )


async def test_run_agent_without_tools_yields_text_deltas():
    events: list[AgentEvent] = []
    async for event in run_agent(
        llm=FakeLLM(),
        tools=[],
        history=[],
        user_message="hi",
        tools_enabled=False,
    ):
        events.append(event)
    kinds = [e.kind for e in events]
    assert "text-delta" in kinds
    assert kinds[-1] == "done"
    joined = "".join(e.data.get("text", "") for e in events if e.kind == "text-delta")
    assert joined == "Hello world"
```

- [ ] **Step 7: Run tests — verify pass**

Run: `uv run pytest tests/unit/test_sse.py tests/unit/test_agent.py -v`
Expected: PASS (3 tests).

- [ ] **Step 8: Commit**

```bash
git add homelab-chatbot/backend/app/models/ homelab-chatbot/backend/app/llm/ homelab-chatbot/backend/tests/unit/test_agent.py homelab-chatbot/backend/tests/unit/test_sse.py
git commit -m "[feat:homelab-chatbot] add streaming agent runner and SSE envelope adapter"
```

---

## Phase 5 — API Routes & App Assembly

### Task 21: Auth middleware + login route

**Files:**
- Create: `homelab-chatbot/backend/app/deps.py`
- Create: `homelab-chatbot/backend/app/routes/__init__.py`
- Create: `homelab-chatbot/backend/app/routes/auth.py`
- Create: `homelab-chatbot/backend/tests/unit/test_auth_route.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/app/routes/__init__.py`** (empty).

- [ ] **Step 2: Write `homelab-chatbot/backend/app/deps.py`**

```python
"""FastAPI dependencies shared across routes."""

from fastapi import Cookie, HTTPException, Request, status

from app.auth import AuthService

SESSION_COOKIE_NAME = "hlcb_session"


def get_auth_service(request: Request) -> AuthService:
    return request.app.state.auth


def require_session(
    request: Request,
    hlcb_session: str | None = Cookie(default=None),
) -> None:
    auth: AuthService = request.app.state.auth
    if not hlcb_session or not auth.verify_session_token(hlcb_session):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="authentication required"
        )
```

- [ ] **Step 3: Write `homelab-chatbot/backend/app/routes/auth.py`**

```python
"""Login / logout endpoints."""

from fastapi import APIRouter, Depends, HTTPException, Response, status
from pydantic import BaseModel

from app.auth import AuthService
from app.deps import SESSION_COOKIE_NAME, get_auth_service

router = APIRouter(prefix="/api/auth", tags=["auth"])


class LoginRequest(BaseModel):
    password: str


@router.post("/login")
async def login(
    body: LoginRequest,
    response: Response,
    auth: AuthService = Depends(get_auth_service),
) -> dict[str, bool]:
    if not auth.check_password(body.password):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="invalid password"
        )
    token = auth.issue_session_token()
    response.set_cookie(
        key=SESSION_COOKIE_NAME,
        value=token,
        httponly=True,
        samesite="lax",
        max_age=60 * 60 * 24 * 7,
    )
    return {"ok": True}


@router.post("/logout")
async def logout(response: Response) -> dict[str, bool]:
    response.delete_cookie(SESSION_COOKIE_NAME)
    return {"ok": True}
```

- [ ] **Step 4: Write failing test `homelab-chatbot/backend/tests/unit/test_auth_route.py`**

```python
import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.auth import AuthService, hash_password
from app.routes.auth import router


@pytest.fixture
def client():
    app = FastAPI()
    app.state.auth = AuthService(
        password_hash=hash_password("letmein"), session_secret="x" * 32
    )
    app.include_router(router)
    return TestClient(app)


def test_login_with_correct_password_sets_cookie(client: TestClient):
    r = client.post("/api/auth/login", json={"password": "letmein"})
    assert r.status_code == 200
    assert "hlcb_session" in r.cookies


def test_login_with_wrong_password_rejected(client: TestClient):
    r = client.post("/api/auth/login", json={"password": "nope"})
    assert r.status_code == 401


def test_logout_clears_cookie(client: TestClient):
    client.post("/api/auth/login", json={"password": "letmein"})
    r = client.post("/api/auth/logout")
    assert r.status_code == 200
```

- [ ] **Step 5: Run test — verify pass**

Run: `uv run pytest tests/unit/test_auth_route.py -v`
Expected: PASS (3 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/deps.py homelab-chatbot/backend/app/routes/ homelab-chatbot/backend/tests/unit/test_auth_route.py
git commit -m "[feat:homelab-chatbot] add login/logout routes and session dependency"
```

---

### Task 22: Conversations CRUD route

**Files:**
- Create: `homelab-chatbot/backend/app/routes/conversations.py`
- Create: `homelab-chatbot/backend/tests/unit/test_conversations_route.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_conversations_route.py`**

```python
import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.auth import AuthService, hash_password
from app.routes import auth as auth_routes
from app.routes import conversations as conv_routes
from app.storage.chat_db import ChatDB


@pytest.fixture
async def app_client(tmp_path):
    app = FastAPI()
    app.state.auth = AuthService(
        password_hash=hash_password("p"), session_secret="x" * 32
    )
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/chat.db")
    await db.init_schema()
    app.state.chat_db = db
    app.include_router(auth_routes.router)
    app.include_router(conv_routes.router)
    client = TestClient(app)
    client.post("/api/auth/login", json={"password": "p"})
    yield client
    await db.close()


async def test_create_then_list_conversation(app_client: TestClient):
    r = app_client.post(
        "/api/conv",
        json={"title": "chat-a", "provider": "anthropic", "model": "claude-sonnet-4-6"},
    )
    assert r.status_code == 200
    conv_id = r.json()["id"]

    r = app_client.get("/api/conv")
    data = r.json()
    assert any(c["id"] == conv_id for c in data)


async def test_get_messages_returns_empty_on_new_conv(app_client: TestClient):
    r = app_client.post(
        "/api/conv", json={"title": "t", "provider": "ollama", "model": "llama3.1"}
    )
    conv_id = r.json()["id"]
    r = app_client.get(f"/api/conv/{conv_id}/messages")
    assert r.status_code == 200
    assert r.json() == []


async def test_delete_conversation(app_client: TestClient):
    r = app_client.post(
        "/api/conv", json={"title": "t", "provider": "ollama", "model": "m"}
    )
    conv_id = r.json()["id"]
    r = app_client.delete(f"/api/conv/{conv_id}")
    assert r.status_code == 200
    r = app_client.get(f"/api/conv/{conv_id}/messages")
    assert r.status_code == 404


async def test_rename_conversation(app_client: TestClient):
    r = app_client.post(
        "/api/conv", json={"title": "old", "provider": "ollama", "model": "m"}
    )
    conv_id = r.json()["id"]
    r = app_client.patch(f"/api/conv/{conv_id}", json={"title": "new"})
    assert r.status_code == 200
    r = app_client.get("/api/conv")
    assert any(c["title"] == "new" for c in r.json())


async def test_list_conversations_requires_auth(tmp_path):
    app = FastAPI()
    app.state.auth = AuthService(
        password_hash=hash_password("p"), session_secret="x" * 32
    )
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/c.db")
    await db.init_schema()
    app.state.chat_db = db
    app.include_router(conv_routes.router)
    client = TestClient(app)
    r = client.get("/api/conv")
    assert r.status_code == 401
    await db.close()
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/routes/conversations.py`**

```python
"""Conversation CRUD endpoints."""

from fastapi import APIRouter, Depends, HTTPException, Request, status
from pydantic import BaseModel

from app.deps import require_session
from app.models.chat import ConversationOut, MessageOut
from app.storage.chat_db import ChatDB

router = APIRouter(
    prefix="/api/conv", tags=["conversations"], dependencies=[Depends(require_session)]
)


class CreateConversationRequest(BaseModel):
    title: str
    provider: str
    model: str


class RenameRequest(BaseModel):
    title: str


def _db(request: Request) -> ChatDB:
    return request.app.state.chat_db


def _to_out(c) -> ConversationOut:
    return ConversationOut(
        id=c.id,
        title=c.title,
        provider=c.provider,
        model=c.model,
        created_at=c.created_at.isoformat(),
        updated_at=c.updated_at.isoformat(),
    )


@router.get("")
async def list_conversations(request: Request) -> list[ConversationOut]:
    db = _db(request)
    convs = await db.list_conversations()
    return [_to_out(c) for c in convs]


@router.post("")
async def create_conversation(
    body: CreateConversationRequest, request: Request
) -> ConversationOut:
    db = _db(request)
    conv = await db.create_conversation(
        title=body.title, provider=body.provider, model=body.model
    )
    return _to_out(conv)


@router.get("/{conv_id}/messages")
async def list_messages(conv_id: str, request: Request) -> list[MessageOut]:
    db = _db(request)
    conv = await db.get_conversation(conv_id)
    if not conv:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND)
    msgs = await db.list_messages(conv_id)
    return [
        MessageOut(
            id=m.id,
            role=m.role,
            content=m.content,
            tool_name=m.tool_name,
            tool_calls=m.tool_calls,
            partial=m.partial,
            created_at=m.created_at.isoformat(),
        )
        for m in msgs
    ]


@router.patch("/{conv_id}")
async def rename(conv_id: str, body: RenameRequest, request: Request) -> dict[str, bool]:
    db = _db(request)
    conv = await db.get_conversation(conv_id)
    if not conv:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND)
    await db.rename_conversation(conv_id, body.title)
    return {"ok": True}


@router.delete("/{conv_id}")
async def delete(conv_id: str, request: Request) -> dict[str, bool]:
    db = _db(request)
    await db.delete_conversation(conv_id)
    return {"ok": True}
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_conversations_route.py -v`
Expected: PASS (5 tests).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/routes/conversations.py homelab-chatbot/backend/tests/unit/test_conversations_route.py
git commit -m "[feat:homelab-chatbot] add conversations CRUD routes with auth"
```

---

### Task 23: Settings route (list providers/models)

**Files:**
- Create: `homelab-chatbot/backend/app/routes/settings.py`
- Create: `homelab-chatbot/backend/tests/unit/test_settings_route.py`

- [ ] **Step 1: Write failing test `homelab-chatbot/backend/tests/unit/test_settings_route.py`**

```python
import httpx
import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.auth import AuthService, hash_password
from app.config import AppConfig, EmbeddingsConfig, LLMConfig, OllamaConfig, PathConfig, RetrievalConfig, SyncConfig
from app.llm.registry import ModelRegistry
from app.routes import auth as auth_routes
from app.routes import settings as settings_routes


def _make_config() -> AppConfig:
    return AppConfig(
        sync=SyncConfig(interval_seconds=180, state_file="/tmp/s.json"),
        repos=[],
        embeddings=EmbeddingsConfig(model="m", cache_dir="/tmp"),
        vector_store=PathConfig(path="/tmp"),
        chat_db=PathConfig(path="/tmp/c.db"),
        kb_db=PathConfig(path="/tmp/k.db"),
        llm=LLMConfig(
            default_provider="anthropic",
            default_model="claude-sonnet-4-6",
            ollama=OllamaConfig(
                host="http://mock-ollama:11434",
                tool_capable_models=["llama3.1"],
            ),
        ),
        retrieval=RetrievalConfig(top_k=5, memory_turns=10),
    )


@pytest.fixture
def client():
    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    app.state.config = _make_config()
    app.state.registry = ModelRegistry("http://mock-ollama:11434")
    app.state.secrets_available = {"anthropic": True, "google": False}
    app.include_router(auth_routes.router)
    app.include_router(settings_routes.router)
    client = TestClient(app)
    client.post("/api/auth/login", json={"password": "p"})
    return client


def test_settings_lists_providers_with_models(respx_mock, client: TestClient):
    respx_mock.get("http://mock-ollama:11434/api/tags").mock(
        return_value=httpx.Response(
            200, json={"models": [{"name": "llama3.1:8b"}]}
        )
    )
    r = client.get("/api/settings")
    assert r.status_code == 200
    data = r.json()
    assert data["default_provider"] == "anthropic"
    providers = {p["id"]: p for p in data["providers"]}
    assert "anthropic" in providers
    assert providers["anthropic"]["available"] is True
    assert providers["google"]["available"] is False
    assert "claude-sonnet-4-6" in providers["anthropic"]["models"]
    assert "llama3.1:8b" in providers["ollama"]["models"]


def test_settings_requires_auth():
    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    app.state.config = _make_config()
    app.state.registry = ModelRegistry("http://x")
    app.state.secrets_available = {"anthropic": False, "google": False}
    app.include_router(settings_routes.router)
    client = TestClient(app)
    assert client.get("/api/settings").status_code == 401
```

- [ ] **Step 2: Run test — verify fails**

Expected: FAIL — module not found.

- [ ] **Step 3: Write `homelab-chatbot/backend/app/routes/settings.py`**

```python
"""Settings endpoint exposing available providers and models."""

from fastapi import APIRouter, Depends, Request
from pydantic import BaseModel

from app.config import AppConfig
from app.deps import require_session
from app.llm.capability import supports_tools
from app.llm.registry import ModelRegistry

router = APIRouter(
    prefix="/api/settings", tags=["settings"], dependencies=[Depends(require_session)]
)


class ProviderInfo(BaseModel):
    id: str
    available: bool
    models: list[str]
    tool_capable: list[str] = []


class SettingsOut(BaseModel):
    default_provider: str
    default_model: str
    providers: list[ProviderInfo]


@router.get("")
async def get_settings(request: Request) -> SettingsOut:
    cfg: AppConfig = request.app.state.config
    registry: ModelRegistry = request.app.state.registry
    secrets_available: dict[str, bool] = request.app.state.secrets_available

    anthropic_models = registry.static_models("anthropic")
    google_models = registry.static_models("google")
    ollama_models = await registry.ollama_models()

    providers = [
        ProviderInfo(
            id="anthropic",
            available=secrets_available.get("anthropic", False),
            models=anthropic_models,
            tool_capable=anthropic_models,
        ),
        ProviderInfo(
            id="google",
            available=secrets_available.get("google", False),
            models=google_models,
            tool_capable=google_models,
        ),
        ProviderInfo(
            id="ollama",
            available=len(ollama_models) > 0,
            models=ollama_models,
            tool_capable=[
                m
                for m in ollama_models
                if supports_tools(
                    provider="ollama",
                    model=m,
                    ollama_whitelist=cfg.llm.ollama.tool_capable_models,
                )
            ],
        ),
    ]

    return SettingsOut(
        default_provider=cfg.llm.default_provider,
        default_model=cfg.llm.default_model,
        providers=providers,
    )
```

- [ ] **Step 4: Run test — verify pass**

Run: `uv run pytest tests/unit/test_settings_route.py -v`
Expected: PASS (2 tests).

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/backend/app/routes/settings.py homelab-chatbot/backend/tests/unit/test_settings_route.py
git commit -m "[feat:homelab-chatbot] add settings route exposing providers and models"
```

---

### Task 24: Chat route + app assembly

**Files:**
- Create: `homelab-chatbot/backend/app/routes/chat.py`
- Create: `homelab-chatbot/backend/app/routes/health.py`
- Modify: `homelab-chatbot/backend/app/main.py`
- Create: `homelab-chatbot/backend/tests/unit/test_chat_route.py`

- [ ] **Step 1: Write `homelab-chatbot/backend/app/routes/health.py`**

```python
"""Health check endpoint (unauthenticated)."""

from fastapi import APIRouter, Request

router = APIRouter(prefix="/api", tags=["health"])


@router.get("/health")
async def health(request: Request) -> dict:
    state = getattr(request.app.state, "sync_state", {})
    return {"status": "ok", "sync": state}
```

- [ ] **Step 2: Write `homelab-chatbot/backend/app/routes/chat.py`**

```python
"""Streaming chat endpoint."""

import json
from typing import AsyncIterator

from fastapi import APIRouter, Depends, HTTPException, Request, status
from fastapi.responses import StreamingResponse

from app.config import AppConfig
from app.deps import require_session
from app.llm.agent import run_agent
from app.llm.capability import supports_tools
from app.llm.provider import build_llm
from app.llm.sse import AgentEvent, format_sse
from app.models.chat import ChatRequest
from app.retrieval.sql_tool import SQLTool
from app.retrieval.vector_tool import VectorSearchTool
from app.storage.chat_db import ChatDB

router = APIRouter(prefix="/api", tags=["chat"], dependencies=[Depends(require_session)])


async def _conv_history(db: ChatDB, conv_id: str, turns: int) -> list[tuple[str, str]]:
    msgs = await db.list_messages(conv_id)
    history = [(m.role, m.content) for m in msgs if m.role in ("user", "assistant")]
    return history[-turns * 2 :]


@router.post("/chat")
async def chat(body: ChatRequest, request: Request) -> StreamingResponse:
    cfg: AppConfig = request.app.state.config
    db: ChatDB = request.app.state.chat_db
    vector_tool: VectorSearchTool = request.app.state.vector_tool
    sql_tool: SQLTool = request.app.state.sql_tool
    secrets = request.app.state.secrets

    if body.conv_id:
        conv = await db.get_conversation(body.conv_id)
        if not conv:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND)
    else:
        title = body.message[:60] or "New chat"
        conv = await db.create_conversation(
            title=title, provider=body.provider, model=body.model
        )

    api_key = None
    if body.provider == "anthropic":
        api_key = secrets.anthropic_api_key
    elif body.provider == "google":
        api_key = secrets.google_api_key

    llm = build_llm(
        provider=body.provider,
        model=body.model,
        api_key=api_key,
        ollama_host=cfg.llm.ollama.host,
    )

    tools_enabled = supports_tools(
        provider=body.provider,
        model=body.model,
        ollama_whitelist=cfg.llm.ollama.tool_capable_models,
    )
    tools = (
        [vector_tool.as_llama_tool(), sql_tool.as_llama_tool(llm)] if tools_enabled else []
    )

    history = await _conv_history(db, conv.id, cfg.retrieval.memory_turns)

    await db.append_message(conv.id, role="user", content=body.message)

    async def _event_stream() -> AsyncIterator[bytes]:
        buffer: list[str] = []
        tool_events: list[dict] = []
        yield format_sse(AgentEvent("conversation", {"id": conv.id}))
        try:
            async for event in run_agent(
                llm=llm,
                tools=tools,
                history=history,
                user_message=body.message,
                tools_enabled=tools_enabled,
            ):
                if event.kind == "text-delta":
                    buffer.append(event.data.get("text", ""))
                elif event.kind in ("tool-call", "tool-result"):
                    tool_events.append({"kind": event.kind, **event.data})
                yield format_sse(event)
        finally:
            full = "".join(buffer)
            await db.append_message(
                conv.id,
                role="assistant",
                content=full,
                tool_calls=json.dumps(tool_events) if tool_events else None,
                partial=False,
            )

    return StreamingResponse(_event_stream(), media_type="text/event-stream")
```

- [ ] **Step 3: Rewrite `homelab-chatbot/backend/app/main.py`**

```python
"""FastAPI application factory wiring together all dependencies."""

import os
from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from app.auth import AuthService
from app.config import AppConfig, load_config
from app.ingestion.embed import Embedder
from app.ingestion.scheduler import SyncScheduler
from app.llm.registry import ModelRegistry
from app.retrieval.sql_tool import SQLTool
from app.retrieval.vector_tool import VectorSearchTool
from app.routes import auth as auth_routes
from app.routes import chat as chat_routes
from app.routes import conversations as conv_routes
from app.routes import health as health_routes
from app.routes import settings as settings_routes
from app.secrets import Secrets
from app.storage.chat_db import ChatDB
from app.storage.lance import VectorStore


def create_app(config_path: str | None = None, with_scheduler: bool = True) -> FastAPI:
    app = FastAPI(title="homelab-chatbot")

    path = Path(config_path or os.environ.get("HLCB_CONFIG_PATH", "config.yaml"))
    cfg = load_config(path) if path.exists() else _minimal_default_config()

    secrets = Secrets()
    app.state.config = cfg
    app.state.secrets = secrets
    app.state.secrets_available = {
        "anthropic": secrets.anthropic_api_key is not None,
        "google": secrets.google_api_key is not None,
    }

    app.state.auth = AuthService(
        password_hash=secrets.auth_password_hash, session_secret=secrets.session_secret
    )

    chat_db = ChatDB(f"sqlite+aiosqlite:///{cfg.chat_db.path}")
    app.state.chat_db = chat_db

    vector_store = VectorStore(Path(cfg.vector_store.path))
    embedder = Embedder(
        model_name=cfg.embeddings.model, cache_dir=cfg.embeddings.cache_dir
    )
    app.state.vector_tool = VectorSearchTool(
        store=vector_store, embedder=embedder, top_k=cfg.retrieval.top_k
    )
    app.state.sql_tool = SQLTool(db_path=Path(cfg.kb_db.path))
    app.state.registry = ModelRegistry(cfg.llm.ollama.host)
    app.state.sync_state = {"last_sync_at": None}

    scheduler: SyncScheduler | None = None

    @app.on_event("startup")
    async def _startup() -> None:
        await chat_db.init_schema()
        nonlocal scheduler
        if with_scheduler:
            scheduler = SyncScheduler(
                config=cfg,
                clone_root=Path("/data/repos"),
                get_token=secrets.github_token,
            )
            scheduler.start()

    @app.on_event("shutdown")
    async def _shutdown() -> None:
        if scheduler:
            scheduler.shutdown()
        await chat_db.close()

    app.include_router(auth_routes.router)
    app.include_router(health_routes.router)
    app.include_router(conv_routes.router)
    app.include_router(chat_routes.router)
    app.include_router(settings_routes.router)

    static_dir = Path(__file__).parent.parent / "static"
    if static_dir.exists():
        app.mount("/", StaticFiles(directory=str(static_dir), html=True), name="frontend")

    return app


def _minimal_default_config() -> AppConfig:
    from app.config import EmbeddingsConfig, LLMConfig, OllamaConfig, PathConfig, RetrievalConfig, SyncConfig

    return AppConfig(
        sync=SyncConfig(interval_seconds=180, state_file="/data/sync_state.json"),
        repos=[],
        embeddings=EmbeddingsConfig(
            model="BAAI/bge-small-en-v1.5", cache_dir="/data/models"
        ),
        vector_store=PathConfig(path="/data/lance"),
        chat_db=PathConfig(path="/data/chat.db"),
        kb_db=PathConfig(path="/data/kb.db"),
        llm=LLMConfig(
            default_provider="anthropic",
            default_model="claude-sonnet-4-6",
            ollama=OllamaConfig(host="http://localhost:11434", tool_capable_models=[]),
        ),
        retrieval=RetrievalConfig(top_k=5, memory_turns=10),
    )


app = create_app()
```

- [ ] **Step 4: Write failing test `homelab-chatbot/backend/tests/unit/test_chat_route.py`**

```python
from pathlib import Path
from typing import AsyncIterator

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient
from llama_index.core.base.llms.types import ChatMessage, ChatResponse

from app.auth import AuthService, hash_password
from app.config import AppConfig, EmbeddingsConfig, LLMConfig, OllamaConfig, PathConfig, RetrievalConfig, SyncConfig
from app.retrieval.sql_tool import SQLTool
from app.retrieval.vector_tool import VectorSearchTool
from app.routes import auth as auth_routes
from app.routes import chat as chat_routes
from app.routes import conversations as conv_routes
from app.storage.chat_db import ChatDB


class StubEmbedder:
    def embed_batch(self, texts): return [[0.0] * 384 for _ in texts]


class StubVectorStore:
    def search(self, **kwargs): return []
    def count(self): return 0


class StubLLM:
    async def astream_chat(self, messages) -> AsyncIterator[ChatResponse]:
        for ch in "hi":
            yield ChatResponse(
                message=ChatMessage(role="assistant", content=ch),
                delta=ch,
                raw={},
            )


@pytest.fixture
async def client(tmp_path: Path, monkeypatch):
    monkeypatch.setattr(
        "app.routes.chat.build_llm", lambda **kwargs: StubLLM()
    )

    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/c.db")
    await db.init_schema()
    app.state.chat_db = db
    app.state.vector_tool = VectorSearchTool(
        store=StubVectorStore(), embedder=StubEmbedder(), top_k=1
    )
    app.state.sql_tool = SQLTool(db_path=tmp_path / "kb.db")
    app.state.config = AppConfig(
        sync=SyncConfig(interval_seconds=60, state_file=str(tmp_path / "s.json")),
        repos=[],
        embeddings=EmbeddingsConfig(model="m", cache_dir=str(tmp_path)),
        vector_store=PathConfig(path=str(tmp_path)),
        chat_db=PathConfig(path=str(tmp_path / "c.db")),
        kb_db=PathConfig(path=str(tmp_path / "kb.db")),
        llm=LLMConfig(
            default_provider="ollama",
            default_model="phi3",
            ollama=OllamaConfig(host="http://x", tool_capable_models=[]),
        ),
        retrieval=RetrievalConfig(top_k=1, memory_turns=2),
    )

    class FakeSecrets:
        anthropic_api_key = None
        google_api_key = None

    app.state.secrets = FakeSecrets()
    app.include_router(auth_routes.router)
    app.include_router(conv_routes.router)
    app.include_router(chat_routes.router)
    c = TestClient(app)
    c.post("/api/auth/login", json={"password": "p"})
    yield c
    await db.close()


async def test_chat_creates_conversation_and_streams(client: TestClient):
    with client.stream(
        "POST",
        "/api/chat",
        json={"message": "hello", "provider": "ollama", "model": "phi3"},
    ) as resp:
        assert resp.status_code == 200
        body = b"".join(resp.iter_bytes())
    text = body.decode("utf-8")
    assert "event: conversation" in text
    assert "event: text-delta" in text
    assert "event: done" in text


async def test_chat_requires_auth(tmp_path: Path):
    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/x.db")
    await db.init_schema()
    app.state.chat_db = db
    app.include_router(chat_routes.router)
    c = TestClient(app)
    r = c.post(
        "/api/chat", json={"message": "hi", "provider": "ollama", "model": "m"}
    )
    assert r.status_code == 401
    await db.close()
```

- [ ] **Step 5: Run all backend tests — verify pass**

Run: `uv run pytest tests/unit -v`
Expected: all previously-written tests + the new 2 chat-route tests pass.

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/backend/app/routes/chat.py homelab-chatbot/backend/app/routes/health.py homelab-chatbot/backend/app/main.py homelab-chatbot/backend/tests/unit/test_chat_route.py
git commit -m "[feat:homelab-chatbot] add chat route with SSE streaming and app assembly"
```

---

## Phase 6 — Frontend

### Task 25: Frontend utility + API client + login page

**Files:**
- Create: `homelab-chatbot/frontend/lib/cn.ts`
- Create: `homelab-chatbot/frontend/lib/api.ts`
- Create: `homelab-chatbot/frontend/app/login/page.tsx`
- Create: `homelab-chatbot/frontend/tests/login.test.tsx`

- [ ] **Step 1: Write `homelab-chatbot/frontend/lib/cn.ts`**

```typescript
import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}
```

- [ ] **Step 2: Write `homelab-chatbot/frontend/lib/api.ts`**

```typescript
export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(path, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });
  if (!resp.ok) {
    throw new ApiError(resp.status, await resp.text());
  }
  return resp.status === 204 ? (undefined as T) : ((await resp.json()) as T);
}

export const api = {
  login: (password: string) =>
    request<{ ok: boolean }>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ password }),
    }),
  logout: () => request<{ ok: boolean }>("/api/auth/logout", { method: "POST" }),

  listConversations: () =>
    request<Conversation[]>("/api/conv"),

  createConversation: (body: { title: string; provider: string; model: string }) =>
    request<Conversation>("/api/conv", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  listMessages: (id: string) =>
    request<Message[]>(`/api/conv/${id}/messages`),

  renameConversation: (id: string, title: string) =>
    request<{ ok: boolean }>(`/api/conv/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ title }),
    }),

  deleteConversation: (id: string) =>
    request<{ ok: boolean }>(`/api/conv/${id}`, { method: "DELETE" }),

  getSettings: () => request<Settings>("/api/settings"),
};

export interface Conversation {
  id: string;
  title: string;
  provider: string;
  model: string;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: number;
  role: "user" | "assistant" | "tool";
  content: string;
  tool_name: string | null;
  tool_calls: string | null;
  partial: boolean;
  created_at: string;
}

export interface ProviderInfo {
  id: string;
  available: boolean;
  models: string[];
  tool_capable: string[];
}

export interface Settings {
  default_provider: string;
  default_model: string;
  providers: ProviderInfo[];
}
```

- [ ] **Step 3: Write `homelab-chatbot/frontend/app/login/page.tsx`**

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

import { api, ApiError } from "@/lib/api";

export default function LoginPage() {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await api.login(password);
      router.push("/");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError("Incorrect password.");
      } else {
        setError("Login failed.");
      }
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="flex min-h-screen items-center justify-center">
      <form
        onSubmit={onSubmit}
        className="w-80 rounded-lg border border-neutral-800 bg-neutral-900 p-6"
      >
        <h1 className="mb-4 text-lg font-semibold">homelab-chatbot</h1>
        <label className="mb-2 block text-sm text-neutral-400">Password</label>
        <input
          type="password"
          className="mb-3 w-full rounded border border-neutral-700 bg-neutral-950 p-2"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          autoFocus
          required
        />
        {error && <p className="mb-3 text-sm text-red-400">{error}</p>}
        <button
          type="submit"
          disabled={loading}
          className="w-full rounded bg-blue-600 p-2 font-medium disabled:opacity-50"
        >
          {loading ? "Signing in..." : "Sign in"}
        </button>
      </form>
    </main>
  );
}
```

- [ ] **Step 4: Write component test `homelab-chatbot/frontend/tests/login.test.tsx`**

```tsx
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

import LoginPage from "@/app/login/page";

const push = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
}));

describe("LoginPage", () => {
  beforeEach(() => {
    push.mockReset();
    global.fetch = vi.fn();
  });

  it("redirects on successful login", async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ ok: true }),
    });
    render(<LoginPage />);
    await userEvent.type(screen.getByLabelText(/password/i), "letmein");
    await userEvent.click(screen.getByRole("button", { name: /sign in/i }));
    await waitFor(() => expect(push).toHaveBeenCalledWith("/"));
  });

  it("shows error on 401", async () => {
    (global.fetch as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: false,
      status: 401,
      text: async () => "nope",
    });
    render(<LoginPage />);
    await userEvent.type(screen.getByLabelText(/password/i), "wrong");
    await userEvent.click(screen.getByRole("button", { name: /sign in/i }));
    expect(await screen.findByText(/incorrect password/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 5: Run frontend tests**

Run:
```bash
cd homelab-chatbot/frontend
npm run test
```
Expected: PASS (2 tests).

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/frontend/lib/ homelab-chatbot/frontend/app/login/ homelab-chatbot/frontend/tests/login.test.tsx
git commit -m "[feat:homelab-chatbot] add frontend API client and login page"
```

---

### Task 26: Chat page with streaming (Vercel AI SDK)

**Files:**
- Create: `homelab-chatbot/frontend/components/chat/ChatThread.tsx`
- Create: `homelab-chatbot/frontend/components/chat/Composer.tsx`
- Create: `homelab-chatbot/frontend/components/chat/StreamingMessage.tsx`
- Create: `homelab-chatbot/frontend/components/chat/ToolCallBadge.tsx`
- Create: `homelab-chatbot/frontend/lib/stream.ts`
- Modify: `homelab-chatbot/frontend/app/page.tsx`
- Create: `homelab-chatbot/frontend/tests/chat-thread.test.tsx`

- [ ] **Step 1: Write `homelab-chatbot/frontend/lib/stream.ts`**

```typescript
import type { Message } from "@/lib/api";

export type StreamEvent =
  | { kind: "conversation"; id: string }
  | { kind: "text-delta"; text: string }
  | { kind: "tool-call"; name: string; args: unknown }
  | { kind: "tool-result"; name: string; summary: string }
  | { kind: "error"; message: string }
  | { kind: "done" };

export async function* parseSseStream(
  response: Response,
): AsyncGenerator<StreamEvent> {
  if (!response.body) return;
  const reader = response.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  while (true) {
    const { value, done } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const frames = buffer.split("\n\n");
    buffer = frames.pop() ?? "";
    for (const frame of frames) {
      const ev = parseFrame(frame);
      if (ev) yield ev;
    }
  }
}

function parseFrame(frame: string): StreamEvent | null {
  const lines = frame.split("\n");
  let event: string | null = null;
  let data: string | null = null;
  for (const line of lines) {
    if (line.startsWith("event:")) event = line.slice(6).trim();
    else if (line.startsWith("data:")) data = line.slice(5).trim();
  }
  if (!event || data === null) return null;
  try {
    const parsed = data === "" ? {} : JSON.parse(data);
    return { kind: event as StreamEvent["kind"], ...parsed } as StreamEvent;
  } catch {
    return null;
  }
}

export interface DisplayMessage {
  id: string;
  role: Message["role"];
  content: string;
  toolEvents?: Array<{ kind: string; name: string; summary?: string }>;
  partial?: boolean;
}
```

- [ ] **Step 2: Write `homelab-chatbot/frontend/components/chat/ToolCallBadge.tsx`**

```tsx
import { cn } from "@/lib/cn";

export function ToolCallBadge({ name }: { name: string }) {
  const icon = name.includes("inventory") ? "📊" : "🔍";
  const label = name.includes("inventory") ? "querying inventory" : "searching docs";
  return (
    <span
      className={cn(
        "inline-flex items-center gap-1 rounded-full border border-neutral-700",
        "bg-neutral-800 px-2 py-0.5 text-xs text-neutral-300",
      )}
    >
      <span>{icon}</span>
      <span>{label}</span>
    </span>
  );
}
```

- [ ] **Step 3: Write `homelab-chatbot/frontend/components/chat/StreamingMessage.tsx`**

```tsx
"use client";

import ReactMarkdown from "react-markdown";

import { ToolCallBadge } from "./ToolCallBadge";
import type { DisplayMessage } from "@/lib/stream";

export function StreamingMessage({ message }: { message: DisplayMessage }) {
  const isUser = message.role === "user";
  return (
    <div className={`flex ${isUser ? "justify-end" : "justify-start"}`}>
      <div
        className={`max-w-3xl rounded-lg px-4 py-3 ${
          isUser ? "bg-blue-600 text-white" : "bg-neutral-800 text-neutral-100"
        }`}
      >
        {message.toolEvents?.length ? (
          <div className="mb-2 flex flex-wrap gap-1">
            {message.toolEvents.map((t, i) => (
              <ToolCallBadge key={i} name={t.name} />
            ))}
          </div>
        ) : null}
        <div className="prose prose-invert max-w-none text-sm">
          <ReactMarkdown>{message.content || "…"}</ReactMarkdown>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 4: Write `homelab-chatbot/frontend/components/chat/Composer.tsx`**

```tsx
"use client";

import { useState } from "react";

export function Composer({
  onSend,
  disabled,
}: {
  onSend: (text: string) => void;
  disabled: boolean;
}) {
  const [value, setValue] = useState("");

  function submit() {
    const trimmed = value.trim();
    if (!trimmed || disabled) return;
    onSend(trimmed);
    setValue("");
  }

  return (
    <div className="border-t border-neutral-800 bg-neutral-900 p-3">
      <textarea
        aria-label="message"
        value={value}
        disabled={disabled}
        onChange={(e) => setValue(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            submit();
          }
        }}
        rows={2}
        placeholder="Ask about the home lab…"
        className="w-full resize-none rounded border border-neutral-700 bg-neutral-950 p-2 text-sm outline-none focus:border-blue-500"
      />
      <div className="mt-2 flex justify-end">
        <button
          onClick={submit}
          disabled={disabled || !value.trim()}
          className="rounded bg-blue-600 px-4 py-1.5 text-sm font-medium disabled:opacity-50"
        >
          Send
        </button>
      </div>
    </div>
  );
}
```

- [ ] **Step 5: Write `homelab-chatbot/frontend/components/chat/ChatThread.tsx`**

```tsx
"use client";

import { useEffect, useRef } from "react";

import { StreamingMessage } from "./StreamingMessage";
import type { DisplayMessage } from "@/lib/stream";

export function ChatThread({ messages }: { messages: DisplayMessage[] }) {
  const endRef = useRef<HTMLDivElement>(null);
  useEffect(() => {
    endRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  return (
    <div className="flex-1 space-y-3 overflow-y-auto p-4">
      {messages.length === 0 && (
        <p className="mt-8 text-center text-neutral-500">
          Start a conversation about your home lab.
        </p>
      )}
      {messages.map((m) => (
        <StreamingMessage key={m.id} message={m} />
      ))}
      <div ref={endRef} />
    </div>
  );
}
```

- [ ] **Step 6: Rewrite `homelab-chatbot/frontend/app/page.tsx`**

```tsx
"use client";

import { useCallback, useEffect, useState } from "react";

import { ChatThread } from "@/components/chat/ChatThread";
import { Composer } from "@/components/chat/Composer";
import { api, type Conversation, type Message } from "@/lib/api";
import { parseSseStream, type DisplayMessage } from "@/lib/stream";

export default function Home() {
  const [provider, setProvider] = useState("anthropic");
  const [model, setModel] = useState("claude-sonnet-4-6");
  const [convId, setConvId] = useState<string | null>(null);
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [streaming, setStreaming] = useState(false);

  useEffect(() => {
    api.getSettings().then((s) => {
      setProvider(s.default_provider);
      setModel(s.default_model);
    }).catch(() => {});
  }, []);

  const send = useCallback(
    async (text: string) => {
      setMessages((prev) => [
        ...prev,
        { id: `u-${Date.now()}`, role: "user", content: text },
        { id: `a-${Date.now()}`, role: "assistant", content: "", partial: true },
      ]);
      setStreaming(true);
      try {
        const resp = await fetch("/api/chat", {
          method: "POST",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            conv_id: convId,
            message: text,
            provider,
            model,
          }),
        });
        if (!resp.ok) throw new Error("chat failed");

        for await (const ev of parseSseStream(resp)) {
          if (ev.kind === "conversation") {
            setConvId(ev.id);
          } else if (ev.kind === "text-delta") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                copy[copy.length - 1] = {
                  ...last,
                  content: last.content + ev.text,
                };
              }
              return copy;
            });
          } else if (ev.kind === "tool-call" || ev.kind === "tool-result") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                const events = [...(last.toolEvents ?? []),
                  { kind: ev.kind, name: ev.name, summary: (ev as {summary?: string}).summary ?? "" }];
                copy[copy.length - 1] = { ...last, toolEvents: events };
              }
              return copy;
            });
          } else if (ev.kind === "done") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                copy[copy.length - 1] = { ...last, partial: false };
              }
              return copy;
            });
          }
        }
      } finally {
        setStreaming(false);
      }
    },
    [convId, provider, model],
  );

  return (
    <main className="flex h-screen flex-col">
      <header className="flex items-center gap-2 border-b border-neutral-800 p-3 text-sm">
        <span className="text-neutral-400">Provider:</span>
        <span>{provider}</span>
        <span className="text-neutral-400">· Model:</span>
        <span>{model}</span>
      </header>
      <ChatThread messages={messages} />
      <Composer onSend={send} disabled={streaming} />
    </main>
  );
}
```

- [ ] **Step 7: Write component test `homelab-chatbot/frontend/tests/chat-thread.test.tsx`**

```tsx
import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";

import { ChatThread } from "@/components/chat/ChatThread";
import type { DisplayMessage } from "@/lib/stream";

const sample: DisplayMessage[] = [
  { id: "1", role: "user", content: "hello" },
  { id: "2", role: "assistant", content: "hi there" },
];

describe("ChatThread", () => {
  it("renders all messages", () => {
    render(<ChatThread messages={sample} />);
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(screen.getByText("hi there")).toBeInTheDocument();
  });

  it("shows placeholder when empty", () => {
    render(<ChatThread messages={[]} />);
    expect(screen.getByText(/start a conversation/i)).toBeInTheDocument();
  });
});
```

- [ ] **Step 8: Run frontend tests**

Run: `npm run test`
Expected: PASS.

- [ ] **Step 9: Commit**

```bash
git add homelab-chatbot/frontend/components/ homelab-chatbot/frontend/lib/stream.ts homelab-chatbot/frontend/app/page.tsx homelab-chatbot/frontend/tests/chat-thread.test.tsx
git commit -m "[feat:homelab-chatbot] add chat UI with SSE streaming and tool-call badges"
```

---

### Task 27: Conversation sidebar with SWR

**Files:**
- Create: `homelab-chatbot/frontend/components/sidebar/ConversationSidebar.tsx`
- Create: `homelab-chatbot/frontend/components/sidebar/NewChatButton.tsx`
- Modify: `homelab-chatbot/frontend/app/page.tsx` (mount sidebar, wire conv selection)
- Create: `homelab-chatbot/frontend/tests/sidebar.test.tsx`

- [ ] **Step 1: Write `homelab-chatbot/frontend/components/sidebar/NewChatButton.tsx`**

```tsx
"use client";

export function NewChatButton({ onClick }: { onClick: () => void }) {
  return (
    <button
      onClick={onClick}
      className="mb-2 w-full rounded bg-blue-600 p-2 text-sm font-medium"
    >
      + New chat
    </button>
  );
}
```

- [ ] **Step 2: Write `homelab-chatbot/frontend/components/sidebar/ConversationSidebar.tsx`**

```tsx
"use client";

import useSWR from "swr";

import { NewChatButton } from "./NewChatButton";
import { api, type Conversation } from "@/lib/api";
import { cn } from "@/lib/cn";

const fetcher = () => api.listConversations();

export function ConversationSidebar({
  activeId,
  onSelect,
  onNew,
}: {
  activeId: string | null;
  onSelect: (c: Conversation) => void;
  onNew: () => void;
}) {
  const { data, mutate } = useSWR<Conversation[]>("/api/conv", fetcher, {
    refreshInterval: 0,
    revalidateOnFocus: true,
  });

  async function handleDelete(id: string, e: React.MouseEvent) {
    e.stopPropagation();
    if (!confirm("Delete this conversation?")) return;
    await api.deleteConversation(id);
    await mutate();
    if (activeId === id) onNew();
  }

  return (
    <aside className="flex w-64 flex-col border-r border-neutral-800 bg-neutral-900 p-2">
      <NewChatButton onClick={onNew} />
      <ul className="flex-1 space-y-1 overflow-y-auto">
        {(data ?? []).map((c) => (
          <li key={c.id}>
            <button
              onClick={() => onSelect(c)}
              className={cn(
                "flex w-full items-center justify-between rounded px-2 py-1.5 text-left text-sm",
                activeId === c.id ? "bg-neutral-800" : "hover:bg-neutral-800/60",
              )}
            >
              <span className="truncate">{c.title}</span>
              <span
                role="button"
                aria-label={`delete ${c.title}`}
                onClick={(e) => handleDelete(c.id, e)}
                className="ml-2 text-neutral-500 hover:text-red-400"
              >
                ×
              </span>
            </button>
          </li>
        ))}
        {(data ?? []).length === 0 && (
          <li className="p-2 text-center text-sm text-neutral-500">
            No conversations yet.
          </li>
        )}
      </ul>
    </aside>
  );
}
```

- [ ] **Step 3: Update `homelab-chatbot/frontend/app/page.tsx` to mount sidebar**

Replace the `<main>` wrapper to add the sidebar and wire selection:

```tsx
// add imports at top:
// import { ConversationSidebar } from "@/components/sidebar/ConversationSidebar";
// import { mutate } from "swr";

// inside component, add:
// const onSelect = useCallback(async (c: Conversation) => {
//   setConvId(c.id);
//   setProvider(c.provider);
//   setModel(c.model);
//   const msgs = await api.listMessages(c.id);
//   setMessages(
//     msgs.map((m: Message) => ({ id: String(m.id), role: m.role, content: m.content }))
//   );
// }, []);
// const onNew = useCallback(() => {
//   setConvId(null);
//   setMessages([]);
// }, []);

// after send() finishes (in the `done` branch), trigger conversation list refresh:
// await mutate("/api/conv");

// Replace the outer <main> with:
// return (
//   <div className="flex h-screen">
//     <ConversationSidebar activeId={convId} onSelect={onSelect} onNew={onNew} />
//     <main className="flex flex-1 flex-col">
//       {/* existing header, ChatThread, Composer */}
//     </main>
//   </div>
// );
```

The engineer should integrate these changes into the existing `page.tsx`. The final file:

```tsx
"use client";

import { useCallback, useEffect, useState } from "react";
import { mutate } from "swr";

import { ChatThread } from "@/components/chat/ChatThread";
import { Composer } from "@/components/chat/Composer";
import { ConversationSidebar } from "@/components/sidebar/ConversationSidebar";
import { api, type Conversation, type Message } from "@/lib/api";
import { parseSseStream, type DisplayMessage } from "@/lib/stream";

export default function Home() {
  const [provider, setProvider] = useState("anthropic");
  const [model, setModel] = useState("claude-sonnet-4-6");
  const [convId, setConvId] = useState<string | null>(null);
  const [messages, setMessages] = useState<DisplayMessage[]>([]);
  const [streaming, setStreaming] = useState(false);

  useEffect(() => {
    api.getSettings().then((s) => {
      setProvider(s.default_provider);
      setModel(s.default_model);
    }).catch(() => {});
  }, []);

  const onSelect = useCallback(async (c: Conversation) => {
    setConvId(c.id);
    setProvider(c.provider);
    setModel(c.model);
    const msgs = await api.listMessages(c.id);
    setMessages(
      msgs.map((m: Message) => ({
        id: String(m.id),
        role: m.role,
        content: m.content,
      })),
    );
  }, []);

  const onNew = useCallback(() => {
    setConvId(null);
    setMessages([]);
  }, []);

  const send = useCallback(
    async (text: string) => {
      setMessages((prev) => [
        ...prev,
        { id: `u-${Date.now()}`, role: "user", content: text },
        { id: `a-${Date.now()}`, role: "assistant", content: "", partial: true },
      ]);
      setStreaming(true);
      try {
        const resp = await fetch("/api/chat", {
          method: "POST",
          credentials: "include",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ conv_id: convId, message: text, provider, model }),
        });
        if (!resp.ok) throw new Error("chat failed");

        for await (const ev of parseSseStream(resp)) {
          if (ev.kind === "conversation") {
            setConvId(ev.id);
          } else if (ev.kind === "text-delta") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                copy[copy.length - 1] = { ...last, content: last.content + ev.text };
              }
              return copy;
            });
          } else if (ev.kind === "tool-call" || ev.kind === "tool-result") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                const events = [
                  ...(last.toolEvents ?? []),
                  {
                    kind: ev.kind,
                    name: ev.name,
                    summary: (ev as { summary?: string }).summary ?? "",
                  },
                ];
                copy[copy.length - 1] = { ...last, toolEvents: events };
              }
              return copy;
            });
          } else if (ev.kind === "done") {
            setMessages((prev) => {
              const copy = [...prev];
              const last = copy[copy.length - 1];
              if (last?.role === "assistant") {
                copy[copy.length - 1] = { ...last, partial: false };
              }
              return copy;
            });
          }
        }
        await mutate("/api/conv");
      } finally {
        setStreaming(false);
      }
    },
    [convId, provider, model],
  );

  return (
    <div className="flex h-screen">
      <ConversationSidebar activeId={convId} onSelect={onSelect} onNew={onNew} />
      <main className="flex flex-1 flex-col">
        <header className="flex items-center gap-2 border-b border-neutral-800 p-3 text-sm">
          <span className="text-neutral-400">Provider:</span>
          <span>{provider}</span>
          <span className="text-neutral-400">· Model:</span>
          <span>{model}</span>
        </header>
        <ChatThread messages={messages} />
        <Composer onSend={send} disabled={streaming} />
      </main>
    </div>
  );
}
```

- [ ] **Step 4: Write component test `homelab-chatbot/frontend/tests/sidebar.test.tsx`**

```tsx
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen } from "@testing-library/react";
import { SWRConfig } from "swr";

import { ConversationSidebar } from "@/components/sidebar/ConversationSidebar";

describe("ConversationSidebar", () => {
  beforeEach(() => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => [
        {
          id: "a",
          title: "First",
          provider: "anthropic",
          model: "x",
          created_at: "2026-04-16T00:00:00",
          updated_at: "2026-04-16T00:00:00",
        },
      ],
    });
  });

  it("renders conversation titles", async () => {
    render(
      <SWRConfig value={{ provider: () => new Map() }}>
        <ConversationSidebar activeId={null} onSelect={() => {}} onNew={() => {}} />
      </SWRConfig>,
    );
    expect(await screen.findByText("First")).toBeInTheDocument();
  });

  it("renders New chat button", () => {
    render(
      <SWRConfig value={{ provider: () => new Map() }}>
        <ConversationSidebar activeId={null} onSelect={() => {}} onNew={() => {}} />
      </SWRConfig>,
    );
    expect(screen.getByRole("button", { name: /new chat/i })).toBeInTheDocument();
  });
});
```

- [ ] **Step 5: Run frontend tests**

Run: `npm run test`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/frontend/components/sidebar/ homelab-chatbot/frontend/app/page.tsx homelab-chatbot/frontend/tests/sidebar.test.tsx
git commit -m "[feat:homelab-chatbot] add conversation sidebar with SWR-backed list"
```

---

### Task 28: Provider picker + settings page

**Files:**
- Create: `homelab-chatbot/frontend/components/settings/ProviderPicker.tsx`
- Create: `homelab-chatbot/frontend/app/settings/page.tsx`
- Modify: `homelab-chatbot/frontend/app/page.tsx` (mount ProviderPicker in header)
- Create: `homelab-chatbot/frontend/tests/provider-picker.test.tsx`

- [ ] **Step 1: Write `homelab-chatbot/frontend/components/settings/ProviderPicker.tsx`**

```tsx
"use client";

import useSWR from "swr";

import { api, type Settings } from "@/lib/api";

const fetcher = () => api.getSettings();

export function ProviderPicker({
  provider,
  model,
  onChange,
}: {
  provider: string;
  model: string;
  onChange: (p: string, m: string) => void;
}) {
  const { data } = useSWR<Settings>("/api/settings", fetcher);
  if (!data) {
    return <span className="text-xs text-neutral-500">loading providers…</span>;
  }

  const current = data.providers.find((p) => p.id === provider);
  const models = current?.models ?? [];

  return (
    <div className="flex items-center gap-2 text-sm">
      <select
        aria-label="provider"
        className="rounded border border-neutral-700 bg-neutral-950 px-2 py-1"
        value={provider}
        onChange={(e) => {
          const next = data.providers.find((p) => p.id === e.target.value);
          const firstModel = next?.models[0] ?? "";
          onChange(e.target.value, firstModel);
        }}
      >
        {data.providers.map((p) => (
          <option key={p.id} value={p.id} disabled={!p.available}>
            {p.id}
            {!p.available ? " (unavailable)" : ""}
          </option>
        ))}
      </select>
      <select
        aria-label="model"
        className="rounded border border-neutral-700 bg-neutral-950 px-2 py-1"
        value={model}
        onChange={(e) => onChange(provider, e.target.value)}
      >
        {models.map((m) => (
          <option key={m} value={m}>
            {m}
          </option>
        ))}
      </select>
    </div>
  );
}
```

- [ ] **Step 2: Update `homelab-chatbot/frontend/app/page.tsx` header to use the picker**

Replace the header element with:

```tsx
<header className="flex items-center gap-3 border-b border-neutral-800 p-3">
  <ProviderPicker
    provider={provider}
    model={model}
    onChange={(p, m) => {
      setProvider(p);
      setModel(m);
    }}
  />
</header>
```

And add the import: `import { ProviderPicker } from "@/components/settings/ProviderPicker";`

- [ ] **Step 3: Write `homelab-chatbot/frontend/app/settings/page.tsx`**

```tsx
"use client";

import useSWR from "swr";

import { api, type Settings } from "@/lib/api";

const fetcher = () => api.getSettings();

export default function SettingsPage() {
  const { data, error } = useSWR<Settings>("/api/settings", fetcher);

  if (error) return <p className="p-4 text-red-400">Failed to load settings.</p>;
  if (!data) return <p className="p-4 text-neutral-500">Loading…</p>;

  return (
    <main className="p-6">
      <h1 className="mb-4 text-xl font-semibold">Settings</h1>
      <p className="mb-2 text-sm text-neutral-400">
        Default: <span className="text-neutral-200">{data.default_provider}</span> / {data.default_model}
      </p>
      <div className="space-y-4">
        {data.providers.map((p) => (
          <section key={p.id} className="rounded border border-neutral-800 p-4">
            <h2 className="mb-2 font-medium">
              {p.id} {p.available ? (
                <span className="text-green-400 text-xs">available</span>
              ) : (
                <span className="text-red-400 text-xs">unavailable</span>
              )}
            </h2>
            <p className="text-sm text-neutral-400">
              {p.models.length === 0
                ? "No models available."
                : `Models: ${p.models.join(", ")}`}
            </p>
            {p.id === "ollama" && p.models.length > 0 && (
              <p className="mt-2 text-xs text-neutral-500">
                Tool-capable: {p.tool_capable.join(", ") || "none"}
              </p>
            )}
          </section>
        ))}
      </div>
    </main>
  );
}
```

- [ ] **Step 4: Write component test `homelab-chatbot/frontend/tests/provider-picker.test.tsx`**

```tsx
import { describe, expect, it, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SWRConfig } from "swr";

import { ProviderPicker } from "@/components/settings/ProviderPicker";

describe("ProviderPicker", () => {
  beforeEach(() => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({
        default_provider: "anthropic",
        default_model: "claude-sonnet-4-6",
        providers: [
          {
            id: "anthropic",
            available: true,
            models: ["claude-sonnet-4-6", "claude-haiku-4-5"],
            tool_capable: ["claude-sonnet-4-6"],
          },
          {
            id: "ollama",
            available: true,
            models: ["llama3.1:8b"],
            tool_capable: ["llama3.1:8b"],
          },
        ],
      }),
    });
  });

  it("changes model when provider changes", async () => {
    const changes: string[][] = [];
    function Wrapper() {
      return (
        <SWRConfig value={{ provider: () => new Map() }}>
          <ProviderPicker
            provider="anthropic"
            model="claude-sonnet-4-6"
            onChange={(p, m) => changes.push([p, m])}
          />
        </SWRConfig>
      );
    }
    render(<Wrapper />);
    await waitFor(() => screen.getByLabelText(/provider/i));
    await userEvent.selectOptions(screen.getByLabelText(/provider/i), "ollama");
    expect(changes[0]).toEqual(["ollama", "llama3.1:8b"]);
  });
});
```

- [ ] **Step 5: Run frontend tests**

Run: `npm run test`
Expected: PASS.

- [ ] **Step 6: Verify production build**

Run: `npm run build`
Expected: build succeeds; `out/` directory populated.

- [ ] **Step 7: Commit**

```bash
git add homelab-chatbot/frontend/components/settings/ homelab-chatbot/frontend/app/settings/ homelab-chatbot/frontend/app/page.tsx homelab-chatbot/frontend/tests/provider-picker.test.tsx
git commit -m "[feat:homelab-chatbot] add provider/model picker and settings page"
```

---

## Phase 7 — Deployment & finishing

### Task 29: Multi-stage Dockerfile

Build the frontend with Node, then ship only static assets + a Python runtime image. Keep the final image slim and runnable as a non-root user.

**Files:**
- Create: `homelab-chatbot/Dockerfile`
- Create: `homelab-chatbot/.dockerignore`

- [ ] **Step 1: Write `homelab-chatbot/.dockerignore`**

```
# VCS & editor
.git/
.gitignore
.vscode/
.idea/

# Python
**/__pycache__/
**/*.py[cod]
**/.venv/
**/.pytest_cache/
**/.ruff_cache/
**/.mypy_cache/
**/htmlcov/
**/.coverage

# Node
**/node_modules/
**/.next/
**/out/

# Runtime data (never bake into image)
data/
config/config.yaml
config/.env

# Tests & docs
**/tests/
**/*.md
!README.md
```

- [ ] **Step 2: Write `homelab-chatbot/Dockerfile`**

```dockerfile
# syntax=docker/dockerfile:1.7

# Stage 1: frontend build
FROM node:22-alpine AS frontend-builder
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ ./
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build
# Next.js output:"export" writes static assets to out/.

# Stage 2: backend dependency build
FROM python:3.13-slim-bookworm AS backend-deps
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 PIP_NO_CACHE_DIR=1
RUN apt-get update \
    && apt-get install -y --no-install-recommends git curl ca-certificates build-essential \
    && rm -rf /var/lib/apt/lists/*
# Install uv (Astral) for deterministic, fast installs.
COPY --from=ghcr.io/astral-sh/uv:0.4.27 /uv /usr/local/bin/uv
WORKDIR /app
COPY backend/pyproject.toml backend/uv.lock* ./
RUN uv venv /opt/venv && \
    VIRTUAL_ENV=/opt/venv uv pip install --python /opt/venv/bin/python -r pyproject.toml

# Stage 3: final runtime
FROM python:3.13-slim-bookworm
ENV PYTHONDONTWRITEBYTECODE=1 PYTHONUNBUFFERED=1 \
    PATH="/opt/venv/bin:${PATH}" \
    VIRTUAL_ENV=/opt/venv \
    HLCB_DATA_DIR=/data \
    HLCB_CONFIG_PATH=/config/config.yaml
RUN apt-get update \
    && apt-get install -y --no-install-recommends git ca-certificates tini \
    && rm -rf /var/lib/apt/lists/* \
    && useradd --system --uid 1000 --home-dir /app --shell /usr/sbin/nologin app \
    && mkdir -p /data /config /app/static \
    && chown -R app:app /data /config /app
COPY --from=backend-deps /opt/venv /opt/venv
WORKDIR /app
COPY --chown=app:app backend/ /app/
COPY --chown=app:app --from=frontend-builder /src/frontend/out /app/static
USER app
EXPOSE 8000
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
  CMD curl -fsS http://127.0.0.1:8000/api/health || exit 1
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["uvicorn", "app.main:create_app", "--factory", "--host", "0.0.0.0", "--port", "8000"]
```

- [ ] **Step 3: Build the image locally**

Run: `docker build -t adeotek/homelab-chatbot:dev homelab-chatbot/`
Expected: all 3 stages complete; final image under 1.5 GB.

- [ ] **Step 4: Smoke-run the image without config to confirm it fails loudly**

Run: `docker run --rm adeotek/homelab-chatbot:dev`
Expected: exits non-zero with a clear "config not found at /config/config.yaml" error.

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/Dockerfile homelab-chatbot/.dockerignore
git commit -m "[feat:homelab-chatbot] add multi-stage Dockerfile"
```

---

### Task 30: docker-compose + config samples

Provide a compose file that mounts the config directory and a persistent data volume, plus tracked sample configs that document every knob without leaking secrets.

**Files:**
- Create: `homelab-chatbot/docker-compose.yml`
- Create: `homelab-chatbot/config/config.example.yaml`
- Create: `homelab-chatbot/config/.env.example`

- [ ] **Step 1: Write `homelab-chatbot/docker-compose.yml`**

```yaml
services:
  chatbot:
    image: adeotek/homelab-chatbot:${HLCB_TAG:-latest}
    build:
      context: .
    container_name: homelab-chatbot
    restart: unless-stopped
    ports:
      # LAN-only: bind to host LAN IP via env, default loopback for safety.
      - "${HLCB_BIND_ADDR:-127.0.0.1}:${HLCB_PORT:-8000}:8000"
    env_file:
      - ./config/.env
    volumes:
      - ./config:/config:ro
      - hlcb-data:/data
    # Ollama runs on the host; use host.docker.internal on Linux via extra_hosts.
    extra_hosts:
      - "host.docker.internal:host-gateway"
    healthcheck:
      test: ["CMD", "curl", "-fsS", "http://127.0.0.1:8000/api/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 20s

volumes:
  hlcb-data:
```

- [ ] **Step 2: Write `homelab-chatbot/config/config.example.yaml`**

```yaml
# Copy to config.yaml and edit. Never commit config.yaml.
# Schema matches app.config.AppConfig (Task 4).

sync:
  # Polling interval in seconds for git sync.
  interval_seconds: 180
  # File on the persistent data volume tracking last commit SHA per repo.
  state_file: "/data/sync_state.json"

repos:
  - name: "homelab-docs"
    url: "https://github.com/example-org/homelab-docs.git"
    branch: "main"
    # Personal access token env var (set in .env).
    token_env: "HLCB_GIT_TOKEN_HOMELAB_DOCS"
    include_globs: ["**/*.md"]

  - name: "runbooks"
    url: "https://github.com/example-org/runbooks.git"
    branch: "main"
    token_env: "HLCB_GIT_TOKEN_RUNBOOKS"
    include_globs: ["**/*.md"]

embeddings:
  model: "BAAI/bge-small-en-v1.5"
  cache_dir: "/data/models"

vector_store:
  path: "/data/lance"

chat_db:
  path: "/data/chat.db"

kb_db:
  path: "/data/kb.db"

llm:
  default_provider: "anthropic"
  default_model: "claude-sonnet-4-6"
  ollama:
    host: "http://host.docker.internal:11434"
    # Whitelist of models known to support tool-calling. The UI shows a
    # banner for any selected Ollama model not in this list.
    tool_capable_models:
      - "llama3.1:8b"
      - "llama3.1:70b"
      - "qwen2.5:7b"
      - "qwen2.5:14b"
      - "mistral-nemo:12b"

retrieval:
  top_k: 6
  memory_turns: 10
```

- [ ] **Step 3: Write `homelab-chatbot/config/.env.example`**

```dotenv
# Copy to .env and fill in. Never commit .env.

# Session signing key. Generate with:
#   python -c "import secrets; print(secrets.token_urlsafe(48))"
HLCB_SESSION_SECRET=replace-me-with-48-bytes-of-entropy

# Shared bcrypt-hashed password (single credential for the whole household).
# Generate with:
#   python -c "import bcrypt; print(bcrypt.hashpw(b'yourpw', bcrypt.gensalt(12)).decode())"
HLCB_AUTH_PASSWORD_HASH=$2b$12$replace-with-real-bcrypt-hash

# LLM API keys (leave blank for providers you don't use).
HLCB_ANTHROPIC_API_KEY=
HLCB_GOOGLE_API_KEY=

# GitHub personal access tokens, one per repo (name-keyed, uppercased).
HLCB_GIT_TOKEN_HOMELAB_DOCS=
HLCB_GIT_TOKEN_RUNBOOKS=

# Optional: bind address for docker-compose (default 127.0.0.1 / LAN-only).
HLCB_BIND_ADDR=0.0.0.0
HLCB_PORT=8000
HLCB_TAG=latest
```

- [ ] **Step 4: Validate compose syntax**

Run: `docker compose -f homelab-chatbot/docker-compose.yml config`
Expected: prints the resolved compose config with no errors.

- [ ] **Step 5: Commit**

```bash
git add homelab-chatbot/docker-compose.yml homelab-chatbot/config/config.example.yaml homelab-chatbot/config/.env.example
git commit -m "[feat:homelab-chatbot] add docker-compose and sample configs"
```

---

### Task 31: Top-level Makefile

A single Makefile at `homelab-chatbot/` orchestrates backend + frontend workflows. Mirrors the monorepo convention (`make build`, `make test`, `make lint`) while delegating to each stack's native tooling.

**Files:**
- Create: `homelab-chatbot/Makefile`

- [ ] **Step 1: Write `homelab-chatbot/Makefile`**

```makefile
# Makefile for homelab-chatbot

IMAGE       ?= adeotek/homelab-chatbot
TAG         ?= dev
BACKEND_DIR := backend
FRONTEND_DIR:= frontend

.PHONY: help all build build-backend build-frontend docker-build \
        run dev test test-backend test-frontend test-integration \
        lint lint-backend lint-frontend fmt fmt-backend fmt-frontend \
        deps deps-backend deps-frontend clean

help:
	@echo "Targets:"
	@echo "  build           - Build frontend static assets (out/) and sync backend deps"
	@echo "  docker-build    - Build the Docker image ($(IMAGE):$(TAG))"
	@echo "  run             - docker compose up (foreground)"
	@echo "  dev             - Print dev-server launch commands"
	@echo "  test            - Run backend + frontend unit tests"
	@echo "  test-integration- Run integration tests (requires RUN_INTEGRATION_TESTS=1)"
	@echo "  lint            - Lint backend (ruff) and frontend (eslint)"
	@echo "  fmt             - Format backend (ruff format) and frontend (prettier)"
	@echo "  clean           - Remove build artifacts"

all: build

build: build-frontend build-backend

build-backend:
	cd $(BACKEND_DIR) && uv sync

build-frontend:
	cd $(FRONTEND_DIR) && npm ci && npm run build

docker-build:
	docker build -t $(IMAGE):$(TAG) .

run:
	docker compose up --build

dev:
	@echo "Start backend:  cd $(BACKEND_DIR) && uv run uvicorn app.main:create_app --factory --reload"
	@echo "Start frontend: cd $(FRONTEND_DIR) && npm run dev"

test: test-backend test-frontend

test-backend:
	cd $(BACKEND_DIR) && uv run pytest -q

test-frontend:
	cd $(FRONTEND_DIR) && npm run test -- --run

test-integration:
	cd $(BACKEND_DIR) && RUN_INTEGRATION_TESTS=1 uv run pytest -v tests/integration

lint: lint-backend lint-frontend

lint-backend:
	cd $(BACKEND_DIR) && uv run ruff check .

lint-frontend:
	cd $(FRONTEND_DIR) && npm run lint

fmt: fmt-backend fmt-frontend

fmt-backend:
	cd $(BACKEND_DIR) && uv run ruff format .

fmt-frontend:
	cd $(FRONTEND_DIR) && npm run format

deps: deps-backend deps-frontend

deps-backend:
	cd $(BACKEND_DIR) && uv sync

deps-frontend:
	cd $(FRONTEND_DIR) && npm ci

clean:
	rm -rf $(BACKEND_DIR)/.pytest_cache $(BACKEND_DIR)/.ruff_cache
	rm -rf $(FRONTEND_DIR)/.next $(FRONTEND_DIR)/out $(FRONTEND_DIR)/node_modules
```

- [ ] **Step 2: Verify make targets parse**

Run: `make -C homelab-chatbot help`
Expected: prints the help block.

- [ ] **Step 3: Commit**

```bash
git add homelab-chatbot/Makefile
git commit -m "[feat:homelab-chatbot] add top-level Makefile"
```

---

### Task 32: Integration test harness

Opt-in integration suite exercises the stitched-together backend (FastAPI + SQLite + LanceDB + local embeddings) without hitting any external LLM. Uses `httpx.AsyncClient` against an ASGI app, a tmp-path config, and a fake LLM that returns scripted tool calls so the full agent + tool pipeline is verified end-to-end.

**Files:**
- Create: `homelab-chatbot/backend/tests/integration/__init__.py`
- Create: `homelab-chatbot/backend/tests/integration/conftest.py`
- Create: `homelab-chatbot/backend/tests/integration/test_end_to_end.py`
- Create: `homelab-chatbot/backend/tests/integration/fixtures/sample_repo/networking.md`
- Create: `homelab-chatbot/backend/tests/integration/fixtures/make_xlsx.py`
- Create: `homelab-chatbot/backend/tests/integration/fixtures/sample.xlsx` (generated)

- [ ] **Step 1: Create fixtures directory and sample Markdown**

Run:
```bash
mkdir -p homelab-chatbot/backend/tests/integration/fixtures/sample_repo
```

Write `homelab-chatbot/backend/tests/integration/fixtures/sample_repo/networking.md`:

```markdown
# Homelab Networking

## VLANs
VLAN 10 is the management network. VLAN 20 is for guest devices.

## DHCP
The DHCP server runs on the OPNsense firewall and leases addresses in 10.0.10.0/24.
```

- [ ] **Step 2: Generate `sample.xlsx` with a setup script**

Write `homelab-chatbot/backend/tests/integration/fixtures/make_xlsx.py`:

```python
"""Regenerate the integration-test Excel fixture. Run manually when the schema changes."""

from pathlib import Path

import pandas as pd

HERE = Path(__file__).parent


def main() -> None:
    devices = pd.DataFrame(
        [
            {"hostname": "nas-01", "role": "storage", "cpu_cores": 8, "ram_gb": 32},
            {"hostname": "proxmox-01", "role": "hypervisor", "cpu_cores": 16, "ram_gb": 64},
            {"hostname": "pihole-01", "role": "dns", "cpu_cores": 2, "ram_gb": 4},
        ]
    )
    with pd.ExcelWriter(HERE / "sample.xlsx", engine="openpyxl") as w:
        devices.to_excel(w, sheet_name="devices", index=False)


if __name__ == "__main__":
    main()
```

Run it once to produce the binary fixture:

Run: `cd homelab-chatbot/backend && uv run python tests/integration/fixtures/make_xlsx.py`
Expected: `tests/integration/fixtures/sample.xlsx` exists. Check it in to git.

- [ ] **Step 3: Write `homelab-chatbot/backend/tests/integration/__init__.py`**

Empty file — marks the directory as a package.

- [ ] **Step 4: Write `homelab-chatbot/backend/tests/integration/conftest.py`**

Builds an `AppConfig` that matches the real shape from Task 4 (`sync`, `repos`, `embeddings`, `vector_store`, `chat_db`, `kb_db`, `llm`, `retrieval`), writes it to a temp YAML, and passes that path to `create_app(config_path=..., with_scheduler=False)`. The scheduler is disabled so tests run the orchestrator synchronously on demand.

```python
"""Integration-test fixtures. Enabled only when RUN_INTEGRATION_TESTS=1."""

from __future__ import annotations

import os
import shutil
from collections.abc import AsyncIterator
from dataclasses import dataclass
from pathlib import Path

import pytest
import pytest_asyncio
import yaml
from httpx import ASGITransport, AsyncClient

from app.main import create_app

INTEGRATION_ENABLED = os.environ.get("RUN_INTEGRATION_TESTS") == "1"

pytestmark = pytest.mark.skipif(
    not INTEGRATION_ENABLED,
    reason="set RUN_INTEGRATION_TESTS=1 to run integration tests",
)

FIXTURES = Path(__file__).parent / "fixtures"


@dataclass
class IntegrationEnv:
    config_path: Path
    data_dir: Path
    repo_path: Path
    excel_path: Path


@pytest.fixture
def env(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> IntegrationEnv:
    data_dir = tmp_path / "data"
    (data_dir / "repos").mkdir(parents=True)
    (data_dir / "excel").mkdir(parents=True)

    # Copy sample repo and Excel fixture into the temp data dir.
    repo_dst = data_dir / "repos" / "sample"
    shutil.copytree(FIXTURES / "sample_repo", repo_dst)
    excel_dst = data_dir / "excel" / "sample.xlsx"
    shutil.copy(FIXTURES / "sample.xlsx", excel_dst)

    # Write a YAML config that matches the real AppConfig schema (Task 4).
    config_body = {
        "sync": {
            "interval_seconds": 3600,  # disabled in practice; scheduler is off
            "state_file": str(data_dir / "sync_state.json"),
        },
        "repos": [
            {
                "name": "sample",
                "url": str(repo_dst),  # local path — GitSync treats it as a clone source
                "branch": "main",
                "token_env": "HLCB_GIT_TOKEN_SAMPLE",
                "include_globs": ["**/*.md"],
            }
        ],
        "embeddings": {
            "model": "BAAI/bge-small-en-v1.5",
            "cache_dir": str(data_dir / "models"),
        },
        "vector_store": {"path": str(data_dir / "lance")},
        "chat_db": {"path": str(data_dir / "chat.db")},
        "kb_db": {"path": str(data_dir / "kb.db")},
        "llm": {
            "default_provider": "anthropic",
            "default_model": "claude-sonnet-4-6",
            "ollama": {"host": "http://127.0.0.1:11434", "tool_capable_models": []},
        },
        "retrieval": {"top_k": 4, "memory_turns": 5},
    }
    config_path = tmp_path / "config.yaml"
    config_path.write_text(yaml.safe_dump(config_body))

    # Auth: set a known password hash and session secret via env (matches Task 7).
    # Password "test" — generated with: bcrypt.hashpw(b"test", bcrypt.gensalt(12))
    monkeypatch.setenv(
        "HLCB_AUTH_PASSWORD_HASH",
        "$2b$12$C6UzMDM.H6dfI/f/IKcEeu6aVm0t7n5Zz5Vp5oZ3qL9y.3aKpVQ7e",
    )
    monkeypatch.setenv("HLCB_SESSION_SECRET", "test-secret-48-bytes-of-padding-padding-padding-x")
    monkeypatch.setenv("HLCB_GIT_TOKEN_SAMPLE", "unused-for-local-path")

    return IntegrationEnv(
        config_path=config_path,
        data_dir=data_dir,
        repo_path=repo_dst,
        excel_path=excel_dst,
    )


@pytest_asyncio.fixture
async def client(env: IntegrationEnv) -> AsyncIterator[AsyncClient]:
    app = create_app(config_path=str(env.config_path), with_scheduler=False)
    # Trigger startup hooks (ChatDB.init_schema) explicitly under ASGITransport.
    async with ASGITransport(app=app) as transport:
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            yield c
```

- [ ] **Step 5: Write `homelab-chatbot/backend/tests/integration/test_end_to_end.py`**

Exercises health, auth, and end-to-end ingestion + retrieval without an external LLM. The chat-stream test monkey-patches `build_llm` to return a canned LLM that deterministically invokes `vector_search` — this verifies the agent wiring and SSE framing without needing network access or API keys.

```python
"""End-to-end smoke tests: health, auth, ingestion, and retrieval wiring."""

from __future__ import annotations

import json

import pytest
from httpx import AsyncClient

from app.config import load_config
from app.ingestion.orchestrator import IngestionOrchestrator


@pytest.mark.asyncio
async def test_health(client: AsyncClient) -> None:
    r = await client.get("/api/health")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"


@pytest.mark.asyncio
async def test_login_sets_session_cookie(client: AsyncClient) -> None:
    r = await client.post(
        "/api/auth/login",
        json={"username": "user", "password": "test"},
    )
    assert r.status_code == 200
    assert "hlcb_session" in r.cookies


@pytest.mark.asyncio
async def test_login_rejects_wrong_password(client: AsyncClient) -> None:
    r = await client.post(
        "/api/auth/login",
        json={"username": "user", "password": "nope"},
    )
    assert r.status_code == 401


def test_ingestion_populates_vector_store(env) -> None:
    """Run the orchestrator directly against the fixture repo and verify
    the vector store received chunks. Synchronous — run_once() is sync."""
    cfg = load_config(env.config_path)
    orch = IngestionOrchestrator(
        config=cfg,
        clone_root=env.data_dir / "repos",
        get_token=lambda name: None,  # local paths don't need tokens
    )
    results = orch.run_once()
    assert len(results) == 1
    # The sample repo has one markdown file — expect at least one chunk.
    count = orch.vector_store.count()
    assert count >= 1, f"expected chunks in vector store, got {count}"


@pytest.mark.asyncio
async def test_chat_stream_emits_sse_events(
    env, client: AsyncClient, monkeypatch: pytest.MonkeyPatch
) -> None:
    """End-to-end chat stream with a stub LLM. Verifies the SSE framing
    (event: <kind> / data: <json>) and that a 'done' event always fires."""
    # First populate the vector store so retrieval has something to find.
    cfg = load_config(env.config_path)
    orch = IngestionOrchestrator(
        config=cfg,
        clone_root=env.data_dir / "repos",
        get_token=lambda name: None,
    )
    orch.run_once()

    # Stub build_llm to return a fake LLM that emits a single text chunk.
    from app.llm import provider as provider_module

    class _FakeLLM:
        async def astream_chat(self, messages):  # noqa: ANN001
            async def _gen():
                yield type("Delta", (), {"delta": "VLAN 10 is management."})()

            return _gen()

    monkeypatch.setattr(provider_module, "build_llm", lambda **kw: _FakeLLM())

    # Log in first so the session cookie is attached.
    await client.post("/api/auth/login", json={"username": "user", "password": "test"})

    async with client.stream(
        "POST",
        "/api/chat",
        json={
            "provider": "anthropic",
            "model": "claude-sonnet-4-6",
            "message": "Which VLAN is management?",
        },
    ) as resp:
        assert resp.status_code == 200
        kinds: list[str] = []
        payloads: list[dict] = []
        current_kind: str | None = None
        async for line in resp.aiter_lines():
            if line.startswith("event: "):
                current_kind = line[len("event: "):].strip()
            elif line.startswith("data: ") and current_kind:
                kinds.append(current_kind)
                payloads.append(json.loads(line[len("data: "):]))
                current_kind = None

    assert "done" in kinds, f"expected a 'done' event, got {kinds}"
```

- [ ] **Step 6: Run the integration suite**

Run: `cd homelab-chatbot/backend && RUN_INTEGRATION_TESTS=1 uv run pytest -v tests/integration`
Expected: all five tests pass; without the env var they are skipped.

Note: the `HLCB_AUTH_PASSWORD_HASH` value in the fixture must be a real bcrypt hash of the string `test`. Regenerate with `python -c "import bcrypt; print(bcrypt.hashpw(b'test', bcrypt.gensalt(12)).decode())"` and paste it into `conftest.py` before running for the first time.

- [ ] **Step 7: Commit**

```bash
git add homelab-chatbot/backend/tests/integration/
git commit -m "[test:homelab-chatbot] add end-to-end integration test harness"
```

---

### Task 33: README + GitHub Actions CI

Replace the README stub with real runtime/config/deploy docs. Add two workflows: a CI workflow that lints and tests on PRs, plus a manual release workflow that builds and pushes the Docker image (matching the `sql-toolbox` monorepo convention).

**Files:**
- Modify: `homelab-chatbot/README.md` (replace the stub from Task 1)
- Create: `.github/workflows/homelab-chatbot-ci.yml`
- Create: `.github/workflows/homelab-chatbot-docker-build-push.yml`
- Modify: `README.md` (monorepo root — add a bullet under "Tools")

- [ ] **Step 1: Replace `homelab-chatbot/README.md`**

```markdown
# homelab-chatbot

Self-hosted RAG chatbot for a home-lab knowledge base. Ingests Markdown from two private GitHub repositories plus a multi-sheet Excel spreadsheet, then answers natural-language questions via **Ollama**, **Claude**, or **Gemini** (user-selectable per conversation).

- **Backend:** Python 3.13 + FastAPI + LlamaIndex + LanceDB + SQLite
- **Frontend:** Next.js 15 (static export) + Vercel AI SDK streaming UI
- **Deploy:** single Docker container on a LAN host

## Quick start

1. Copy configs and fill in secrets:
   ```bash
   cp config/config.example.yaml config/config.yaml
   cp config/.env.example config/.env
   # edit both
   ```
2. Build and run:
   ```bash
   make docker-build TAG=latest
   docker compose up -d
   ```
3. Browse to `http://<host>:8000`, log in, pick a provider + model, ask a question.

## Local development

```bash
# Backend (terminal 1)
cd backend && uv sync
uv run uvicorn app.main:create_app --factory --reload

# Frontend (terminal 2)
cd frontend && npm ci
npm run dev
```

The Next.js dev server proxies `/api/*` to `http://127.0.0.1:8000` during development.

## Configuration

- `config/config.yaml` — non-secret knobs (providers, repos, retrieval settings). See `config.example.yaml`.
- `config/.env` — API keys, session secret, bcrypt-hashed user passwords. Never commit.

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

- Git repos are polled every `ingestion.poll_seconds` seconds. Change detection uses commit SHA + file-path diff.
- Excel sheets are loaded into an embedded SQLite (`kb.db`) on first run and on file-mtime change. All queries run read-only (`PRAGMA query_only=ON`).
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
```

- [ ] **Step 2: Add bullet to monorepo `README.md`**

Edit `/home/georg/projects/adeotek-tools/README.md`, replacing:

```markdown
## Tools

- Git Repos Backup (`git-repos-backup`) [Read more...](./git-repos-backup/README.md)
```

with:

```markdown
## Tools

- Git Repos Backup (`git-repos-backup`) [Read more...](./git-repos-backup/README.md)
- SQL Toolbox (`sql-toolbox`) [Read more...](./sql-toolbox/README.md)
- Homelab Chatbot (`homelab-chatbot`) [Read more...](./homelab-chatbot/README.md)
```

- [ ] **Step 3: Write `.github/workflows/homelab-chatbot-ci.yml`**

```yaml
name: homelab-chatbot CI

on:
  pull_request:
    paths:
      - "homelab-chatbot/**"
      - ".github/workflows/homelab-chatbot-ci.yml"
  push:
    branches: [main]
    paths:
      - "homelab-chatbot/**"
      - ".github/workflows/homelab-chatbot-ci.yml"

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.13"

      - name: Install uv
        uses: astral-sh/setup-uv@v3
        with:
          version: "0.4.27"

      - name: Sync backend deps
        working-directory: homelab-chatbot/backend
        run: uv sync

      - name: Lint (ruff)
        working-directory: homelab-chatbot/backend
        run: uv run ruff check .

      - name: Unit tests
        working-directory: homelab-chatbot/backend
        run: uv run pytest -q

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Node
        uses: actions/setup-node@v4
        with:
          node-version: "22"
          cache: "npm"
          cache-dependency-path: homelab-chatbot/frontend/package-lock.json

      - name: Install
        working-directory: homelab-chatbot/frontend
        run: npm ci

      - name: Lint
        working-directory: homelab-chatbot/frontend
        run: npm run lint

      - name: Unit tests
        working-directory: homelab-chatbot/frontend
        run: npm run test -- --run

      - name: Build (static export)
        working-directory: homelab-chatbot/frontend
        run: npm run build
```

- [ ] **Step 4: Write `.github/workflows/homelab-chatbot-docker-build-push.yml`**

```yaml
name: Build and Push homelab-chatbot Docker Image

on:
  workflow_dispatch:
    inputs:
      version_tag:
        description: "Version tag for the Docker image (e.g. 0.1.0)"
        required: false
        default: ""

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKER_HUB_USERNAME }}
          password: ${{ secrets.DOCKER_HUB_PASSWORD }}

      - name: Set up Docker tags
        id: docker_tags
        run: |
          TAGS="adeotek/homelab-chatbot:latest"
          if [ -n "${{ github.event.inputs.version_tag }}" ]; then
            TAGS="$TAGS,adeotek/homelab-chatbot:${{ github.event.inputs.version_tag }}"
          fi
          echo "tags=$TAGS" >> "$GITHUB_OUTPUT"

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Build and push Docker image
        uses: docker/build-push-action@v6
        with:
          context: ./homelab-chatbot
          file: ./homelab-chatbot/Dockerfile
          push: true
          tags: ${{ steps.docker_tags.outputs.tags }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
```

- [ ] **Step 5: Validate workflow syntax locally (optional)**

Run (if `actionlint` is installed):
```bash
actionlint .github/workflows/homelab-chatbot-ci.yml .github/workflows/homelab-chatbot-docker-build-push.yml
```
Expected: no errors. If the tool isn't installed, skip — GitHub will validate on push.

- [ ] **Step 6: Commit**

```bash
git add homelab-chatbot/README.md README.md .github/workflows/homelab-chatbot-ci.yml .github/workflows/homelab-chatbot-docker-build-push.yml
git commit -m "[docs:homelab-chatbot] add README and CI/release workflows"
```

- [ ] **Step 7: Create a monorepo version tag once the first Docker build passes**

Run (only after the manual Docker build workflow succeeds on `main`):
```bash
git tag homelab-chatbot/v0.1.0
git push origin homelab-chatbot/v0.1.0
```

Then trigger the "Build and Push homelab-chatbot Docker Image" workflow manually with `version_tag=0.1.0`.

---

## Done

All phases complete. The system is buildable, testable, dockerized, and documented. See `homelab-chatbot/README.md` for runtime instructions and the design spec for architectural rationale.

