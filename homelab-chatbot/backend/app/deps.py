"""FastAPI dependencies shared across routes."""

from typing import cast

from fastapi import Cookie, HTTPException, Request, status

from app.auth import AuthService

SESSION_COOKIE_NAME = "hlcb_session"


def get_auth_service(request: Request) -> AuthService:
    return cast(AuthService, request.app.state.auth)


def require_session(
    request: Request,
    hlcb_session: str | None = Cookie(default=None),
) -> None:
    auth = get_auth_service(request)
    if not hlcb_session or not auth.verify_session_token(hlcb_session):
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED, detail="authentication required"
        )
