"""Contracts for the grounded public-web Gemini adapter."""

from __future__ import annotations

import asyncio

from app.ingestion.gemini_adapter import GeminiGroundedCandidate, GeminiWebAdapter
from app.ingestion.sources import SourceSearchRequest


class FakeGeminiClient:
    def __init__(self, candidates: list[GeminiGroundedCandidate]) -> None:
        self.candidates = candidates
        self.request = None

    async def search(self, request: object) -> list[GeminiGroundedCandidate]:
        self.request = request
        return self.candidates


def _request() -> SourceSearchRequest:
    return SourceSearchRequest(
        category_name="AI research",
        search_keywords="agents",
        source_label="OpenAI blog",
        website_input="https://openai.com/blog",
        special_requirements="prefer primary sources",
    )


def test_grounded_adapter_includes_category_source_and_requirements_in_prompt() -> None:
    """A missing search-context field would produce irrelevant public-web candidates."""
    client = FakeGeminiClient([])

    asyncio.run(GeminiWebAdapter(client).search(_request()))

    assert client.request is not None
    assert client.request.google_search is True
    assert client.request.url_context_urls == ("https://openai.com/blog",)
    for value in ("AI research", "agents", "OpenAI blog", "prefer primary sources"):
        assert value in client.request.prompt


def test_grounded_adapter_rejects_invalid_candidates_and_caps_output() -> None:
    """A missing public URL/citation or over-quota result must never reach persistence."""
    valid = [
        GeminiGroundedCandidate(
            title=f"Article {index}",
            url=f"https://example.com/{index}",
            citation_url=f"https://source.example/{index}",
            source_name="Example",
        )
        for index in range(11)
    ]
    client = FakeGeminiClient(
        [
            GeminiGroundedCandidate("No URL", None, "https://source.example/a", "Example"),
            GeminiGroundedCandidate("No citation", "https://example.com/a", None, "Example"),
            GeminiGroundedCandidate(
                "Private URL",
                "ftp://example.com/a",
                "https://source.example/a",
                "Example",
            ),
            *valid,
        ]
    )

    candidates = asyncio.run(GeminiWebAdapter(client).search(_request()))

    assert [candidate.title for candidate in candidates] == [
        f"Article {index}" for index in range(10)
    ]
    assert all(candidate.citation_url for candidate in candidates)
