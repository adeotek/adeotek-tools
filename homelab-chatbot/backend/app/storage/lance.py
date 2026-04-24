"""Thin LanceDB wrapper for the markdown_chunks table."""

from dataclasses import dataclass
from pathlib import Path

import lancedb
import pyarrow as pa

from app.ingestion.markdown import Chunk

TABLE_NAME = "markdown_chunks"
EMBED_DIM = 384


@dataclass
class SearchHit:
    id: str
    text: str
    repo: str
    file_path: str
    line_start: int
    line_end: int
    commit_sha: str
    heading_path: str
    score: float


class VectorStore:
    """Wraps a single LanceDB table for markdown chunks."""

    def __init__(self, path: Path) -> None:
        self._db = lancedb.connect(str(path))
        self._schema = pa.schema(
            [
                pa.field("id", pa.string()),
                pa.field("vector", pa.list_(pa.float32(), EMBED_DIM)),
                pa.field("text", pa.string()),
                pa.field("repo", pa.string()),
                pa.field("file_path", pa.string()),
                pa.field("line_start", pa.int32()),
                pa.field("line_end", pa.int32()),
                pa.field("commit_sha", pa.string()),
                pa.field("heading_path", pa.string()),
            ]
        )
        if TABLE_NAME not in self._db.list_tables().tables:
            self._db.create_table(TABLE_NAME, schema=self._schema)

    def _table(self) -> lancedb.table.Table:
        return self._db.open_table(TABLE_NAME)

    def upsert(self, chunks: list[Chunk], vectors: list[list[float]]) -> None:
        if len(chunks) != len(vectors):
            raise ValueError("chunks and vectors must have equal length")
        if not chunks:
            return
        rows = [
            {
                "id": c.id,
                "vector": v,
                "text": c.text,
                "repo": c.repo,
                "file_path": c.file_path,
                "line_start": c.line_start,
                "line_end": c.line_end,
                "commit_sha": c.commit_sha,
                "heading_path": c.heading_path,
            }
            for c, v in zip(chunks, vectors, strict=True)
        ]
        ids = [c.id for c in chunks]
        tbl = self._table()
        id_list = ", ".join(f"'{i}'" for i in ids)
        tbl.delete(f"id IN ({id_list})")
        tbl.add(rows)

    def has_file(self, repo: str, file_path: str) -> bool:
        """Return True if any chunks exist for the given repo + file_path."""
        tbl = self._table()
        if tbl.count_rows() == 0:
            return False
        df = tbl.to_pandas()[["repo", "file_path"]]
        return bool(((df["repo"] == repo) & (df["file_path"] == file_path)).any())

    def delete_by_file(self, repo: str, file_path: str) -> None:
        tbl = self._table()
        tbl.delete(f"repo = '{repo}' AND file_path = '{file_path}'")

    def count(self) -> int:
        return self._table().count_rows()

    def stats_by_repo(self) -> dict[str, dict]:
        """Return {repo: {chunks, files}} for all indexed repos."""
        tbl = self._table()
        if tbl.count_rows() == 0:
            return {}
        df = tbl.to_pandas()[["repo", "file_path"]]
        g = df.groupby("repo").agg(chunks=("file_path", "count"), files=("file_path", "nunique"))
        return {
            idx: {"chunks": int(row["chunks"]), "files": int(row["files"])}
            for idx, row in g.iterrows()
        }

    def search(
        self, query_vector: list[float], top_k: int = 5, repo: str | None = None
    ) -> list[SearchHit]:
        tbl = self._table()
        q = tbl.search(query_vector, vector_column_name="vector").limit(top_k)
        if repo:
            q = q.where(f"repo = '{repo}'", prefilter=True)
        results = q.to_list()
        return [
            SearchHit(
                id=r["id"],
                text=r["text"],
                repo=r["repo"],
                file_path=r["file_path"],
                line_start=r["line_start"],
                line_end=r["line_end"],
                commit_sha=r["commit_sha"],
                heading_path=r["heading_path"],
                score=float(r.get("_distance", 0.0)),
            )
            for r in results
        ]
