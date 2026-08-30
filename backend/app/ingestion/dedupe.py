"""URL-first Article duplicate detection independent of a persistence implementation."""

from __future__ import annotations

from enum import Enum
from typing import Protocol

from app.ingestion.normalization import normalized_title_hash, url_hash
from app.ingestion.sources import CandidateArticle


class ArticleFingerprintLookup(Protocol):
    """Reads Article fingerprints, including fingerprints belonging to soft-deleted Articles."""

    async def has_canonical_url_hash(self, fingerprint: str) -> bool: ...

    async def has_normalized_title_hash(self, fingerprint: str) -> bool: ...


class DuplicateReason(str, Enum):
    """The fingerprint that caused an Article candidate to be suppressed."""

    CANONICAL_URL = "canonical_url"
    NORMALIZED_TITLE = "normalized_title"


class DedupeService:
    """Checks canonical URLs before title fallbacks to preserve the strongest identity signal."""

    def __init__(self, lookup: ArticleFingerprintLookup) -> None:
        self._lookup = lookup

    async def find_duplicate(self, candidate: CandidateArticle) -> DuplicateReason | None:
        """Return why a candidate is already known, or None when it is novel."""
        if await self._lookup.has_canonical_url_hash(self.url_hash(candidate)):
            return DuplicateReason.CANONICAL_URL
        if await self._lookup.has_normalized_title_hash(self.title_hash(candidate)):
            return DuplicateReason.NORMALIZED_TITLE
        return None

    @staticmethod
    def url_hash(candidate: CandidateArticle) -> str:
        """Expose the exact URL fingerprint used for duplicate lookup tests and adapters."""
        return url_hash(candidate.canonical_url)

    @staticmethod
    def title_hash(candidate: CandidateArticle) -> str:
        """Expose the exact title fingerprint used for duplicate lookup tests and adapters."""
        return normalized_title_hash(candidate.title)
