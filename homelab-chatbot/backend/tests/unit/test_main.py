from unittest.mock import MagicMock, patch

import pytest
from fastapi.testclient import TestClient

from app.config import (
    AppConfig, EmbeddingsConfig, LLMConfig, OllamaConfig,
    PathConfig, RetrievalConfig, SyncConfig,
)
from app.main import create_app


@pytest.fixture(autouse=True)
def _set_required_env(monkeypatch):
    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", "$2b$12$placeholder_hash_for_tests_only___x")
    monkeypatch.setenv("HLCB_SESSION_SECRET", "test-session-secret-32-chars-long!")


@pytest.fixture
def test_config(tmp_path):
    return AppConfig(
        sync=SyncConfig(interval_seconds=60, state_file=str(tmp_path / "s.json")),
        repos=[],
        embeddings=EmbeddingsConfig(model="m", cache_dir=str(tmp_path / "models")),
        vector_store=PathConfig(path=str(tmp_path / "lance")),
        chat_db=PathConfig(path=str(tmp_path / "chat.db")),
        kb_db=PathConfig(path=str(tmp_path / "kb.db")),
        llm=LLMConfig(
            default_provider="ollama",
            default_model="phi3",
            ollama=OllamaConfig(host="http://localhost:11434", tool_capable_models=[]),
        ),
        retrieval=RetrievalConfig(top_k=5, memory_turns=10),
    )


def test_app_factory_returns_fastapi_instance(test_config):
    with patch("app.main.Embedder", return_value=MagicMock()):
        app = create_app(with_scheduler=False, config=test_config)
    assert app.title == "homelab-chatbot"


def test_health_endpoint_exists(test_config):
    with patch("app.main.Embedder", return_value=MagicMock()):
        app = create_app(with_scheduler=False, config=test_config)
    client = TestClient(app)
    response = client.get("/api/health")
    assert response.status_code == 200
    assert response.json()["status"] == "ok"
