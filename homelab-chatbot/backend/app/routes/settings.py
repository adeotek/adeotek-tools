"""Settings endpoint exposing available providers and models."""

from fastapi import APIRouter, Depends, Request
from pydantic import BaseModel

from app.config import AppConfig
from app.deps import require_session
from app.llm.registry import list_models
from app.llm.routing import supports_tools

router = APIRouter(
    prefix="/api/settings", tags=["settings"], dependencies=[Depends(require_session)]
)


class ProviderInfo(BaseModel):
    id: str
    available: bool
    models: list[str]
    tool_capable: list[str] = []


class SettingsOut(BaseModel):
    default_provider: str
    default_model: str
    providers: list[ProviderInfo]


@router.get("")
async def get_settings(request: Request) -> SettingsOut:
    cfg: AppConfig = request.app.state.config
    secrets_available: dict[str, bool] = request.app.state.secrets_available

    anthropic_models = list_models("anthropic")
    google_models = list_models("google")
    ollama_models = list_models("ollama", ollama_host=cfg.llm.ollama.host)

    providers = [
        ProviderInfo(
            id="anthropic",
            available=secrets_available.get("anthropic", False),
            models=anthropic_models,
            tool_capable=anthropic_models,
        ),
        ProviderInfo(
            id="google",
            available=secrets_available.get("google", False),
            models=google_models,
            tool_capable=google_models,
        ),
        ProviderInfo(
            id="ollama",
            available=len(ollama_models) > 0,
            models=ollama_models,
            tool_capable=[
                m for m in ollama_models if supports_tools("ollama", m, cfg)
            ],
        ),
    ]

    return SettingsOut(
        default_provider=cfg.llm.default_provider,
        default_model=cfg.llm.default_model,
        providers=providers,
    )
