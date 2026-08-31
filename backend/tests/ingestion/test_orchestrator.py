"""Contracts for daily Ingestion Run orchestration."""

from __future__ import annotations

import asyncio
import uuid
from dataclasses import dataclass

from sqlalchemy.ext.asyncio import async_sessionmaker, create_async_engine

from app.ingestion.orchestrator import IngestionOrchestrator, SourceWork, SqlAlchemyIngestionStore
from app.ingestion.sources import CandidateArticle
from tests.postgres import psql, run_alembic_upgrade


@dataclass
class FakeStore:
    inserted: list[tuple[str, str]]
    attempts: list[tuple[str, str]]
    status: str = "queued"

    async def mark_running(self, run_id: str) -> None:
        self.status = "running"

    async def record_candidate(
        self,
        category_id: str,
        source_id: str,
        candidate: CandidateArticle,
    ) -> bool:
        del category_id, source_id
        key = candidate.canonical_url
        if key in self.inserted:
            return False
        self.inserted.append(key)
        return True

    async def record_attempt(self, category_id: str, source_id: str, status: str) -> None:
        del category_id
        self.attempts.append((source_id, status))

    async def finish(self, run_id: str, status: str, counts: object) -> None:
        self.status = status


class FakeAdapter:
    def __init__(self, candidates: list[CandidateArticle] | Exception) -> None:
        self.candidates = candidates

    async def search(self, request: object) -> list[CandidateArticle]:
        if isinstance(self.candidates, Exception):
            raise self.candidates
        return self.candidates


def test_orchestrator_keeps_categories_links_and_survives_one_source_failure() -> None:
    """A source failure must not prevent other Category Articles or final Run accounting."""
    article = CandidateArticle("Shared", "https://example.com/a", "Example")
    store = FakeStore([], [])
    work = [
        SourceWork("category-a", "source-a", FakeAdapter([article] * 11)),
        SourceWork("category-b", "source-b", FakeAdapter([article])),
        SourceWork("category-b", "source-fail", FakeAdapter(RuntimeError("unavailable"))),
    ]

    result = asyncio.run(IngestionOrchestrator(store).run_ingestion("run-1", work))

    assert result.candidate_count == 11
    assert result.inserted_count == 1
    assert result.duplicate_count == 10
    assert result.error_count == 1
    assert store.status == "succeeded"
    assert store.attempts == [
        ("source-a", "succeeded"),
        ("source-b", "succeeded"),
        ("source-fail", "failed"),
    ]


def test_sqlalchemy_store_persists_links_attempts_and_run_counts(
    postgres_container: tuple[str, str],
) -> None:
    """One Article can link to two Categories without a duplicate database row."""
    container_name, database_url = postgres_container
    run_alembic_upgrade(database_url)
    category_a, category_b, source_a, source_b, run_id = (uuid.uuid4() for _ in range(5))
    psql(
        container_name,
        "INSERT INTO categories (id, name) VALUES "
        f"('{category_a}', 'A'), ('{category_b}', 'B'); "
        "INSERT INTO source_settings "
        "(id, category_id, label, website_input, kind, position) VALUES "
        f"('{source_a}', '{category_a}', 'A', '', 'unspecified', 0), "
        f"('{source_b}', '{category_b}', 'B', '', 'unspecified', 0); "
        "INSERT INTO ingestion_runs (id, trigger, idempotency_key, taipei_date, status) VALUES "
        f"('{run_id}', 'manual', 'manual:{run_id}', '2026-08-31', 'queued');",
    )
    article = CandidateArticle("Shared", "https://example.com/a", "Example")
    session_factory = async_sessionmaker(create_async_engine(database_url), expire_on_commit=False)

    async def execute() -> None:
        async with session_factory() as session:
            async with session.begin():
                await IngestionOrchestrator(SqlAlchemyIngestionStore(session)).run_ingestion(
                    str(run_id),
                    [
                        SourceWork(str(category_a), str(source_a), FakeAdapter([article])),
                        SourceWork(str(category_b), str(source_b), FakeAdapter([article])),
                    ],
                )

    asyncio.run(execute())

    assert psql(
        container_name,
        "SELECT status, candidate_count, inserted_count, duplicate_count, error_count "
        f"FROM ingestion_runs WHERE id = '{run_id}'; "
        f"SELECT count(*) FROM articles; SELECT count(*) FROM category_articles; "
        f"SELECT count(*) FROM ingestion_attempts WHERE run_id = '{run_id}';",
    ).splitlines() == ["succeeded,2,1,1,0", "1", "2", "2"]
