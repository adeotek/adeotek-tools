import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.auth import AuthService, hash_password
from app.routes import auth as auth_routes
from app.routes import conversations as conv_routes
from app.storage.chat_db import ChatDB


@pytest.fixture
async def app_client(tmp_path):
    app = FastAPI()
    app.state.auth = AuthService(
        password_hash=hash_password("p"), session_secret="x" * 32
    )
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/chat.db")
    await db.init_schema()
    app.state.chat_db = db
    app.include_router(auth_routes.router)
    app.include_router(conv_routes.router)
    client = TestClient(app)
    client.post("/api/auth/login", json={"password": "p"})
    yield client
    await db.close()


async def test_create_then_list_conversation(app_client: TestClient):
    r = app_client.post(
        "/api/conv",
        json={"title": "chat-a", "provider": "anthropic", "model": "claude-sonnet-4-6"},
    )
    assert r.status_code == 200
    conv_id = r.json()["id"]

    r = app_client.get("/api/conv")
    data = r.json()
    assert any(c["id"] == conv_id for c in data)


async def test_get_messages_returns_empty_on_new_conv(app_client: TestClient):
    r = app_client.post(
        "/api/conv", json={"title": "t", "provider": "ollama", "model": "llama3.1"}
    )
    conv_id = r.json()["id"]
    r = app_client.get(f"/api/conv/{conv_id}/messages")
    assert r.status_code == 200
    assert r.json() == []


async def test_delete_conversation(app_client: TestClient):
    r = app_client.post(
        "/api/conv", json={"title": "t", "provider": "ollama", "model": "m"}
    )
    conv_id = r.json()["id"]
    r = app_client.delete(f"/api/conv/{conv_id}")
    assert r.status_code == 200
    r = app_client.get(f"/api/conv/{conv_id}/messages")
    assert r.status_code == 404


async def test_rename_conversation(app_client: TestClient):
    r = app_client.post(
        "/api/conv", json={"title": "old", "provider": "ollama", "model": "m"}
    )
    conv_id = r.json()["id"]
    r = app_client.patch(f"/api/conv/{conv_id}", json={"title": "new"})
    assert r.status_code == 200
    r = app_client.get("/api/conv")
    assert any(c["title"] == "new" for c in r.json())


async def test_list_conversations_requires_auth(tmp_path):
    app = FastAPI()
    app.state.auth = AuthService(
        password_hash=hash_password("p"), session_secret="x" * 32
    )
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/c.db")
    await db.init_schema()
    app.state.chat_db = db
    app.include_router(conv_routes.router)
    client = TestClient(app)
    r = client.get("/api/conv")
    assert r.status_code == 401
    await db.close()
