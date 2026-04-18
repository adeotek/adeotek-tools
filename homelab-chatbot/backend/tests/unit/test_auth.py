import time

import pytest

from app.auth import (
    AuthService,
    hash_password,
    verify_password,
)


def test_hash_and_verify_password():
    h = hash_password("correct horse battery staple")
    assert verify_password("correct horse battery staple", h) is True
    assert verify_password("wrong", h) is False


@pytest.fixture
def auth():
    pwd_hash = hash_password("secret")
    return AuthService(password_hash=pwd_hash, session_secret="x" * 32)


def test_session_roundtrip(auth: AuthService):
    token = auth.issue_session_token()
    assert auth.verify_session_token(token) is True


def test_tampered_session_rejected(auth: AuthService):
    token = auth.issue_session_token()
    tampered = token[:-1] + ("A" if token[-1] != "A" else "B")
    assert auth.verify_session_token(tampered) is False


def test_expired_session_rejected(auth: AuthService):
    token = auth.issue_session_token()
    time.sleep(2.1)
    assert auth.verify_session_token(token, max_age_seconds=1) is False


def test_check_password_success(auth: AuthService):
    assert auth.check_password("secret") is True
    assert auth.check_password("wrong") is False
