import pytest
from pydantic import ValidationError

from app.secrets import Secrets


def test_required_secrets_present(monkeypatch):
    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", "$2b$12$abc")
    monkeypatch.setenv("HLCB_SESSION_SECRET", "s" * 32)
    s = Secrets()
    assert s.auth_password_hash == "$2b$12$abc"
    assert s.session_secret == "s" * 32


def test_missing_required_secret_raises(monkeypatch):
    monkeypatch.delenv("HLCB_AUTH_PASSWORD_HASH", raising=False)
    monkeypatch.delenv("HLCB_SESSION_SECRET", raising=False)
    with pytest.raises(ValidationError):
        Secrets()


def test_optional_api_keys_default_none(monkeypatch):
    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", "x")
    monkeypatch.setenv("HLCB_SESSION_SECRET", "s" * 32)
    monkeypatch.delenv("HLCB_ANTHROPIC_API_KEY", raising=False)
    s = Secrets()
    assert s.anthropic_api_key is None


def test_github_token_lookup(monkeypatch):
    monkeypatch.setenv("HLCB_AUTH_PASSWORD_HASH", "x")
    monkeypatch.setenv("HLCB_SESSION_SECRET", "s" * 32)
    monkeypatch.setenv("HLCB_GIT_TOKEN_DOCS", "ghp_abc")
    s = Secrets()
    assert s.github_token("HLCB_GIT_TOKEN_DOCS") == "ghp_abc"
    assert s.github_token("HLCB_GIT_TOKEN_MISSING") is None
