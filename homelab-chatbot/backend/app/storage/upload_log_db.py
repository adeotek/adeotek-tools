"""Synchronous sqlite3 wrapper for the upload_log table stored in kb.db.

Placed in kb.db (not chat.db) so the existing SQLTool can query it and the
chatbot can answer questions like "what files were uploaded last week".
"""

import sqlite3
from datetime import datetime, timezone
from pathlib import Path


_CREATE = """
CREATE TABLE IF NOT EXISTS upload_log (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    filename       TEXT    NOT NULL,
    file_size      INTEGER NOT NULL,
    mime_type      TEXT    NOT NULL,
    uploaded_at    TEXT    NOT NULL,
    status         TEXT    NOT NULL,
    chunks_created INTEGER NOT NULL DEFAULT 0,
    replaced       INTEGER NOT NULL DEFAULT 0,
    error_message  TEXT
)
"""


def _utcnow() -> str:
    return datetime.now(timezone.utc).isoformat()


class UploadLogDB:
    """Manages the upload_log table in kb.db."""

    def __init__(self, db_path: Path) -> None:
        self._path = str(db_path)
        with self._connect() as conn:
            conn.execute(_CREATE)

    def _connect(self) -> sqlite3.Connection:
        conn = sqlite3.connect(self._path)
        conn.row_factory = sqlite3.Row
        return conn

    def log_upload(
        self,
        *,
        filename: str,
        file_size: int,
        mime_type: str,
        status: str,
        chunks_created: int,
        replaced: bool,
        error_message: str | None,
    ) -> int:
        with self._connect() as conn:
            cur = conn.execute(
                """
                INSERT INTO upload_log
                    (filename, file_size, mime_type, uploaded_at,
                     status, chunks_created, replaced, error_message)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    filename,
                    file_size,
                    mime_type,
                    _utcnow(),
                    status,
                    chunks_created,
                    int(replaced),
                    error_message,
                ),
            )
            return cur.lastrowid  # type: ignore[return-value]

    def list_upload_log(self, limit: int = 50) -> list[dict]:
        with self._connect() as conn:
            rows = conn.execute(
                """
                SELECT id, filename, file_size, mime_type, uploaded_at,
                       status, chunks_created, replaced, error_message
                FROM upload_log
                ORDER BY uploaded_at DESC
                LIMIT ?
                """,
                (limit,),
            ).fetchall()
        return [dict(r) for r in rows]

    def count_upload_stats(self) -> dict:
        with self._connect() as conn:
            row = conn.execute(
                """
                SELECT COUNT(*) AS uploaded_files,
                       COALESCE(SUM(chunks_created), 0) AS uploaded_chunks
                FROM upload_log
                WHERE status = 'ok'
                """,
            ).fetchone()
        return {"uploaded_files": row["uploaded_files"], "uploaded_chunks": row["uploaded_chunks"]}
