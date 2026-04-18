"""Async SQLAlchemy models and CRUD for conversations and messages."""

import uuid
from datetime import datetime, timezone

from sqlalchemy import ForeignKey, Index, String, delete, select
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column, relationship


def _utcnow() -> datetime:
    return datetime.now(timezone.utc)


class Base(DeclarativeBase):
    pass


class Conversation(Base):
    __tablename__ = "conversations"

    id: Mapped[str] = mapped_column(String, primary_key=True)
    title: Mapped[str] = mapped_column(String, nullable=False)
    provider: Mapped[str] = mapped_column(String, nullable=False)
    model: Mapped[str] = mapped_column(String, nullable=False)
    created_at: Mapped[datetime] = mapped_column(default=_utcnow)
    updated_at: Mapped[datetime] = mapped_column(default=_utcnow)

    messages: Mapped[list["Message"]] = relationship(
        back_populates="conversation",
        cascade="all, delete-orphan",
        order_by="Message.created_at",
    )


class Message(Base):
    __tablename__ = "messages"

    id: Mapped[int] = mapped_column(primary_key=True, autoincrement=True)
    conv_id: Mapped[str] = mapped_column(
        ForeignKey("conversations.id", ondelete="CASCADE"), nullable=False
    )
    role: Mapped[str] = mapped_column(String, nullable=False)
    content: Mapped[str] = mapped_column(String, nullable=False)
    tool_calls: Mapped[str | None] = mapped_column(String, nullable=True)
    tool_name: Mapped[str | None] = mapped_column(String, nullable=True)
    partial: Mapped[bool] = mapped_column(default=False, nullable=False)
    created_at: Mapped[datetime] = mapped_column(default=_utcnow)

    conversation: Mapped[Conversation] = relationship(back_populates="messages")

    __table_args__ = (Index("ix_messages_conv_created", "conv_id", "created_at"),)


class ChatDB:
    """Async wrapper around chat.db for conversations and messages."""

    def __init__(self, url: str) -> None:
        self._engine = create_async_engine(url, echo=False, future=True)
        if url.startswith("sqlite"):
            from sqlalchemy import event as _event

            @_event.listens_for(self._engine.sync_engine, "connect")
            def _set_sqlite_pragma(dbapi_conn, _connection_record):  # type: ignore[no-untyped-def]
                cursor = dbapi_conn.cursor()
                cursor.execute("PRAGMA foreign_keys=ON")
                cursor.close()

        self._session_factory = async_sessionmaker(self._engine, expire_on_commit=False)

    async def init_schema(self) -> None:
        async with self._engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)

    async def close(self) -> None:
        await self._engine.dispose()

    def session(self) -> AsyncSession:
        return self._session_factory()

    async def create_conversation(
        self, title: str, provider: str, model: str
    ) -> Conversation:
        conv = Conversation(
            id=str(uuid.uuid4()), title=title, provider=provider, model=model
        )
        async with self.session() as s:
            s.add(conv)
            await s.commit()
            await s.refresh(conv)
        return conv

    async def append_message(
        self,
        conv_id: str,
        role: str,
        content: str,
        tool_calls: str | None = None,
        tool_name: str | None = None,
        partial: bool = False,
    ) -> Message:
        async with self.session() as s:
            msg = Message(
                conv_id=conv_id,
                role=role,
                content=content,
                tool_calls=tool_calls,
                tool_name=tool_name,
                partial=partial,
            )
            s.add(msg)
            conv = await s.get(Conversation, conv_id)
            if conv:
                conv.updated_at = _utcnow()
            await s.commit()
            await s.refresh(msg)
        return msg

    async def list_messages(self, conv_id: str) -> list[Message]:
        async with self.session() as s:
            result = await s.execute(
                select(Message).where(Message.conv_id == conv_id).order_by(Message.created_at, Message.id)
            )
            return list(result.scalars().all())

    async def list_conversations(self) -> list[Conversation]:
        async with self.session() as s:
            result = await s.execute(
                select(Conversation).order_by(Conversation.updated_at.desc())
            )
            return list(result.scalars().all())

    async def get_conversation(self, conv_id: str) -> Conversation | None:
        async with self.session() as s:
            return await s.get(Conversation, conv_id)

    async def delete_conversation(self, conv_id: str) -> None:
        async with self.session() as s:
            await s.execute(delete(Conversation).where(Conversation.id == conv_id))
            await s.commit()

    async def rename_conversation(self, conv_id: str, title: str) -> None:
        async with self.session() as s:
            conv = await s.get(Conversation, conv_id)
            if conv:
                conv.title = title
                conv.updated_at = _utcnow()
                await s.commit()
