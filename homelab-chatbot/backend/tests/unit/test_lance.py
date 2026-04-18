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
