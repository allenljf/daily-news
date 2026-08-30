"""Application service for the Ingestion Run state machine."""

from __future__ import annotations

import uuid
from datetime import datetime, timezone
from typing import Protocol
from zoneinfo import ZoneInfo

from sqlalchemy.ext.asyncio import AsyncSession

from app.db.models import IngestionRun
from app.ingestion.run_repository import ACTIVE_RUN_STATUSES, IngestionRunRepository
from app.ingestion.schemas import IngestionRunResponse, LatestIngestionRunResponse

TAIPEI_TIMEZONE = ZoneInfo("Asia/Taipei")


class JobLauncher(Protocol):
    """Starts the Cloud Run Job that will execute a queued Ingestion Run."""

    async def launch(self, run_id: uuid.UUID) -> None: ...


class NoopJobLauncher:
    """Local-development launcher; deployment replaces it with the Cloud Run adapter."""

    async def launch(self, run_id: uuid.UUID) -> None:
        del run_id


class IngestionRunService:
    """Coordinates active-run merging and explicit lifecycle transitions."""

    def __init__(self, session: AsyncSession) -> None:
        self._session = session
        self._repository = IngestionRunRepository(session)

    async def latest(self) -> LatestIngestionRunResponse:
        active_run = await self._repository.find_active()
        return LatestIngestionRunResponse(
            last_successful_at=await self._repository.latest_successful_at(),
            active_run=self._response(active_run) if active_run else None,
        )

    async def request_manual_run(self, launcher: JobLauncher) -> IngestionRunResponse:
        async with self._session.begin():
            await self._repository.lock_active_run_creation()
            active_run = await self._repository.find_active()
            if active_run is not None:
                return self._response(active_run)
            request_id = uuid.uuid4()
            run = self._repository.add_manual(
                idempotency_key=f"manual:{request_id}",
                taipei_date=datetime.now(TAIPEI_TIMEZONE).date(),
            )
            await self._session.flush()
            response = self._response(run)
        await launcher.launch(run.id)
        return response

    async def mark_running(self, run_id: uuid.UUID) -> IngestionRun | None:
        return await self._transition(run_id, from_status="queued", to_status="running")

    async def mark_succeeded(self, run_id: uuid.UUID) -> IngestionRun | None:
        return await self._transition(run_id, from_status="running", to_status="succeeded")

    async def mark_failed(self, run_id: uuid.UUID) -> IngestionRun | None:
        return await self._transition(run_id, from_status="running", to_status="failed")

    async def _transition(
        self,
        run_id: uuid.UUID,
        *,
        from_status: str,
        to_status: str,
    ) -> IngestionRun | None:
        async with self._session.begin():
            run = await self._repository.find_by_id_for_update(run_id)
            if run is None or run.status != from_status:
                return None
            run.status = to_status
            if to_status not in ACTIVE_RUN_STATUSES:
                run.finished_at = datetime.now(timezone.utc)
        return run

    @staticmethod
    def _response(run: IngestionRun) -> IngestionRunResponse:
        return IngestionRunResponse(
            id=run.id,
            trigger=run.trigger,
            status=run.status,
            started_at=run.started_at,
            finished_at=run.finished_at,
        )
