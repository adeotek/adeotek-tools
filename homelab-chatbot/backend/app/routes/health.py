"""Health check endpoint (unauthenticated)."""

from fastapi import APIRouter, Request

router = APIRouter(prefix="/api", tags=["health"])


@router.get("/health")
async def health(request: Request) -> dict:
    state = getattr(request.app.state, "sync_state", {})
    return {"status": "ok", "sync": state}
