from pathlib import Path

from app.ingestion.markdown import Chunk, chunk_markdown_file

FIX = Path(__file__).parent.parent / "fixtures"


def test_chunk_produces_chunks_with_metadata():
    chunks = chunk_markdown_file(
        FIX / "sample.md",
        repo="homelab-docs",
        file_path="docs/sample.md",
        commit_sha="abc123",
    )
    assert len(chunks) > 0
    assert all(isinstance(c, Chunk) for c in chunks)
    for c in chunks:
        assert c.repo == "homelab-docs"
        assert c.file_path == "docs/sample.md"
        assert c.commit_sha == "abc123"
        assert c.line_start >= 1
        assert c.line_end >= c.line_start


def test_chunk_preserves_heading_path():
    chunks = chunk_markdown_file(
        FIX / "sample.md",
        repo="r",
        file_path="s.md",
        commit_sha="x",
    )
    headings = [c.heading_path for c in chunks]
    assert any("Networking" in h for h in headings)
    assert any("VLAN" in h for h in headings)


def test_chunk_text_non_empty():
    chunks = chunk_markdown_file(
        FIX / "sample.md",
        repo="r",
        file_path="s.md",
        commit_sha="x",
    )
    assert all(c.text.strip() for c in chunks)


def test_chunk_id_stable_across_runs(tmp_path: Path):
    (tmp_path / "a.md").write_text("# T\n\nHello world.\n")
    c1 = chunk_markdown_file(tmp_path / "a.md", repo="r", file_path="a.md", commit_sha="x")
    c2 = chunk_markdown_file(tmp_path / "a.md", repo="r", file_path="a.md", commit_sha="x")
    assert [c.id for c in c1] == [c.id for c in c2]
