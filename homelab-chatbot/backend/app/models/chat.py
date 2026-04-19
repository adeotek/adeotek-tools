"""Pydantic request/response schemas for the chat API."""

from typing import Literal

from pydantic import BaseModel


class ChatRequest(BaseModel):
    conv_id: str | None = None
    message: str
    provider: str
    model: str


class ConversationOut(BaseModel):
    id: str
    title: str
    provider: str
    model: str
    created_at: str
    updated_at: str


class MessageOut(BaseModel):
    id: int
    role: Literal["user", "assistant", "tool"]
    content: str
    tool_name: str | None = None
    tool_calls: str | None = None
    partial: bool = False
    created_at: str
