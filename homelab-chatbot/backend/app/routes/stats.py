"""Statistics endpoint."""

import logging
from datetime import datetime, timezone

from fastapi import APIRouter, BackgroundTasks, Depends, HTTPException, Request
from pydantic import BaseModel

from app.config import AppConfig, RepoConfig
from app.deps import require_session
from app.ingestion.orchestrator import IngestionOrchestrator
from app.retrieval.sql_tool import SQLTool
from app.retrieval.vector_tool import VectorSearchTool
from app.storage.chat_db import ChatDB

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/stats", tags=["stats"], dependencies=[Depends(require_session)])


class RepoStats(BaseModel):
    name: str
    url: str
    branch: str
    include_globs: list[str]
    chunks: int = 0
    files: int = 0


class KbTableStats(BaseModel):
    table: str
    rows: int


class StatsOut(BaseModel):
    last_sync_at: str | None
    is_syncing: bool
    total_chunks: int
    repos: list[RepoStats]
    kb_tables: list[KbTableStats]
    conversations: int
    messages: int


@router.get("")
async def get_stats(request: Request) -> StatsOut:
    cfg: AppConfig = request.app.state.config
    vector_tool: VectorSearchTool = request.app.state.vector_tool
    sql_tool: SQLTool = request.app.state.sql_tool
    chat_db: ChatDB = request.app.state.chat_db
    sync_state: dict = getattr(request.app.state, "sync_state", {})

    by_repo = vector_tool._store.stats_by_repo()
    total_chunks = sum(v["chunks"] for v in by_repo.values())

    repos = [
        RepoStats(
            name=r.name,
            url=r.url,
            branch=r.branch,
            include_globs=r.include_globs,
            chunks=by_repo.get(r.name, {}).get("chunks", 0),
            files=by_repo.get(r.name, {}).get("files", 0),
        )
        for r in cfg.repos
    ]

    kb_tables = [KbTableStats(**t) for t in sql_tool.table_stats()]
    chat_counts = await chat_db.count_stats()

    return StatsOut(
        last_sync_at=sync_state.get("last_sync_at"),
        is_syncing=bool(sync_state.get("is_syncing")),
        total_chunks=total_chunks,
        repos=repos,
        kb_tables=kb_tables,
        conversations=chat_counts["conversations"],
        messages=chat_counts["messages"],
    )


def _do_sync(
    sync_state: dict,
    orchestrator: IngestionOrchestrator,
    repos: list[RepoConfig],
) -> None:
    sync_state["is_syncing"] = True
    try:
        orchestrator.run_once(repos)
        sync_state["last_sync_at"] = datetime.now(timezone.utc).isoformat()
    except Exception:
        logger.exception("manual sync failed")
    finally:
        sync_state["is_syncing"] = False


@router.post("/sync", status_code=202)
async def trigger_sync(request: Request, background_tasks: BackgroundTasks) -> dict:
    sync_state: dict = getattr(request.app.state, "sync_state", {})
    if sync_state.get("is_syncing"):
        raise HTTPException(status_code=409, detail="sync already in progress")

    orchestrator: IngestionOrchestrator | None = getattr(request.app.state, "orchestrator", None)
    cfg: AppConfig = request.app.state.config
    if not orchestrator or not cfg.repos:
        raise HTTPException(status_code=400, detail="no repos configured")

    background_tasks.add_task(_do_sync, sync_state, orchestrator, cfg.repos)
    return {"ok": True}
