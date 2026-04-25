"""Environment-variable-backed secrets."""

import os

from pydantic import Field
from pydantic_settings import BaseSettings, SettingsConfigDict


class Secrets(BaseSettings):
    """Secrets loaded from environment variables."""

    model_config = SettingsConfigDict(case_sensitive=True, extra="ignore")

    auth_password_hash: str = Field(alias="HLCB_AUTH_PASSWORD_HASH")
    session_secret: str = Field(alias="HLCB_SESSION_SECRET", min_length=16)

    anthropic_api_key: str | None = Field(default=None, alias="HLCB_ANTHROPIC_API_KEY")
    google_api_key: str | None = Field(default=None, alias="HLCB_GOOGLE_API_KEY")

    def github_token(self, env_var: str) -> str | None:
        """Look up a GitHub token by the env-var name referenced in config.yaml."""
        return os.environ.get(env_var)
