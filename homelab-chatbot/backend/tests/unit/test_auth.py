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
    # Corrupt the payload (the part before the first dot). The HMAC is computed
    # over payload+timestamp, so any change here always invalidates the signature,
    # unlike tampering with the last base64 character of the signature which can
    # be a no-op due to unused padding bits in the final base64 group.
    first_dot = token.index(".")
    last_payload_char = token[first_dot - 1]
    replacement = "X" if last_payload_char != "X" else "Y"
    tampered = token[: first_dot - 1] + replacement + token[first_dot:]
    assert auth.verify_session_token(tampered) is False


def test_expired_session_rejected(auth: AuthService):
    token = auth.issue_session_token()
    time.sleep(2.1)
    assert auth.verify_session_token(token, max_age_seconds=1) is False


def test_check_password_success(auth: AuthService):
    assert auth.check_password("secret") is True
    assert auth.check_password("wrong") is False
