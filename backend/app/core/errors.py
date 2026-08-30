from fastapi import HTTPException, status

PROBLEM_JSON_MEDIA_TYPE = "application/problem+json"


class DailyNewsError(Exception):
    """Base exception for expected application errors."""


def problem_http_exception(status_code: int, title: str) -> HTTPException:
    """Create an HTTP error response using the API's problem media type."""
    return HTTPException(
        status_code=status_code,
        detail={"title": title, "status": status_code},
        headers={"content-type": PROBLEM_JSON_MEDIA_TYPE},
    )


def unauthenticated_error() -> HTTPException:
    """Create the standard response for missing or invalid credentials."""
    return problem_http_exception(status.HTTP_401_UNAUTHORIZED, "Authentication required")


def forbidden_error() -> HTTPException:
    """Create the standard response for a valid but unauthorized identity."""
    return problem_http_exception(status.HTTP_403_FORBIDDEN, "Access forbidden")
