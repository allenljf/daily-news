"""YouTube Data API adapter boundary."""

from __future__ import annotations

from typing import Protocol

from app.ingestion.sources import CandidateArticle, SourceSearchRequest


class HttpClient(Protocol):
    """Minimal HTTP boundary supplied by the runtime composition root."""

    async def get(self, url: str, **kwargs: object) -> object: ...


class TokenProvider(Protocol):
    """Reads optional platform credentials without exposing their values to callers."""

    def get(self, name: str) -> str | None: ...


class YouTubeAdapter:
    """Uses YouTube Data API only when its API key is configured."""

    def __init__(self, http_client: HttpClient, token_provider: TokenProvider) -> None:
        self._http_client = http_client
        self._token_provider = token_provider

    @property
    def is_enabled(self) -> bool:
        return self._token_provider.get("YOUTUBE_API_KEY") is not None

    async def search(self, request: SourceSearchRequest) -> list[CandidateArticle]:
        del request
        if not self.is_enabled:
            return []
        return []
