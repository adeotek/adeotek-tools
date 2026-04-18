import sqlite3
from pathlib import Path

import pandas as pd
import pytest

from app.ingestion.excel import ExcelLoader


@pytest.fixture
def xlsx(tmp_path: Path) -> Path:
    path = tmp_path / "inventory.xlsx"
    with pd.ExcelWriter(path) as w:
        pd.DataFrame(
            [
                {"Device Name": "nas-01", "OS Name": "Debian", "RAM GB": 32},
                {"Device Name": "router-01", "OS Name": "OpenWrt", "RAM GB": 1},
            ]
        ).to_excel(w, sheet_name="Devices", index=False)
        pd.DataFrame(
            [{"Service": "plex", "Port": 32400}]
        ).to_excel(w, sheet_name="Services", index=False)
    return path


def test_loader_creates_snake_case_tables(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    conn = sqlite3.connect(db_path)
    tables = [
        r[0]
        for r in conn.execute(
            "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
        )
    ]
    assert "devices" in tables
    assert "services" in tables


def test_loader_normalizes_column_names(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    conn = sqlite3.connect(db_path)
    cols = [r[1] for r in conn.execute("PRAGMA table_info(devices)")]
    assert "device_name" in cols
    assert "os_name" in cols
    assert "ram_gb" in cols


def test_loader_rewrites_on_reload(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    loader.load(xlsx)  # second call must not duplicate
    conn = sqlite3.connect(db_path)
    count = conn.execute("SELECT COUNT(*) FROM devices").fetchone()[0]
    assert count == 2


def test_loader_records_meta(tmp_path: Path, xlsx: Path):
    db_path = tmp_path / "kb.db"
    loader = ExcelLoader(db_path)
    loader.load(xlsx)
    conn = sqlite3.connect(db_path)
    rows = conn.execute(
        "SELECT sheet, table_name FROM _kb_meta ORDER BY sheet"
    ).fetchall()
    assert ("Devices", "devices") in rows
    assert ("Services", "services") in rows
