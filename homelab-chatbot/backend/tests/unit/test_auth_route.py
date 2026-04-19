import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.auth import AuthService, hash_password
from app.routes.auth import router


@pytest.fixture
def client():
    app = FastAPI()
    app.state.auth = AuthService(
        password_hash=hash_password("letmein"), session_secret="x" * 32
    )
    app.include_router(router)
    return TestClient(app)


def test_login_with_correct_password_sets_cookie(client: TestClient):
    r = client.post("/api/auth/login", json={"password": "letmein"})
    assert r.status_code == 200
    assert "hlcb_session" in r.cookies


def test_login_with_wrong_password_rejected(client: TestClient):
    r = client.post("/api/auth/login", json={"password": "nope"})
    assert r.status_code == 401


def test_logout_clears_cookie(client: TestClient):
    client.post("/api/auth/login", json={"password": "letmein"})
    r = client.post("/api/auth/logout")
    assert r.status_code == 200
