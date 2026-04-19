from pathlib import Path
from typing import AsyncIterator

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.auth import AuthService, hash_password
from app.config import (
    AppConfig, EmbeddingsConfig, LLMConfig, OllamaConfig,
    PathConfig, RetrievalConfig, SyncConfig,
)
from app.retrieval.sql_tool import SQLTool
from app.retrieval.vector_tool import VectorSearchTool
from app.routes import auth as auth_routes
from app.routes import chat as chat_routes
from app.routes import conversations as conv_routes
from app.storage.chat_db import ChatDB

try:
    from llama_index.core.base.llms.types import ChatMessage, ChatResponse
except ImportError:
    from llama_index.core.llms import ChatMessage, ChatResponse


class StubEmbedder:
    def embed_batch(self, texts):
        return [[0.0] * 384 for _ in texts]


class StubVectorStore:
    def search(self, **kwargs):
        return []

    def count(self):
        return 0


class StubLLM:
    async def astream_chat(self, messages) -> AsyncIterator[ChatResponse]:
        for ch in "hi":
            yield ChatResponse(
                message=ChatMessage(role="assistant", content=ch),
                delta=ch,
                raw={},
            )


@pytest.fixture
async def client(tmp_path: Path, monkeypatch):
    monkeypatch.setattr("app.routes.chat.build_llm", lambda **kwargs: StubLLM())

    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/c.db")
    await db.init_schema()
    app.state.chat_db = db
    app.state.vector_tool = VectorSearchTool(
        store=StubVectorStore(), embedder=StubEmbedder(), top_k=1
    )
    app.state.sql_tool = SQLTool(db_path=tmp_path / "kb.db")
    app.state.config = AppConfig(
        sync=SyncConfig(interval_seconds=60, state_file=str(tmp_path / "s.json")),
        repos=[],
        embeddings=EmbeddingsConfig(model="m", cache_dir=str(tmp_path)),
        vector_store=PathConfig(path=str(tmp_path)),
        chat_db=PathConfig(path=str(tmp_path / "c.db")),
        kb_db=PathConfig(path=str(tmp_path / "kb.db")),
        llm=LLMConfig(
            default_provider="ollama",
            default_model="phi3",
            ollama=OllamaConfig(host="http://x", tool_capable_models=[]),
        ),
        retrieval=RetrievalConfig(top_k=1, memory_turns=2),
    )

    class FakeSecrets:
        anthropic_api_key = None
        google_api_key = None

    app.state.secrets = FakeSecrets()
    app.include_router(auth_routes.router)
    app.include_router(conv_routes.router)
    app.include_router(chat_routes.router)
    c = TestClient(app)
    c.post("/api/auth/login", json={"password": "p"})
    yield c
    await db.close()


async def test_chat_creates_conversation_and_streams(client: TestClient):
    with client.stream(
        "POST",
        "/api/chat",
        json={"message": "hello", "provider": "ollama", "model": "phi3"},
    ) as resp:
        assert resp.status_code == 200
        body = b"".join(resp.iter_bytes())
    text = body.decode("utf-8")
    assert "event: conversation" in text
    assert "event: text-delta" in text
    assert "event: done" in text


async def test_chat_requires_auth(tmp_path: Path):
    app = FastAPI()
    app.state.auth = AuthService(hash_password("p"), "x" * 32)
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/x.db")
    await db.init_schema()
    app.state.chat_db = db
    app.include_router(chat_routes.router)
    c = TestClient(app)
    r = c.post("/api/chat", json={"message": "hi", "provider": "ollama", "model": "m"})
    assert r.status_code == 401
    await db.close()
