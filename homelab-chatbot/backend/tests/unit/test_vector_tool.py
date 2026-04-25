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
