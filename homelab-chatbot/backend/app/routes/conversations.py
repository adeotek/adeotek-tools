"""Conversation CRUD endpoints."""

from fastapi import APIRouter, Depends, HTTPException, Request, status
from pydantic import BaseModel

from app.deps import require_session
from app.models.chat import ConversationOut, MessageOut
from app.storage.chat_db import ChatDB

router = APIRouter(
    prefix="/api/conv", tags=["conversations"], dependencies=[Depends(require_session)]
)


class CreateConversationRequest(BaseModel):
    title: str
    provider: str
    model: str


class RenameRequest(BaseModel):
    title: str


def _db(request: Request) -> ChatDB:
    return request.app.state.chat_db


def _to_out(c) -> ConversationOut:
    return ConversationOut(
        id=c.id,
        title=c.title,
        provider=c.provider,
        model=c.model,
        created_at=c.created_at.isoformat(),
        updated_at=c.updated_at.isoformat(),
    )


@router.get("")
async def list_conversations(request: Request) -> list[ConversationOut]:
    db = _db(request)
    convs = await db.list_conversations()
    return [_to_out(c) for c in convs]


@router.post("")
async def create_conversation(
    body: CreateConversationRequest, request: Request
) -> ConversationOut:
    db = _db(request)
    conv = await db.create_conversation(
        title=body.title, provider=body.provider, model=body.model
    )
    return _to_out(conv)


@router.get("/{conv_id}/messages")
async def list_messages(conv_id: str, request: Request) -> list[MessageOut]:
    db = _db(request)
    conv = await db.get_conversation(conv_id)
    if not conv:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND)
    msgs = await db.list_messages(conv_id)
    return [
        MessageOut(
            id=m.id,
            role=m.role,
            content=m.content,
            tool_name=m.tool_name,
            tool_calls=m.tool_calls,
            partial=m.partial,
            created_at=m.created_at.isoformat(),
        )
        for m in msgs
    ]


@router.patch("/{conv_id}")
async def rename(conv_id: str, body: RenameRequest, request: Request) -> dict[str, bool]:
    db = _db(request)
    conv = await db.get_conversation(conv_id)
    if not conv:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND)
    await db.rename_conversation(conv_id, body.title)
    return {"ok": True}


@router.delete("/{conv_id}")
async def delete(conv_id: str, request: Request) -> dict[str, bool]:
    db = _db(request)
    await db.delete_conversation(conv_id)
    return {"ok": True}
