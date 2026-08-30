"""Contracts for Article duplicate detection before persistence."""

from __future__ import annotations

import asyncio

from app.ingestion.dedupe import DedupeService, DuplicateReason
from app.ingestion.sources import MAX_CANDIDATES_PER_SOURCE, CandidateArticle, limit_candidates


class FakeFingerprintLookup:
    """Records URL/title checks while representing all Articles, including soft-deleted ones."""

    def __init__(
        self,
        *,
        matching_url_hash: str | None = None,
        matching_title_hash: str | None = None,
    ) -> None:
        self.matching_url_hash = matching_url_hash
        self.matching_title_hash = matching_title_hash
        self.checks: list[str] = []

    async def has_canonical_url_hash(self, fingerprint: str) -> bool:
        self.checks.append("url")
        return fingerprint == self.matching_url_hash

    async def has_normalized_title_hash(self, fingerprint: str) -> bool:
        self.checks.append("title")
        return fingerprint == self.matching_title_hash


def _candidate(
    *,
    title: str = "Shared title",
    url: str = "https://example.com/article",
) -> CandidateArticle:
    return CandidateArticle(title=title, canonical_url=url, source_name="Example")


def test_dedupe_checks_title_when_a_different_url_has_the_same_title() -> None:
    """A broken title fallback would let a duplicate Article be persisted."""
    candidate = _candidate(url="https://example.com/new-location")
    lookup = FakeFingerprintLookup(matching_title_hash=DedupeService.title_hash(candidate))

    result = asyncio.run(DedupeService(lookup).find_duplicate(candidate))

    assert result is DuplicateReason.NORMALIZED_TITLE
    assert lookup.checks == ["url", "title"]


def test_soft_deleted_article_still_matches_by_canonical_url_before_title() -> None:
    """A deleted Article's URL fingerprint must suppress re-ingestion without title fallback."""
    candidate = _candidate(url="https://example.com/suppressed")
    lookup = FakeFingerprintLookup(matching_url_hash=DedupeService.url_hash(candidate))

    result = asyncio.run(DedupeService(lookup).find_duplicate(candidate))

    assert result is DuplicateReason.CANONICAL_URL
    assert lookup.checks == ["url"]


def test_source_candidates_are_capped_at_ten() -> None:
    """An adapter returning more than ten candidates must not exceed the source quota."""
    candidates = [_candidate(url=f"https://example.com/{index}") for index in range(11)]

    limited = limit_candidates(candidates)

    assert len(limited) == MAX_CANDIDATES_PER_SOURCE
    assert [candidate.canonical_url for candidate in limited] == [
        f"https://example.com/{index}" for index in range(10)
    ]
