import pytest

from app.storage.chat_db import (
    ChatDB,
)


@pytest.fixture
async def db(tmp_path):
    db = ChatDB(f"sqlite+aiosqlite:///{tmp_path}/test.db")
    await db.init_schema()
    yield db
    await db.close()


async def test_create_conversation(db: ChatDB):
    conv = await db.create_conversation(
        title="my first chat", provider="anthropic", model="claude-sonnet-4-6"
    )
    assert conv.id
    assert conv.title == "my first chat"
    assert conv.provider == "anthropic"


async def test_append_and_list_messages(db: ChatDB):
    conv = await db.create_conversation(title="t", provider="ollama", model="llama3.1")
    await db.append_message(conv.id, role="user", content="hello")
    await db.append_message(conv.id, role="assistant", content="hi there")
    msgs = await db.list_messages(conv.id)
    assert [m.role for m in msgs] == ["user", "assistant"]
    assert msgs[0].content == "hello"


async def test_list_conversations_orders_by_updated_desc(db: ChatDB):
    c1 = await db.create_conversation(title="first", provider="anthropic", model="x")
    _c2 = await db.create_conversation(title="second", provider="anthropic", model="x")
    await db.append_message(c1.id, role="user", content="new msg")
    convs = await db.list_conversations()
    assert convs[0].id == c1.id  # c1 updated most recently


async def test_delete_conversation_cascades_messages(db: ChatDB):
    conv = await db.create_conversation(title="t", provider="anthropic", model="x")
    await db.append_message(conv.id, role="user", content="bye")
    await db.delete_conversation(conv.id)
    msgs = await db.list_messages(conv.id)
    assert msgs == []


async def test_partial_flag_defaults_false(db: ChatDB):
    conv = await db.create_conversation(title="t", provider="anthropic", model="x")
    msg = await db.append_message(conv.id, role="assistant", content="x")
    assert msg.partial is False
