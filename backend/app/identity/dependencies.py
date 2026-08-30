"""FastAPI dependencies for authenticated API routes."""

import secrets
from typing import Annotated

from fastapi import Depends
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer

from app.core.config import Settings, get_settings
from app.core.errors import forbidden_error, unauthenticated_error
from app.identity.firebase import (
    InvalidFirebaseTokenError,
    TokenVerifier,
    VerifiedIdentity,
    get_firebase_token_verifier,
)

bearer_scheme = HTTPBearer(auto_error=False)


async def require_allowed_identity(
    credentials: Annotated[
        HTTPAuthorizationCredentials | None,
        Depends(bearer_scheme),
    ],
    verifier: Annotated[TokenVerifier, Depends(get_firebase_token_verifier)],
    settings: Annotated[Settings, Depends(get_settings)],
) -> VerifiedIdentity:
    """Require a valid Firebase token for the configured single-user email."""
    if credentials is None or credentials.scheme.lower() != "bearer":
        raise unauthenticated_error()

    try:
        identity = verifier.verify(credentials.credentials)
    except InvalidFirebaseTokenError:
        raise unauthenticated_error() from None

    if settings.allowed_user_email is None or not secrets.compare_digest(
        identity.email,
        settings.allowed_user_email,
    ):
        raise forbidden_error()

    return identity
