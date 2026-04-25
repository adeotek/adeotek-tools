from unittest.mock import patch

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.auth import AuthService, hash_password
from app.config import (
    AppConfig,
    EmbeddingsConfig,
    LLMConfig,
    OllamaConfig,
    PathConfig,
    RetrievalConfig,
    SyncConfig,
)
from app.routes import auth as auth_routes
from app.routes import settings as settings_routes


def _make_config() -> AppConfig:
    return AppConfig(
        sync=SyncConfig(interval_seconds=180, state_file="/tmp/s.json"),
        repos=[],
        embeddings=EmbeddingsConfig(model="m", cache_dir="/tmp"),
        vector_store=PathConfig(path="/tmp"),
        chat_db=PathConfig(path="/tmp/c.db"),
        kb_db=PathConfig(path="/tmp/k.db"),
        llm=LLMConfig(
            default_provider="anthropic",
            default_model="claude-sonnet-4-6",
            ollama=OllamaConfig(
                host="http://mock-ollama:11434",
                tool_capable_models=["llama3.1"],
            ),
        ),
        retrieval=RetrievalConfig(top_k=5, memory_turns=10),
    )


@pytest.fixture
def client():
    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    app.state.config = _make_config()
    app.state.secrets_available = {"anthropic": True, "google": False}
    app.include_router(auth_routes.router)
    app.include_router(settings_routes.router)
    client = TestClient(app)
    client.post("/api/auth/login", json={"password": "p"})
    return client


def test_settings_lists_providers_with_models(client: TestClient):
    with patch("app.llm.registry._fetch_ollama_models", return_value=["llama3.1:8b"]):
        r = client.get("/api/settings")
    assert r.status_code == 200
    data = r.json()
    assert data["default_provider"] == "anthropic"
    providers = {p["id"]: p for p in data["providers"]}
    assert "anthropic" in providers
    assert providers["anthropic"]["available"] is True
    assert providers["google"]["available"] is False
    assert "claude-sonnet-4-6" in providers["anthropic"]["models"]
    assert "llama3.1:8b" in providers["ollama"]["models"]


def test_settings_requires_auth():
    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    app.state.config = _make_config()
    app.state.secrets_available = {"anthropic": False, "google": False}
    app.include_router(settings_routes.router)
    client = TestClient(app)
    assert client.get("/api/settings").status_code == 401
