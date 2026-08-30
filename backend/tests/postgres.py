"""Reusable disposable PostgreSQL support for backend integration tests."""

from __future__ import annotations

import os
import secrets
import subprocess
import time
from collections.abc import Iterator
from pathlib import Path

import pytest

BACKEND_ROOT = Path(__file__).resolve().parents[1]
POSTGRES_IMAGE = "postgres:16-alpine"
POSTGRES_USER = "daily_news"
POSTGRES_PASSWORD = "daily_news"
POSTGRES_DB = "daily_news_test"


def _docker(*args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["docker", *args],
        cwd=BACKEND_ROOT,
        text=True,
        capture_output=True,
        check=check,
    )


def _wait_for_postgres(container_name: str) -> None:
    deadline = time.time() + 30
    while time.time() < deadline:
        result = _docker(
            "exec",
            container_name,
            "pg_isready",
            "-U",
            POSTGRES_USER,
            "-d",
            POSTGRES_DB,
            check=False,
        )
        if result.returncode == 0:
            return
        time.sleep(1)
    pytest.fail("Timed out waiting for PostgreSQL test container to become ready.")


def psql(container_name: str, sql: str) -> str:
    """Execute SQL in a disposable test database and return compact output."""
    result = _docker(
        "exec",
        "-e",
        f"PGPASSWORD={POSTGRES_PASSWORD}",
        container_name,
        "psql",
        "-U",
        POSTGRES_USER,
        "-d",
        POSTGRES_DB,
        "-v",
        "ON_ERROR_STOP=1",
        "-At",
        "-F",
        ",",
        "-c",
        sql,
    )
    return result.stdout.strip()


def run_alembic_upgrade(database_url: str) -> None:
    """Apply the production migration chain to an isolated test database."""
    subprocess.run(
        ["uv", "run", "alembic", "upgrade", "head"],
        cwd=BACKEND_ROOT,
        env=os.environ | {"DATABASE_URL": database_url},
        text=True,
        capture_output=True,
        check=True,
    )


@pytest.fixture
def postgres_container() -> Iterator[tuple[str, str]]:
    """Start an isolated PostgreSQL 16 database and remove it after each test."""
    container_name = f"daily-news-test-{secrets.token_hex(4)}"
    _docker(
        "run",
        "--rm",
        "--detach",
        "--publish",
        "127.0.0.1::5432",
        "--name",
        container_name,
        "-e",
        f"POSTGRES_USER={POSTGRES_USER}",
        "-e",
        f"POSTGRES_PASSWORD={POSTGRES_PASSWORD}",
        "-e",
        f"POSTGRES_DB={POSTGRES_DB}",
        POSTGRES_IMAGE,
    )
    try:
        _wait_for_postgres(container_name)
        port = _docker("port", container_name, "5432/tcp").stdout.strip().rsplit(":", 1)[-1]
        yield (
            container_name,
            f"postgresql+asyncpg://{POSTGRES_USER}:{POSTGRES_PASSWORD}@127.0.0.1:"
            f"{port}/{POSTGRES_DB}",
        )
    finally:
        _docker("rm", "-f", container_name, check=False)
