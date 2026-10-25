"""Health check endpoint (unauthenticated)."""

from typing import Any

from fastapi import APIRouter, Request

router = APIRouter(prefix="/api", tags=["health"])


@router.get("/health")
async def health(request: Request) -> dict[str, Any]:
    state = getattr(request.app.state, "sync_state", {})
    return {"status": "ok", "sync": state}
