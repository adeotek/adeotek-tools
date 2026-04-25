from __future__ import annotations

import json
import urllib.request

ANTHROPIC_MODELS: list[str] = [
    "claude-opus-4-7",
    "claude-sonnet-4-6",
    "claude-haiku-4-5-20251001",
]

GOOGLE_MODELS: list[str] = [
    "gemini-2.5-pro",
    "gemini-2.0-flash",
    "gemini-2.0-flash-lite",
]


def list_models(provider: str, ollama_host: str = "http://localhost:11434") -> list[str]:
    """Return available models for the given provider.

    For ollama, fetches live from <ollama_host>/api/tags.
    Raises ValueError for unknown providers.
    """
    match provider:
        case "anthropic":
            return list(ANTHROPIC_MODELS)
        case "google":
            return list(GOOGLE_MODELS)
        case "ollama":
            return _fetch_ollama_models(ollama_host)
        case _:
            raise ValueError(f"unknown provider: {provider!r}")


def _fetch_ollama_models(host: str) -> list[str]:
    url = f"{host.rstrip('/')}/api/tags"
    try:
        with urllib.request.urlopen(url, timeout=5) as resp:
            data = json.loads(resp.read())
        return [m["name"] for m in data.get("models", []) if "name" in m]
    except Exception:
        return []
