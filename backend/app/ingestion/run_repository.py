"""Transaction-safe persistence operations for Ingestion Runs."""

from __future__ import annotations

import uuid
from datetime import date, datetime

from sqlalchemy import select, text
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.models import IngestionRun

ACTIVE_RUN_STATUSES = ("queued", "running")
_ACTIVE_RUN_LOCK_ID = 824_706_321


class IngestionRunRepository:
    """Database adapter for the single globally active Ingestion Run."""

    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def lock_active_run_creation(self) -> None:
        """Acquire a transaction-scoped PostgreSQL lock before checking active state."""
        await self._session.execute(
            text("SELECT pg_advisory_xact_lock(:lock_id)"),
            {"lock_id": _ACTIVE_RUN_LOCK_ID},
        )

    async def find_active(self) -> IngestionRun | None:
        return await self._session.scalar(
            select(IngestionRun)
            .where(IngestionRun.status.in_(ACTIVE_RUN_STATUSES))
            .order_by(IngestionRun.started_at.desc())
            .limit(1)
        )

    def add_manual(self, *, idempotency_key: str, taipei_date: date) -> IngestionRun:
        run = IngestionRun(
            trigger="manual",
            idempotency_key=idempotency_key,
            taipei_date=taipei_date,
            status="queued",
        )
        self._session.add(run)
        return run

    async def find_by_id_for_update(self, run_id: uuid.UUID) -> IngestionRun | None:
        return await self._session.scalar(
            select(IngestionRun).where(IngestionRun.id == run_id).with_for_update()
        )

    async def latest_successful_at(self) -> datetime | None:
        return await self._session.scalar(
            select(IngestionRun.finished_at)
            .where(IngestionRun.status == "succeeded")
            .order_by(IngestionRun.finished_at.desc())
            .limit(1)
        )
