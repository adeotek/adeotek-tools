from __future__ import annotations

from app.config import AppConfig


def supports_tools(provider: str, model: str, config: AppConfig) -> bool:
    """Return True if the provider+model combination supports native tool-calling."""
    match provider:
        case "anthropic" | "google":
            return True
        case "ollama":
            capable = config.llm.ollama.tool_capable_models
            return any(model.startswith(m) for m in capable)
        case _:
            return False


def get_available_tools(
    provider: str,
    model: str,
    config: AppConfig,
    vector_tool,
    sql_tool,
) -> list:
    """Return the list of LlamaIndex tools appropriate for this provider+model."""
    if supports_tools(provider, model, config):
        return [vector_tool, sql_tool]
    return [vector_tool]
