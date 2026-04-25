"""End-to-end smoke tests: health, auth, ingestion, and retrieval wiring."""

from __future__ import annotations

import pytest
from httpx import AsyncClient

import app.routes.chat as chat_module


@pytest.mark.asyncio
async def test_health(client: AsyncClient) -> None:
    r = await client.get("/api/health")
    assert r.status_code == 200
    assert r.json()["status"] == "ok"


@pytest.mark.asyncio
async def test_login_sets_session_cookie(client: AsyncClient) -> None:
    r = await client.post("/api/auth/login", json={"password": "test"})
    assert r.status_code == 200
    assert "hlcb_session" in r.cookies


@pytest.mark.asyncio
async def test_login_rejects_wrong_password(client: AsyncClient) -> None:
    r = await client.post("/api/auth/login", json={"password": "nope"})
    assert r.status_code == 401


def test_ingestion_populates_vector_store(env) -> None:
    """Verify the env fixture populated the vector store during setup."""
    assert env.chunk_count >= 1, f"expected chunks in vector store, got {env.chunk_count}"


@pytest.mark.asyncio
async def test_chat_stream_emits_sse_events(
    env, client: AsyncClient, monkeypatch: pytest.MonkeyPatch
) -> None:
    """SSE framing and 'done' event with a stub LLM. Vector store pre-populated by env fixture."""

    class _FakeLLM:
        async def astream_chat(self, messages):  # noqa: ANN001
            async def _gen():
                yield type("Delta", (), {"delta": "VLAN 10 is management."})()

            return _gen()

    monkeypatch.setattr(chat_module, "build_llm", lambda **kw: _FakeLLM())
    # Force tools-disabled path (ReActAgent API differs across llama_index versions)
    monkeypatch.setattr(chat_module, "supports_tools", lambda *a, **kw: False)

    await client.post("/api/auth/login", json={"password": "test"})

    async with client.stream(
        "POST",
        "/api/chat",
        json={
            "provider": "anthropic",
            "model": "claude-sonnet-4-6",
            "message": "Which VLAN is management?",
        },
    ) as resp:
        assert resp.status_code == 200
        kinds: list[str] = []
        current_kind: str | None = None
        async for line in resp.aiter_lines():
            if line.startswith("event: "):
                current_kind = line[len("event: "):].strip()
            elif line.startswith("data: ") and current_kind:
                kinds.append(current_kind)
                current_kind = None

    assert "done" in kinds, f"expected a 'done' event, got {kinds}"
