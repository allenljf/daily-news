"""PostgreSQL-backed HTTP contracts for manual Ingestion Runs."""

from __future__ import annotations

import uuid
from collections.abc import AsyncIterator
from datetime import date

import pytest
from fastapi.testclient import TestClient
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from app.db.engine import get_session
from app.identity.dependencies import require_allowed_identity
from app.identity.firebase import VerifiedIdentity
from app.ingestion.router import get_job_launcher
from app.main import create_app
from tests.postgres import psql, run_alembic_upgrade


class FakeJobLauncher:
    """A test double that records launch requests without calling Cloud Run."""

    def __init__(self) -> None:
        self.run_ids: list[uuid.UUID] = []

    async def launch(self, run_id: uuid.UUID) -> None:
        self.run_ids.append(run_id)


@pytest.fixture
def migrated_database(postgres_container: tuple[str, str]) -> tuple[str, str]:
    """Apply production migrations to an isolated PostgreSQL database."""
    run_alembic_upgrade(postgres_container[1])
    return postgres_container


@pytest.fixture
def authenticated_client(
    migrated_database: tuple[str, str],
) -> tuple[TestClient, FakeJobLauncher]:
    """Use the production app, real PostgreSQL, fake identity, and fake job launcher."""
    _, database_url = migrated_database
    session_factory = async_sessionmaker(create_async_engine(database_url), expire_on_commit=False)
    app = create_app()
    launcher = FakeJobLauncher()

    async def fake_identity() -> VerifiedIdentity:
        return VerifiedIdentity(uid="test-user", email="reader@example.com")

    async def override_session() -> AsyncIterator[AsyncSession]:
        async with session_factory() as session:
            yield session

    app.dependency_overrides[require_allowed_identity] = fake_identity
    app.dependency_overrides[get_session] = override_session
    app.dependency_overrides[get_job_launcher] = lambda: launcher
    with TestClient(app) as client:
        yield client, launcher


def test_latest_run_has_no_success_timestamp_before_any_success(
    authenticated_client: tuple[TestClient, FakeJobLauncher],
) -> None:
    """A regression: an empty database must report null, not an invented update timestamp."""
    client, _ = authenticated_client

    response = client.get("/v1/ingestion-runs/latest")

    assert response.status_code == 200
    assert response.json() == {"last_successful_at": None, "active_run": None}


def test_manual_run_returns_202_and_merges_an_existing_active_run(
    authenticated_client: tuple[TestClient, FakeJobLauncher],
) -> None:
    """A regression: a second active Run must not launch another background Job."""
    client, launcher = authenticated_client

    first_response = client.post("/v1/ingestion-runs")
    second_response = client.post("/v1/ingestion-runs")

    assert first_response.status_code == 202
    assert first_response.json()["trigger"] == "manual"
    assert first_response.json()["status"] == "queued"
    assert second_response.status_code == 202
    assert second_response.json()["id"] == first_response.json()["id"]
    assert launcher.run_ids == [uuid.UUID(first_response.json()["id"])]


def test_completed_scheduled_run_does_not_block_a_new_manual_run(
    authenticated_client: tuple[TestClient, FakeJobLauncher],
    migrated_database: tuple[str, str],
) -> None:
    """A regression: only queued or running Runs can block a manual trigger."""
    client, launcher = authenticated_client
    container_name, _ = migrated_database
    scheduled_run_id = uuid.uuid4()
    psql(
        container_name,
        "INSERT INTO ingestion_runs "
        "(id, trigger, idempotency_key, taipei_date, started_at, finished_at, status) "
        f"VALUES ('{scheduled_run_id}', 'scheduled', 'scheduled:2026-08-30', "
        f"'{date(2026, 8, 30)}', '2026-08-30T00:00:00+00:00', "
        "'2026-08-30T00:01:00+00:00', 'succeeded');",
    )

    response = client.post("/v1/ingestion-runs")
    latest = client.get("/v1/ingestion-runs/latest")

    assert response.status_code == 202
    assert response.json()["trigger"] == "manual"
    assert response.json()["id"] != str(scheduled_run_id)
    assert launcher.run_ids == [uuid.UUID(response.json()["id"])]
    assert latest.status_code == 200
    assert latest.json()["last_successful_at"] == "2026-08-30T00:01:00Z"
