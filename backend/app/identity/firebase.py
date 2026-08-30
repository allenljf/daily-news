"""Firebase Admin token verification adapter."""

from dataclasses import dataclass
from functools import lru_cache
from typing import Protocol


@dataclass(frozen=True)
class VerifiedIdentity:
    """The identity claims needed by the single-user API."""

    uid: str
    email: str


class InvalidFirebaseTokenError(Exception):
    """Raised when Firebase cannot verify an ID token."""


class TokenVerifier(Protocol):
    """Verifies an ID token without exposing Firebase to route dependencies."""

    def verify(self, token: str) -> VerifiedIdentity: ...


class FirebaseAdminTokenVerifier:
    """Verify Firebase ID tokens using Application Default Credentials."""

    def verify(self, token: str) -> VerifiedIdentity:
        import firebase_admin
        from firebase_admin import auth

        try:
            firebase_admin.get_app()
        except ValueError:
            firebase_admin.initialize_app()

        try:
            claims = auth.verify_id_token(token)
        except Exception as error:
            raise InvalidFirebaseTokenError from error

        uid = claims.get("uid")
        email = claims.get("email")
        if not isinstance(uid, str) or not uid or not isinstance(email, str) or not email:
            raise InvalidFirebaseTokenError
        return VerifiedIdentity(uid=uid, email=email)


@lru_cache
def get_firebase_token_verifier() -> TokenVerifier:
    """Return the production Firebase Admin verifier."""
    return FirebaseAdminTokenVerifier()
