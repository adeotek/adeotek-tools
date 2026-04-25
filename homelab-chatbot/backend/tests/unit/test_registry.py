import json
from unittest.mock import MagicMock, patch

import pytest

from app.llm.registry import ANTHROPIC_MODELS, GOOGLE_MODELS, list_models


def test_anthropic_returns_static_list():
    models = list_models("anthropic")
    assert models == ANTHROPIC_MODELS
    assert "claude-sonnet-4-6" in models


def test_google_returns_static_list():
    models = list_models("google")
    assert models == GOOGLE_MODELS
    assert "gemini-2.0-flash" in models


def test_ollama_fetches_from_api():
    fake_response = json.dumps({"models": [{"name": "llama3.1"}, {"name": "qwen2.5"}]}).encode()
    mock_resp = MagicMock()
    mock_resp.read.return_value = fake_response
    mock_resp.__enter__ = lambda s: s
    mock_resp.__exit__ = MagicMock(return_value=False)

    with patch("urllib.request.urlopen", return_value=mock_resp):
        models = list_models("ollama", ollama_host="http://localhost:11434")
    assert models == ["llama3.1", "qwen2.5"]


def test_ollama_unreachable_returns_empty():
    with patch("urllib.request.urlopen", side_effect=OSError("connection refused")):
        models = list_models("ollama")
    assert models == []


def test_unknown_provider_raises():
    with pytest.raises(ValueError, match="unknown provider"):
        list_models("openai")


def test_returns_copy_not_original():
    models = list_models("anthropic")
    models.append("fake")
    assert "fake" not in list_models("anthropic")
