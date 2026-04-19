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

    def schema_summary(self) -> str:
        with self._open() as conn:
            tables = [
                r[0]
                for r in conn.execute(
                    "SELECT name FROM sqlite_master WHERE type='table' "
                    "AND name NOT LIKE 'sqlite_%' AND name != '_kb_meta'"
                )
            ]
            lines = []
            for t in tables:
                cols = [r[1] for r in conn.execute(f'PRAGMA table_info("{t}")')]
                lines.append(f"{t}({', '.join(cols)})")
            return "\n".join(lines)

    def as_llama_tool(self, llm):
        """Return a LlamaIndex FunctionTool that uses `llm` to generate SQL."""
        from llama_index.core.tools import FunctionTool

        schema = self.schema_summary()

        async def _run(question: str) -> dict:
            """Answer an inventory question by generating and executing a SELECT."""
            prompt = (
                "You are a SQLite expert. Produce a single SELECT statement "
                "answering the user's question. Use only these tables:\n\n"
                f"{schema}\n\n"
                "Return ONLY the SQL, no prose, no fenced block.\n\n"
                f"Question: {question}"
            )
            response = await llm.acomplete(prompt)
            sql = str(response).strip().strip("`")
            if sql.lower().startswith("sql"):
                sql = sql[3:].strip()
            try:
                result = self.run_select(sql)
            except Exception as e:  # noqa: BLE001
                return {"sql": sql, "error": str(e), "rows": []}
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
