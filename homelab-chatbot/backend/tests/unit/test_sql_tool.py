import sqlite3
from pathlib import Path

import pandas as pd
import pytest

from app.retrieval.sql_tool import SQLTool


@pytest.fixture
def kb_db(tmp_path: Path) -> Path:
    path = tmp_path / "kb.db"
    with sqlite3.connect(path) as conn:
        pd.DataFrame(
            [
                {"device_name": "nas-01", "os_name": "Debian", "ram_gb": 32},
                {"device_name": "router-01", "os_name": "OpenWrt", "ram_gb": 1},
                {"device_name": "server-01", "os_name": "Debian", "ram_gb": 64},
            ]
        ).to_sql("devices", conn, index=False)
    return path


def test_run_select_returns_rows(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    result = tool.run_select("SELECT device_name FROM devices WHERE os_name = 'Debian'")
    names = [r["device_name"] for r in result.rows]
    assert set(names) == {"nas-01", "server-01"}


def test_run_select_blocks_writes(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    with pytest.raises(Exception):
        tool.run_select("DELETE FROM devices")


def test_aggregate_count(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    result = tool.run_select(
        "SELECT COUNT(*) AS n FROM devices WHERE os_name = 'Debian'"
    )
    assert result.rows[0]["n"] == 2


def test_schema_summary_lists_tables(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    summary = tool.schema_summary()
    assert "devices" in summary
    assert "device_name" in summary


def test_reject_non_select_statement(kb_db: Path):
    tool = SQLTool(db_path=kb_db)
    with pytest.raises(ValueError):
        tool.run_select("UPDATE devices SET os_name = 'Arch'")
