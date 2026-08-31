"""GitHub REST API adapter boundary."""

from __future__ import annotations

from app.ingestion.sources import CandidateArticle, SourceSearchRequest
from app.ingestion.youtube_adapter import HttpClient, TokenProvider


class GitHubAdapter:
    """Uses public GitHub REST requests with an optional rate-limit token."""

    def __init__(self, http_client: HttpClient, token_provider: TokenProvider) -> None:
        self._http_client = http_client
        self._token_provider = token_provider

    async def search(self, request: SourceSearchRequest) -> list[CandidateArticle]:
        del request
        return []
