"""CLI entrypoint for a Cloud Run daily-news ingestion Job."""

from __future__ import annotations

import argparse
import os
import uuid

from sqlalchemy.ext.asyncio import AsyncSession

from app.ingestion.orchestrator import IngestionOrchestrator, SourceWork, SqlAlchemyIngestionStore


async def run_job(
    run_id: uuid.UUID,
    session: AsyncSession,
    work_items: list[SourceWork],
) -> None:
    """Execute a pre-composed ingestion batch in one database transaction."""
    async with session.begin():
        await IngestionOrchestrator(SqlAlchemyIngestionStore(session)).run_ingestion(
            str(run_id),
            work_items,
        )


def main() -> None:
    parser = argparse.ArgumentParser(description="Run Daily News ingestion for one Ingestion Run.")
    parser.add_argument(
        "--run-id",
        default=os.getenv("RUN_ID"),
        help="Ingestion Run UUID (or RUN_ID).",
    )
    args = parser.parse_args()
    if not args.run_id:
        parser.error("--run-id or RUN_ID is required")


if __name__ == "__main__":
    main()
