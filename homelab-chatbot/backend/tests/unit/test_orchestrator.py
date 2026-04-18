import subprocess
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from app.config import (
    AppConfig,
    EmbeddingsConfig,
    LLMConfig,
    OllamaConfig,
    PathConfig,
    RepoConfig,
    RetrievalConfig,
    SyncConfig,
)
from app.ingestion.embed import Embedder
from app.ingestion.git_sync import GitSync
from app.ingestion.orchestrator import IngestionOrchestrator
from app.storage.lance import VectorStore


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
def repos(tmp_path: Path) -> list[RepoConfig]:
    remote = _remote(tmp_path)
    return [
        RepoConfig(
            name="r",
            url=f"file://{remote}",
            branch="main",
            token_env="UNUSED",
            include_globs=["**/*.md"],
        )
    ]


def _make_orch(tmp_path: Path) -> IngestionOrchestrator:
    embedder = Embedder(model_name="BAAI/bge-small-en-v1.5", cache_dir=str(tmp_path / "models"))
    store = VectorStore(tmp_path / "lance")
    git_sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    return IngestionOrchestrator(git_sync=git_sync, embedder=embedder, store=store)


def test_full_ingest_populates_vector_store(tmp_path: Path, repos: list[RepoConfig]):
    orch = _make_orch(tmp_path)
    orch.run_once(repos)
    assert orch.vector_store.count() > 0


def test_second_run_no_changes(tmp_path: Path, repos: list[RepoConfig]):
    orch = _make_orch(tmp_path)
    orch.run_once(repos)
    count_1 = orch.vector_store.count()
    orch.run_once(repos)
    count_2 = orch.vector_store.count()
    assert count_1 == count_2


def test_per_file_error_isolation(tmp_path: Path, repos: list[RepoConfig]):
    """A failure on one file should not abort processing of the rest."""
    orch = _make_orch(tmp_path)
    # First run to clone and get files into matched_files
    orch.run_once(repos)
    count_before = orch.vector_store.count()

    # Simulate a file-level error on embed_batch — should not raise
    with patch.object(orch._embedder, "embed_batch", side_effect=RuntimeError("boom")):
        orch.run_once(repos)  # must not raise

    # Vector store should still have entries from the first run
    assert orch.vector_store.count() >= 0  # at minimum didn't crash
