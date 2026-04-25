from llama_index.core.llms import LLM


def build_llm(
    *,
    provider: str,
    model: str,
    api_key: str | None = None,
    ollama_host: str | None = None,
) -> LLM:
    match provider:
        case "anthropic":
            from llama_index.llms.anthropic import Anthropic
            if not api_key:
                raise ValueError("api_key required for anthropic provider")
            return Anthropic(model=model, api_key=api_key)
        case "google":
            from llama_index.llms.google_genai import GoogleGenAI
            if not api_key:
                raise ValueError("api_key required for google provider")
            return GoogleGenAI(model=model, api_key=api_key)
        case "ollama":
            from llama_index.llms.ollama import Ollama
            return Ollama(
                model=model,
                base_url=ollama_host or "http://localhost:11434",
                request_timeout=120.0,
            )
        case _:
            raise ValueError(f"unknown provider: {provider!r}")
