"""Database-independent contracts shared by ingestion source adapters."""

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from typing import Protocol

MAX_CANDIDATES_PER_SOURCE = 10


@dataclass(frozen=True)
class SourceSearchRequest:
    """The Category and Source Setting context supplied to one adapter search."""

    category_name: str
    search_keywords: str | None
    source_label: str
    website_input: str
    special_requirements: str | None


@dataclass(frozen=True)
class CandidateArticle:
    """A source-provided candidate before database deduplication and persistence."""

    title: str
    canonical_url: str
    source_name: str
    summary: str | None = None
    citation_url: str | None = None
    published_at: datetime | None = None


class SourceAdapter(Protocol):
    """Retrieves at most ten Article candidates for one Source Setting."""

    async def search(self, request: SourceSearchRequest) -> list[CandidateArticle]: ...


def limit_candidates(candidates: list[CandidateArticle]) -> list[CandidateArticle]:
    """Enforce the per-Source Setting candidate quota at each adapter boundary."""
    return candidates[:MAX_CANDIDATES_PER_SOURCE]
