"""Password hashing and signed session cookies."""

import bcrypt
from itsdangerous import BadSignature, SignatureExpired, TimestampSigner

SESSION_TOKEN_PAYLOAD = "authenticated"
DEFAULT_SESSION_MAX_AGE = 60 * 60 * 24 * 7  # 7 days


def hash_password(plain: str) -> str:
    """Hash a plaintext password with bcrypt and return the encoded hash."""
    return bcrypt.hashpw(plain.encode("utf-8"), bcrypt.gensalt()).decode("utf-8")


def verify_password(plain: str, password_hash: str) -> bool:
    """Check a plaintext password against a bcrypt hash."""
    try:
        return bcrypt.checkpw(plain.encode("utf-8"), password_hash.encode("utf-8"))
    except ValueError:
        return False


class AuthService:
    """Issues and validates signed session tokens."""

    def __init__(self, password_hash: str, session_secret: str) -> None:
        self._password_hash = password_hash
        self._signer = TimestampSigner(session_secret)

    def check_password(self, plain: str) -> bool:
        return verify_password(plain, self._password_hash)

    def issue_session_token(self) -> str:
        return self._signer.sign(SESSION_TOKEN_PAYLOAD).decode("utf-8")

    def verify_session_token(
        self, token: str, max_age_seconds: int = DEFAULT_SESSION_MAX_AGE
    ) -> bool:
        try:
            payload = self._signer.unsign(token, max_age=max_age_seconds).decode("utf-8")
            return payload == SESSION_TOKEN_PAYLOAD
        except (BadSignature, SignatureExpired):
            return False
