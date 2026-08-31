"""Coordinates source isolation, candidate quotas, and Ingestion Run accounting."""

from __future__ import annotations

import uuid
from dataclasses import dataclass
from datetime import datetime, timedelta, timezone
from typing import Protocol

from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.models import Article, CategoryArticle, IngestionAttempt, IngestionRun
from app.ingestion.normalization import normalized_title_hash, url_hash
from app.ingestion.sources import MAX_CANDIDATES_PER_SOURCE, CandidateArticle


class SourceAdapter(Protocol):
    async def search(self, request: object) -> list[CandidateArticle]: ...


class IngestionStore(Protocol):
    async def mark_running(self, run_id: str) -> None: ...
    async def record_candidate(
        self,
        category_id: str,
        source_id: str,
        candidate: CandidateArticle,
    ) -> bool: ...
    async def record_attempt(self, category_id: str, source_id: str, status: str) -> None: ...
    async def finish(self, run_id: str, status: str, counts: "IngestionResult") -> None: ...


@dataclass(frozen=True)
class SourceWork:
    category_id: str
    source_id: str
    adapter: SourceAdapter


@dataclass(frozen=True)
class IngestionResult:
    candidate_count: int = 0
    inserted_count: int = 0
    duplicate_count: int = 0
    error_count: int = 0


class IngestionOrchestrator:
    def __init__(self, store: IngestionStore) -> None:
        self._store = store

    async def run_ingestion(self, run_id: str, work_items: list[SourceWork]) -> IngestionResult:
        await self._store.mark_running(run_id)
        counts = IngestionResult()
        for work in work_items:
            try:
                candidates = (await work.adapter.search(work))[:MAX_CANDIDATES_PER_SOURCE]
                inserted = 0
                duplicates = 0
                for candidate in candidates:
                    if await self._store.record_candidate(
                        work.category_id,
                        work.source_id,
                        candidate,
                    ):
                        inserted += 1
                    else:
                        duplicates += 1
                await self._store.record_attempt(work.category_id, work.source_id, "succeeded")
                counts = IngestionResult(
                    counts.candidate_count + len(candidates), counts.inserted_count + inserted,
                    counts.duplicate_count + duplicates, counts.error_count,
                )
            except Exception:
                await self._store.record_attempt(work.category_id, work.source_id, "failed")
                counts = IngestionResult(
                    counts.candidate_count, counts.inserted_count, counts.duplicate_count,
                    counts.error_count + 1,
                )
        await self._store.finish(run_id, "succeeded", counts)
        return counts


class SqlAlchemyIngestionStore:
    """Writes Articles, Category Articles, Attempts, and final Run counters in one session."""

    def __init__(self, session: AsyncSession) -> None:
        self._session = session
        self._run_id: str | None = None

    async def mark_running(self, run_id: str) -> None:
        run = await self._session.get(IngestionRun, uuid.UUID(run_id))
        if run is None:
            raise ValueError("Ingestion Run not found")
        if run.status == "queued":
            run.status = "running"
        self._run_id = run_id

    async def record_candidate(
        self,
        category_id: str,
        source_id: str,
        candidate: CandidateArticle,
    ) -> bool:
        canonical_hash = url_hash(candidate.canonical_url)
        title_hash = normalized_title_hash(candidate.title)
        article = await self._session.scalar(
            select(Article).where(Article.canonical_url_hash == canonical_hash)
        )
        if article is None:
            article = await self._session.scalar(
                select(Article).where(Article.normalized_title_hash == title_hash)
            )
        created = article is None
        if article is None:
            now = datetime.now(timezone.utc)
            article = Article(
                title=candidate.title,
                normalized_title_hash=title_hash,
                canonical_url=candidate.canonical_url,
                canonical_url_hash=canonical_hash,
                summary=candidate.summary,
                published_at=candidate.published_at,
                first_seen_at=now,
                expires_at=now + timedelta(days=30),
            )
            self._session.add(article)
            await self._session.flush()
        link = await self._session.get(
            CategoryArticle,
            {"category_id": uuid.UUID(category_id), "article_id": article.id},
        )
        if link is None:
            self._session.add(
                CategoryArticle(
                    category_id=uuid.UUID(category_id),
                    article_id=article.id,
                    source_setting_id=uuid.UUID(source_id),
                )
            )
        return created

    async def record_attempt(self, category_id: str, source_id: str, status: str) -> None:
        if self._run_id is None:
            raise RuntimeError("mark_running must be called before recording attempts")
        self._session.add(
            IngestionAttempt(
                run_id=uuid.UUID(self._run_id),
                category_id=uuid.UUID(category_id),
                source_setting_id=uuid.UUID(source_id),
                status=status,
            )
        )

    async def finish(self, run_id: str, status: str, counts: IngestionResult) -> None:
        run = await self._session.get(IngestionRun, uuid.UUID(run_id))
        if run is None:
            raise ValueError("Ingestion Run not found")
        run.status = status
        run.finished_at = datetime.now(timezone.utc)
        run.candidate_count = counts.candidate_count
        run.inserted_count = counts.inserted_count
        run.duplicate_count = counts.duplicate_count
        run.error_count = counts.error_count
