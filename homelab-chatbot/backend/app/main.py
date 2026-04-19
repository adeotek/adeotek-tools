"""FastAPI application factory wiring together all dependencies."""

import os
from pathlib import Path

from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

from app.auth import AuthService
from app.config import AppConfig, load_config
from app.ingestion.embed import Embedder
from app.ingestion.excel import ExcelLoader
from app.ingestion.git_sync import GitSync
from app.ingestion.orchestrator import IngestionOrchestrator
from app.ingestion.scheduler import SyncScheduler
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


def create_app(
    config_path: str | None = None,
    with_scheduler: bool = True,
    config: AppConfig | None = None,
) -> FastAPI:
    app = FastAPI(title="homelab-chatbot")

    if config is not None:
        cfg = config
    else:
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
    embedder = Embedder(model_name=cfg.embeddings.model, cache_dir=cfg.embeddings.cache_dir)
    app.state.vector_tool = VectorSearchTool(
        store=vector_store, embedder=embedder, top_k=cfg.retrieval.top_k
    )
    app.state.sql_tool = SQLTool(db_path=Path(cfg.kb_db.path))
    app.state.sync_state: dict = {"last_sync_at": None}

    scheduler: SyncScheduler | None = None

    @app.on_event("startup")
    async def _startup() -> None:
        await chat_db.init_schema()
        nonlocal scheduler
        if with_scheduler and cfg.repos:
            git_sync = GitSync(
                clone_root=Path("/data/repos"),
                get_token=secrets.github_token,
            )
            orchestrator = IngestionOrchestrator(
                git_sync=git_sync,
                embedder=embedder,
                store=vector_store,
                excel_loader=ExcelLoader(Path(cfg.kb_db.path)),
            )
            scheduler = SyncScheduler(
                orchestrator=orchestrator,
                interval_seconds=cfg.sync.interval_seconds,
            )
            scheduler.start(cfg.repos)

    @app.on_event("shutdown")
    async def _shutdown() -> None:
        if scheduler:
            scheduler.stop()
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
    from app.config import (
        EmbeddingsConfig, LLMConfig, OllamaConfig,
        PathConfig, RetrievalConfig, SyncConfig,
    )
    return AppConfig(
        sync=SyncConfig(interval_seconds=180, state_file="/data/sync_state.json"),
        repos=[],
        embeddings=EmbeddingsConfig(model="BAAI/bge-small-en-v1.5", cache_dir="/data/models"),
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


try:
    app = create_app()
except Exception:  # noqa: BLE001 — allow startup without env vars (e.g. import-time in tests)
    from fastapi import FastAPI as _FastAPI
    app = _FastAPI(title="homelab-chatbot")
