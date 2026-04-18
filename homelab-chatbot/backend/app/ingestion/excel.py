"""Load Excel sheets into a SQLite database as per-sheet tables."""

import hashlib
import re
import sqlite3
from datetime import datetime, timezone
from pathlib import Path

import pandas as pd

META_TABLE = "_kb_meta"


def normalize_snake_case(name: str) -> str:
    s = name.strip().lower()
    s = re.sub(r"[^\w]+", "_", s)
    s = re.sub(r"_+", "_", s).strip("_")
    if not s:
        s = "col"
    if s[0].isdigit():
        s = "_" + s
    return s


class ExcelLoader:
    """Rebuilds SQLite tables from an Excel workbook."""

    def __init__(self, db_path: Path) -> None:
        self._db_path = Path(db_path)

    def _source_hash(self, path: Path) -> str:
        h = hashlib.sha256()
        h.update(path.read_bytes())
        return h.hexdigest()

    def load(self, xlsx_path: Path) -> dict[str, str]:
        """Replace tables and return a {sheet -> table_name} map."""
        xlsx = pd.ExcelFile(xlsx_path)
        sheet_map: dict[str, str] = {}

        with sqlite3.connect(self._db_path) as conn:
            conn.execute(
                f"CREATE TABLE IF NOT EXISTS {META_TABLE} "
                "(sheet TEXT PRIMARY KEY, table_name TEXT, columns_json TEXT, "
                "source_hash TEXT, last_rebuilt_at TEXT)"
            )

            for sheet in xlsx.sheet_names:
                df = xlsx.parse(sheet)
                df.columns = [normalize_snake_case(c) for c in df.columns]
                table = normalize_snake_case(sheet)
                df.to_sql(table, conn, if_exists="replace", index=False)
                sheet_map[sheet] = table
                conn.execute(
                    f"INSERT OR REPLACE INTO {META_TABLE} "
                    "(sheet, table_name, columns_json, source_hash, last_rebuilt_at) "
                    "VALUES (?, ?, ?, ?, ?)",
                    (
                        sheet,
                        table,
                        ",".join(df.columns),
                        self._source_hash(xlsx_path),
                        datetime.now(timezone.utc).isoformat(),
                    ),
                )
            conn.commit()
        return sheet_map
