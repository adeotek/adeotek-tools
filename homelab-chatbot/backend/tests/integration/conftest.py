"""Integration-test fixtures. Enabled only when RUN_INTEGRATION_TESTS=1."""

from __future__ import annotations

import os
import shutil
import subprocess
from collections.abc import AsyncIterator
from dataclasses import dataclass
from pathlib import Path

import pytest
import pytest_asyncio
import yaml
from httpx import ASGITransport, AsyncClient

from app.config import load_config
from app.ingestion.embed import Embedder
from app.ingestion.git_sync import GitSync
from app.ingestion.orchestrator import IngestionOrchestrator
from app.main import create_app
from app.storage.lance import VectorStore

INTEGRATION_ENABLED = os.environ.get("RUN_INTEGRATION_TESTS") == "1"

pytestmark = pytest.mark.skipif(
    not INTEGRATION_ENABLED,
    reason="set RUN_INTEGRATION_TESTS=1 to run integration tests",
)

FIXTURES = Path(__file__).parent / "fixtures"

# bcrypt hash of "test", rounds=12
_TEST_PASSWORD_HASH = "$2b$12$peKySxDbGpoq1s52XG3oCOMGooA5HrWbZ6BNG6L0FP0cZeM6Pi8dS"


@dataclass
class IntegrationEnv:
    config_path: Path
    data_dir: Path
    repo_src: Path
    excel_path: Path
    chunk_count: int


def _init_git_repo(path: Path) -> None:
    """Turn a plain directory into a bare-minimum git repo so GitSync can clone it."""
    subprocess.run(["git", "init", "-b", "main", str(path)], check=True, capture_output=True)
    subprocess.run(
        ["git", "-C", str(path), "config", "user.email", "test@test.com"],
        check=True, capture_output=True,
    )
    subprocess.run(
        ["git", "-C", str(path), "config", "user.name", "Test"],
        check=True, capture_output=True,
    )
    subprocess.run(["git", "-C", str(path), "add", "."], check=True, capture_output=True)
    subprocess.run(
        ["git", "-C", str(path), "commit", "-m", "init"],
        check=True, capture_output=True,
    )


@pytest.fixture
def env(tmp_path: Path, monkeypatch: pytest.MonkeyPatch) -> IntegrationEnv:
    data_dir = tmp_path / "data"
    (data_dir / "repos").mkdir(parents=True)

    repo_src = tmp_path / "sample_repo_src"
    shutil.copytree(FIXTURES / "sample_repo", repo_src)
    _init_git_repo(repo_src)

    excel_dst = data_dir / "sample.xlsx"
    shutil.copy(FIXTURES / "sample.xlsx", excel_dst)

    config_body = {
        "sync": {
            "interval_seconds": 3600,
            "state_file": str(data_dir / "sync_state.json"),
        },
        "repos": [
            {
                "name": "sample",
                "url": str(repo_src),
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

    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", _TEST_PASSWORD_HASH)
    monkeypatch.setenv("HLCB_SESSION_SECRET", "test-secret-48-bytes-of-padding-padding-padding-x")
    monkeypatch.setenv("HLCB_GIT_TOKEN_SAMPLE", "unused-for-local-path")

    # Create empty kb.db so SQLTool can open it read-only without errors.
    import sqlite3
    sqlite3.connect(str(data_dir / "kb.db")).close()

    # Run ingestion once up-front so the lance table is pre-populated before
    # create_app() opens its own VectorStore on the same path.
    cfg = load_config(config_path)
    git_sync = GitSync(clone_root=data_dir / "repos", get_token=lambda name: None)
    embedder = Embedder(model_name=cfg.embeddings.model, cache_dir=str(data_dir / "models"))
    store = VectorStore(data_dir / "lance")
    orch = IngestionOrchestrator(git_sync=git_sync, embedder=embedder, store=store)
    orch.run_once(cfg.repos)
    chunk_count = store.count()

    return IntegrationEnv(
        config_path=config_path,
        data_dir=data_dir,
        repo_src=repo_src,
        excel_path=excel_dst,
        chunk_count=chunk_count,
    )


@pytest_asyncio.fixture
async def client(env: IntegrationEnv) -> AsyncIterator[AsyncClient]:
    app = create_app(config_path=str(env.config_path), with_scheduler=False)
    # ASGITransport doesn't trigger lifespan events; initialize the DB schema manually.
    await app.state.chat_db.init_schema()
    async with ASGITransport(app=app) as transport:
        async with AsyncClient(transport=transport, base_url="http://test") as c:
            yield c
        await app.state.chat_db.close()
