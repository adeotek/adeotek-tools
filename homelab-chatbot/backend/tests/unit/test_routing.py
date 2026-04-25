from unittest.mock import MagicMock

from app.config import (
    AppConfig,
    EmbeddingsConfig,
    LLMConfig,
    OllamaConfig,
    PathConfig,
    RetrievalConfig,
    SyncConfig,
)
from app.llm.routing import get_available_tools, supports_tools


def _make_config(tool_capable_models: list[str]) -> AppConfig:
    return AppConfig(
        sync=SyncConfig(interval_seconds=180, state_file="/tmp/s.json"),
        repos=[],
        embeddings=EmbeddingsConfig(model="BAAI/bge-small-en-v1.5", cache_dir="/tmp"),
        vector_store=PathConfig(path="/tmp/lance"),
        chat_db=PathConfig(path="/tmp/chat.db"),
        kb_db=PathConfig(path="/tmp/kb.db"),
        llm=LLMConfig(
            default_provider="anthropic",
            default_model="claude-sonnet-4-6",
            ollama=OllamaConfig(
                host="http://localhost:11434",
                tool_capable_models=tool_capable_models,
            ),
        ),
        retrieval=RetrievalConfig(top_k=5, memory_turns=10),
    )


def test_anthropic_always_supports_tools():
    cfg = _make_config([])
    assert supports_tools("anthropic", "claude-haiku-4-5-20251001", cfg) is True


def test_google_always_supports_tools():
    cfg = _make_config([])
    assert supports_tools("google", "gemini-2.0-flash", cfg) is True


def test_ollama_whitelisted_model_supports_tools():
    cfg = _make_config(["llama3.1", "qwen2.5"])
    assert supports_tools("ollama", "llama3.1", cfg) is True


def test_ollama_whitelisted_model_with_tag_supports_tools():
    cfg = _make_config(["llama3.1", "qwen2.5"])
    assert supports_tools("ollama", "llama3.1:8b", cfg) is True


def test_ollama_non_whitelisted_model_no_tools():
    cfg = _make_config(["llama3.1"])
    assert supports_tools("ollama", "mistral", cfg) is False


def test_unknown_provider_no_tools():
    cfg = _make_config([])
    assert supports_tools("openai", "gpt-4o", cfg) is False


def test_get_available_tools_with_capable_provider():
    cfg = _make_config([])
    vec = MagicMock(name="vector_tool")
    sql = MagicMock(name="sql_tool")
    tools = get_available_tools("anthropic", "claude-sonnet-4-6", cfg, vec, sql)
    assert tools == [vec, sql]


def test_get_available_tools_without_capability():
    cfg = _make_config([])  # empty whitelist
    vec = MagicMock(name="vector_tool")
    sql = MagicMock(name="sql_tool")
    tools = get_available_tools("ollama", "mistral", cfg, vec, sql)
    assert tools == [vec]
