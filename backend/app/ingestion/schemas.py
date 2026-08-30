"""HTTP DTOs for Ingestion Run status."""

from __future__ import annotations

import uuid
from datetime import datetime

from pydantic import BaseModel


class IngestionRunResponse(BaseModel):
    """The public state of one asynchronous Ingestion Run."""

    id: uuid.UUID
    trigger: str
    status: str
    started_at: datetime
    finished_at: datetime | None


class LatestIngestionRunResponse(BaseModel):
    """Homepage refresh metadata and any currently active background Run."""

    last_successful_at: datetime | None
    active_run: IngestionRunResponse | None
