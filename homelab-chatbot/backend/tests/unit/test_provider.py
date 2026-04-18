from unittest.mock import MagicMock, patch
import pytest
from app.llm.provider import build_llm


def test_build_ollama():
    with patch("llama_index.llms.ollama.Ollama") as MockOllama:
        MockOllama.return_value = MagicMock()
        result = build_llm(provider="ollama", model="llama3.1")
        MockOllama.assert_called_once_with(
            model="llama3.1",
            base_url="http://localhost:11434",
            request_timeout=120.0,
        )


def test_build_ollama_custom_host():
    with patch("llama_index.llms.ollama.Ollama") as MockOllama:
        MockOllama.return_value = MagicMock()
        build_llm(provider="ollama", model="llama3.1", ollama_host="http://my-nas:11434")
        MockOllama.assert_called_once_with(
            model="llama3.1",
            base_url="http://my-nas:11434",
            request_timeout=120.0,
        )


def test_build_anthropic():
    with patch("llama_index.llms.anthropic.Anthropic") as MockAnthropic:
        MockAnthropic.return_value = MagicMock()
        build_llm(provider="anthropic", model="claude-haiku-4-5-20251001", api_key="sk-test")
        MockAnthropic.assert_called_once_with(model="claude-haiku-4-5-20251001", api_key="sk-test")


def test_build_google():
    with patch("llama_index.llms.google_genai.GoogleGenAI") as MockGemini:
        MockGemini.return_value = MagicMock()
        build_llm(provider="google", model="gemini-2.0-flash", api_key="gkey")
        MockGemini.assert_called_once_with(model="gemini-2.0-flash", api_key="gkey")


def test_unknown_provider_raises():
    with pytest.raises(ValueError, match="unknown provider"):
        build_llm(provider="openai", model="gpt-4o")


def test_anthropic_missing_key_raises():
    with pytest.raises(ValueError, match="api_key required"):
        build_llm(provider="anthropic", model="claude-haiku-4-5-20251001")


def test_google_missing_key_raises():
    with pytest.raises(ValueError, match="api_key required"):
        build_llm(provider="google", model="gemini-2.0-flash")
