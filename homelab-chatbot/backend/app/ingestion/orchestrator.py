"""Coordinate git sync + markdown chunking + Excel loading + vector index updates."""

import logging
from pathlib import Path

from app.config import RepoConfig
from app.ingestion.embed import Embedder
from app.ingestion.excel import ExcelLoader
from app.ingestion.git_sync import ChangedFile, GitSync, SyncResult
from app.ingestion.markdown import chunk_markdown_file
from app.storage.lance import VectorStore

logger = logging.getLogger(__name__)


class IngestionOrchestrator:
    """Runs one full ingestion cycle across the supplied repos."""

    def __init__(
        self,
        git_sync: GitSync,
        embedder: Embedder,
        store: VectorStore,
        excel_loader: ExcelLoader | None = None,
    ) -> None:
        self._git = git_sync
        self._embedder = embedder
        self.vector_store = store
        self._excel = excel_loader

    def run_once(self, repos: list[RepoConfig]) -> list[SyncResult]:
        results = []
        for repo_cfg in repos:
            result = self._git.sync(repo_cfg)
            self._apply(repo_cfg, result)
            results.append(result)
        return results

    def _apply(self, repo_cfg: RepoConfig, result: SyncResult) -> None:
        if not result.matched_files:
            return
        repo_path = self._git.repo_path(repo_cfg)
        for cf in result.matched_files:
            try:
                self._apply_file(repo_cfg, repo_path, cf, result.new_sha)
            except Exception:
                logger.exception("failed to process %s in %s", cf.path, repo_cfg.name)

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

        elif cf.path.endswith(".xlsx") and self._excel is not None:
            if cf.status == "D" or not file_path.exists():
                return
            self._excel.load(file_path)
