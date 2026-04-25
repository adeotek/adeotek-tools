"""Application configuration loaded from YAML + environment variables."""

from pathlib import Path
from typing import Literal

import yaml
from pydantic import BaseModel, Field, field_validator

Provider = Literal["anthropic", "google", "ollama"]


class SyncConfig(BaseModel):
    interval_seconds: int = Field(ge=30, le=3600)
    state_file: str
    clone_root: str = "/data/repos"


class RepoConfig(BaseModel):
    name: str
    url: str
    branch: str = "main"
    token_env: str
    include_globs: list[str] = Field(default_factory=lambda: ["**/*.md"])


class EmbeddingsConfig(BaseModel):
    model: str
    cache_dir: str


class PathConfig(BaseModel):
    path: str


class OllamaConfig(BaseModel):
    host: str
    tool_capable_models: list[str] = Field(default_factory=list)


class LLMConfig(BaseModel):
    default_provider: Provider
    default_model: str
    ollama: OllamaConfig


class RetrievalConfig(BaseModel):
    top_k: int = Field(ge=1, le=50, default=5)
    memory_turns: int = Field(ge=0, le=100, default=10)


class AppConfig(BaseModel):
    sync: SyncConfig
    repos: list[RepoConfig]
    embeddings: EmbeddingsConfig
    vector_store: PathConfig
    chat_db: PathConfig
    kb_db: PathConfig
    llm: LLMConfig
    retrieval: RetrievalConfig

    @field_validator("repos")
    @classmethod
    def unique_repo_names(cls, v: list[RepoConfig]) -> list[RepoConfig]:
        names = [r.name for r in v]
        if len(names) != len(set(names)):
            raise ValueError("repo names must be unique")
        return v


def load_config(path: Path) -> AppConfig:
    """Load configuration from a YAML file and validate it."""
    if not path.exists():
        raise FileNotFoundError(f"config file not found: {path}")
    with path.open() as f:
        raw = yaml.safe_load(f)
    return AppConfig.model_validate(raw)
