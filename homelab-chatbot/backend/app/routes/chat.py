"""Streaming chat endpoint."""

import json
from typing import AsyncIterator

from fastapi import APIRouter, Depends, HTTPException, Request, status
from fastapi.responses import StreamingResponse

from app.config import AppConfig
from app.deps import require_session
from app.llm.agent import run_agent
from app.llm.provider import build_llm
from app.llm.routing import supports_tools
from app.llm.sse import AgentEvent, format_sse
from app.models.chat import ChatRequest
from app.retrieval.sql_tool import SQLTool
from app.retrieval.vector_tool import VectorSearchTool
from app.storage.chat_db import ChatDB

router = APIRouter(prefix="/api", tags=["chat"], dependencies=[Depends(require_session)])


async def _conv_history(db: ChatDB, conv_id: str, turns: int) -> list[tuple[str, str]]:
    msgs = await db.list_messages(conv_id)
    history = [(m.role, m.content) for m in msgs if m.role in ("user", "assistant")]
    return history[-turns * 2 :]


@router.post("/chat")
async def chat(body: ChatRequest, request: Request) -> StreamingResponse:
    cfg: AppConfig = request.app.state.config
    db: ChatDB = request.app.state.chat_db
    vector_tool: VectorSearchTool = request.app.state.vector_tool
    sql_tool: SQLTool = request.app.state.sql_tool
    secrets = request.app.state.secrets

    if body.conv_id:
        conv = await db.get_conversation(body.conv_id)
        if not conv:
            raise HTTPException(status_code=status.HTTP_404_NOT_FOUND)
    else:
        title = body.message[:60] or "New chat"
        conv = await db.create_conversation(
            title=title, provider=body.provider, model=body.model
        )

    api_key = None
    if body.provider == "anthropic":
        api_key = secrets.anthropic_api_key
    elif body.provider == "google":
        api_key = secrets.google_api_key

    llm = build_llm(
        provider=body.provider,
        model=body.model,
        api_key=api_key,
        ollama_host=cfg.llm.ollama.host,
    )

    tools_enabled = supports_tools(body.provider, body.model, cfg)
    tools = (
        [vector_tool.as_llama_tool(), sql_tool.as_llama_tool(llm)] if tools_enabled else []
    )

    history = await _conv_history(db, conv.id, cfg.retrieval.memory_turns)
    await db.append_message(conv.id, role="user", content=body.message)

    async def _event_stream() -> AsyncIterator[bytes]:
        buffer: list[str] = []
        tool_events: list[dict] = []
        yield format_sse(AgentEvent("conversation", {"id": conv.id}))
        try:
            async for event in run_agent(
                llm=llm,
                tools=tools,
                history=history,
                user_message=body.message,
                tools_enabled=tools_enabled,
                vector_tool=vector_tool if tools_enabled else None,
            ):
                if event.kind == "text-delta":
                    buffer.append(event.data.get("text", ""))
                elif event.kind in ("tool-call", "tool-result"):
                    tool_events.append({"kind": event.kind, **event.data})
                yield format_sse(event)
        finally:
            full = "".join(buffer)
            await db.append_message(
                conv.id,
                role="assistant",
                content=full,
                tool_calls=json.dumps(tool_events) if tool_events else None,
                partial=False,
            )

    return StreamingResponse(_event_stream(), media_type="text/event-stream")
