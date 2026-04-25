from pathlib import Path

import pytest

from app.config import AppConfig, load_config

FIXTURE_DIR = Path(__file__).parent.parent / "fixtures"


def test_load_config_from_yaml():
    cfg = load_config(FIXTURE_DIR / "config_valid.yaml")
    assert isinstance(cfg, AppConfig)
    assert cfg.sync.interval_seconds == 180
    assert len(cfg.repos) == 1
    assert cfg.repos[0].name == "homelab-docs"
    assert cfg.llm.default_provider == "anthropic"
    assert "llama3.1" in cfg.llm.ollama.tool_capable_models


def test_load_config_missing_file_raises():
    with pytest.raises(FileNotFoundError):
        load_config(Path("/nonexistent/config.yaml"))


def test_invalid_provider_rejected(tmp_path: Path):
    bad = tmp_path / "bad.yaml"
    bad.write_text(
        "sync: {interval_seconds: 60, state_file: /tmp/s.json}\n"
        "repos: []\n"
        "embeddings: {model: m, cache_dir: /tmp}\n"
        "vector_store: {path: /tmp}\n"
        "chat_db: {path: /tmp/c.db}\n"
        "kb_db: {path: /tmp/k.db}\n"
        "llm:\n"
        "  default_provider: invalid_provider\n"
        "  default_model: x\n"
        "  ollama: {host: http://localhost:11434, tool_capable_models: []}\n"
        "retrieval: {top_k: 5, memory_turns: 10}\n"
    )
    with pytest.raises(ValueError):
        load_config(bad)
