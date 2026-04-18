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
