"""Meta API adapter boundary, enabled only by an authorized platform token."""

from __future__ import annotations

from app.ingestion.sources import CandidateArticle, SourceSearchRequest
from app.ingestion.youtube_adapter import HttpClient, TokenProvider


class MetaAdapter:
    """Uses authorized Meta APIs and never falls back to general web search."""

    def __init__(self, http_client: HttpClient, token_provider: TokenProvider) -> None:
        self._http_client = http_client
        self._token_provider = token_provider

    @property
    def is_enabled(self) -> bool:
        return self._token_provider.get("META_ACCESS_TOKEN") is not None

    async def search(self, request: SourceSearchRequest) -> list[CandidateArticle]:
        del request
        if not self.is_enabled:
            return []
        return []
