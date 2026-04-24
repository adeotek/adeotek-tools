"""Read-only SQL tool over kb.db (Excel-sourced tables)."""

import re
import sqlite3
from dataclasses import dataclass
from pathlib import Path


@dataclass
class QueryResult:
    sql: str
    rows: list[dict]


class SQLTool:
    """Executes read-only SELECT statements against kb.db."""

    TOOL_NAME = "query_homelab_inventory"

    def __init__(self, db_path: Path) -> None:
        self._db_path = Path(db_path)

    def _open(self) -> sqlite3.Connection:
        if not self._db_path.exists():
            self._db_path.parent.mkdir(parents=True, exist_ok=True)
            sqlite3.connect(str(self._db_path)).close()
        conn = sqlite3.connect(f"file:{self._db_path}?mode=ro", uri=True)
        conn.row_factory = sqlite3.Row
        return conn

    def run_select(self, sql: str) -> QueryResult:
        if not self._is_select(sql):
            raise ValueError(f"only SELECT statements allowed; got: {sql[:80]}")
        with self._open() as conn:
            cur = conn.execute(sql)
            rows = [dict(r) for r in cur.fetchall()]
        return QueryResult(sql=sql, rows=rows)

    def _is_select(self, sql: str) -> bool:
        cleaned = re.sub(r"--[^\n]*", "", sql).strip().rstrip(";").strip()
        if ";" in cleaned:
            return False
        first = cleaned.split(None, 1)[0].lower() if cleaned else ""
        if first == "select":
            return True
        if first == "with":
            return bool(re.search(r"\bselect\b", cleaned, re.IGNORECASE))
        return False

    _SYSTEM_TABLES = frozenset({"_kb_meta", "upload_log"})

    def table_stats(self) -> list[dict]:
        """Return [{table, rows}] for every user inventory table in kb.db."""
        if not self._db_path.exists():
            return []
        with self._open() as conn:
            tables = [
                r[0]
                for r in conn.execute(
                    "SELECT name FROM sqlite_master WHERE type='table' "
                    "AND name NOT LIKE 'sqlite_%'"
                )
                if r[0] not in self._SYSTEM_TABLES
            ]
            return [
                {"table": t, "rows": conn.execute(f'SELECT COUNT(*) FROM "{t}"').fetchone()[0]}
                for t in tables
            ]

    def schema_summary(self) -> str:
        """Return a schema description with sample rows for every table in kb.db.

        Includes upload_log so the LLM can answer questions about uploaded files,
        but excludes internal SQLite and metadata tables.
        """
        with self._open() as conn:
            tables = [
                r[0]
                for r in conn.execute(
                    "SELECT name FROM sqlite_master WHERE type='table' "
                    "AND name NOT LIKE 'sqlite_%' AND name != '_kb_meta'"
                )
            ]
            if not tables:
                return ""
            lines = []
            for t in tables:
                cols = [r[1] for r in conn.execute(f'PRAGMA table_info("{t}")')]
                lines.append(f"Table: {t}  Columns: {', '.join(cols)}")
                sample = conn.execute(f'SELECT * FROM "{t}" LIMIT 3').fetchall()
                if sample:
                    lines.append("  Sample rows:")
                    for row in sample:
                        lines.append(f"    {dict(row)}")
            return "\n".join(lines)

    def as_llama_tool(self, llm):
        """Return a LlamaIndex FunctionTool that uses `llm` to generate SQL."""
        from llama_index.core.tools import FunctionTool

        async def _run(question: str) -> dict:
            """Answer an inventory question by generating and executing a SELECT."""
            schema = self.schema_summary()
            if not schema:
                return {"sql": "", "rows": [], "error": "No inventory tables found. Upload an Excel file first."}
            prompt = (
                "You are a SQLite expert. Produce a single SELECT statement "
                "answering the user's question.\n\n"
                "Available tables (with sample rows showing actual value formats):\n\n"
                f"{schema}\n\n"
                "Rules:\n"
                "- Use ONLY the tables and columns listed above.\n"
                "- Match the exact column names shown.\n"
                "- If unsure of the exact value, use LIKE with wildcards.\n"
                "- Return ONLY the SQL statement, no prose, no markdown fences.\n\n"
                f"Question: {question}"
            )
            response = await llm.acomplete(prompt)
            sql = str(response).strip().strip("`")
            if sql.lower().startswith("sql"):
                sql = sql[3:].strip()
            try:
                result = self.run_select(sql)
            except Exception as e:  # noqa: BLE001
                return {"sql": sql, "error": str(e), "rows": [], "schema": schema}
            if not result.rows:
                return {
                    "sql": result.sql,
                    "rows": [],
                    "hint": (
                        "Query executed but returned no rows. "
                        "The SQL above may use wrong column names or an exact value that "
                        "doesn't match the stored data. "
                        "Try again using LIKE with wildcards (e.g. WHERE col LIKE '%value%'). "
                        "Available tables and sample data:\n" + schema
                    ),
                }
            return {"sql": result.sql, "rows": result.rows}

        return FunctionTool.from_defaults(
            async_fn=_run,
            name=self.TOOL_NAME,
            description=(
                "Answer questions about the homelab inventory (devices, services, etc.) "
                "using structured SQL queries. Use for counting, filtering, aggregating "
                "or looking up specific items in tabular data."
            ),
        )
