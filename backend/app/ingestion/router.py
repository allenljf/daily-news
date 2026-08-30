"""Authenticated HTTP routes for Ingestion Run status and manual triggering."""

from __future__ import annotations

from typing import Annotated

from fastapi import APIRouter, Depends, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.engine import get_session
from app.identity.dependencies import require_allowed_identity
from app.identity.firebase import VerifiedIdentity
from app.ingestion.run_service import IngestionRunService, JobLauncher, NoopJobLauncher
from app.ingestion.schemas import IngestionRunResponse, LatestIngestionRunResponse

router = APIRouter(prefix="/v1/ingestion-runs", tags=["ingestion-runs"])


def get_job_launcher() -> JobLauncher:
    """Provide the launcher seam until Cloud Run configuration is added in deployment work."""
    return NoopJobLauncher()


@router.get("/latest", response_model=LatestIngestionRunResponse)
async def get_latest_run(
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> LatestIngestionRunResponse:
    return await IngestionRunService(session).latest()


@router.post("", response_model=IngestionRunResponse, status_code=status.HTTP_202_ACCEPTED)
async def request_manual_run(
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
    launcher: Annotated[JobLauncher, Depends(get_job_launcher)],
) -> IngestionRunResponse:
    return await IngestionRunService(session).request_manual_run(launcher)
