from collections.abc import Callable

from fastapi import Depends, FastAPI
from fastapi.testclient import TestClient

from app.core.config import Settings, get_settings
from app.identity.dependencies import require_allowed_identity
from app.identity.firebase import (
    InvalidFirebaseTokenError,
    VerifiedIdentity,
    get_firebase_token_verifier,
)


class FakeTokenVerifier:
    def __init__(self, verify: Callable[[str], VerifiedIdentity]) -> None:
        self._verify = verify

    def verify(self, token: str) -> VerifiedIdentity:
        return self._verify(token)


def create_protected_client(
    verifier: FakeTokenVerifier,
    allowed_email: str = "allowed@example.com",
) -> TestClient:
    app = FastAPI()

    @app.get("/protected")
    async def protected(
        identity: VerifiedIdentity = Depends(require_allowed_identity),
    ) -> dict[str, str]:
        return {"uid": identity.uid}

    app.dependency_overrides[get_firebase_token_verifier] = lambda: verifier
    app.dependency_overrides[get_settings] = lambda: Settings(
        allowed_user_email=allowed_email,
    )
    return TestClient(app)


def test_missing_bearer_token_returns_401() -> None:
    client = create_protected_client(
        FakeTokenVerifier(lambda token: VerifiedIdentity(uid="user", email="allowed@example.com")),
    )

    response = client.get("/protected")

    assert response.status_code == 401
    assert response.headers["content-type"] == "application/problem+json"


def test_invalid_firebase_token_returns_401() -> None:
    def reject_token(token: str) -> VerifiedIdentity:
        raise InvalidFirebaseTokenError

    client = create_protected_client(FakeTokenVerifier(reject_token))

    response = client.get("/protected", headers={"Authorization": "Bearer invalid-token"})

    assert response.status_code == 401
    assert response.headers["content-type"] == "application/problem+json"


def test_valid_token_with_non_allowlisted_email_returns_403() -> None:
    client = create_protected_client(
        FakeTokenVerifier(lambda token: VerifiedIdentity(uid="user", email="other@example.com")),
    )

    response = client.get("/protected", headers={"Authorization": "Bearer valid-token"})

    assert response.status_code == 403
    assert response.headers["content-type"] == "application/problem+json"
