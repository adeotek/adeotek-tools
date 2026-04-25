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
