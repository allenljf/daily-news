"""Grounded public-web Gemini source adapter with no SDK or secret dependency."""

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Protocol
from urllib.parse import urlsplit

from app.ingestion.sources import CandidateArticle, SourceSearchRequest, limit_candidates


@dataclass(frozen=True)
class GeminiSearchRequest:
    """SDK-neutral Gemini request that enables current grounding tools."""

    prompt: str
    google_search: bool
    url_context_urls: tuple[str, ...]


@dataclass(frozen=True)
class GeminiGroundedCandidate:
    """A structured candidate and its grounding citation returned by a Gemini client."""

    title: str
    url: str | None
    citation_url: str | None
    source_name: str
    summary: str | None = None
    published_at: datetime | None = None


class GeminiClient(Protocol):
    """The narrow boundary implemented by the selected Gemini SDK adapter."""

    async def search(self, request: GeminiSearchRequest) -> list[GeminiGroundedCandidate]: ...


class GeminiWebAdapter:
    """Turns Google-Search-grounded Gemini results into safe public Article candidates."""

    def __init__(self, client: GeminiClient) -> None:
        self._client = client

    async def search(self, request: SourceSearchRequest) -> list[CandidateArticle]:
        response = await self._client.search(self._gemini_request(request))
        return limit_candidates(
            [
                CandidateArticle(
                    title=candidate.title,
                    canonical_url=candidate.url,
                    citation_url=candidate.citation_url,
                    source_name=candidate.source_name,
                    summary=candidate.summary,
                    published_at=candidate.published_at,
                )
                for candidate in response
                if candidate.url
                and candidate.citation_url
                and _is_public_http_url(candidate.url)
                and _is_public_http_url(candidate.citation_url)
            ]
        )

    @staticmethod
    def _gemini_request(request: SourceSearchRequest) -> GeminiSearchRequest:
        url_context_urls = (
            (request.website_input,) if _is_public_http_url(request.website_input) else ()
        )
        prompt = "\n".join(
            (
                "Find recent, publicly verifiable news candidates.",
                f"Category: {request.category_name}",
                f"Keywords: {request.search_keywords or request.category_name}",
                f"Source Setting: {request.source_label} ({request.website_input or 'open web'})",
                f"Special requirements: {request.special_requirements or 'none'}",
                "Return a public article URL and a public grounding citation for every candidate.",
            )
        )
        return GeminiSearchRequest(
            prompt=prompt,
            google_search=True,
            url_context_urls=url_context_urls,
        )


def _is_public_http_url(value: str) -> bool:
    parsed = urlsplit(value)
    return parsed.scheme in {"http", "https"} and bool(parsed.hostname)
