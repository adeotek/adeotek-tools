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
