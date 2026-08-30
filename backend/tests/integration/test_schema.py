import json
import os
import secrets
import subprocess
import time
from collections.abc import Iterator
from pathlib import Path

import pytest

BACKEND_ROOT = Path(__file__).resolve().parents[2]
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


def _container_port(container_name: str) -> str:
    result = _docker("port", container_name, "5432/tcp")
    return result.stdout.strip().rsplit(":", maxsplit=1)[-1]


def _psql(container_name: str, sql: str) -> str:
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
        "-At",
        "-F",
        ",",
        "-c",
        sql,
    )
    return result.stdout.strip()


@pytest.fixture
def postgres_container() -> Iterator[tuple[str, str]]:
    container_name = f"daily-news-schema-test-{secrets.token_hex(4)}"
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
        port = _container_port(container_name)
        database_url = (
            f"postgresql+asyncpg://{POSTGRES_USER}:{POSTGRES_PASSWORD}"
            f"@127.0.0.1:{port}/{POSTGRES_DB}"
        )
        yield container_name, database_url
    finally:
        _docker("rm", "-f", container_name, check=False)


def test_initial_migration_creates_daily_news_schema(
    postgres_container: tuple[str, str],
) -> None:
    container_name, database_url = postgres_container
    environment = os.environ | {"DATABASE_URL": database_url}

    subprocess.run(
        ["uv", "run", "alembic", "upgrade", "head"],
        cwd=BACKEND_ROOT,
        env=environment,
        text=True,
        capture_output=True,
        check=True,
    )

    tables = {
        row.split(",")[0]
        for row in _psql(
            container_name,
            """
            SELECT tablename
            FROM pg_tables
            WHERE schemaname = 'public'
            ORDER BY tablename
            """,
        ).splitlines()
        if row
    }
    assert tables == {
        "alembic_version",
        "articles",
        "categories",
        "category_articles",
        "ingestion_attempts",
        "ingestion_runs",
        "source_settings",
    }

    indexes = {
        row[0]: {"unique": row[1] == "t", "definition": row[2]}
        for row in (
            line.split(",", maxsplit=2)
            for line in _psql(
                container_name,
                """
                SELECT indexname, indexdef LIKE '% UNIQUE INDEX %', indexdef
                FROM pg_indexes
                WHERE schemaname = 'public'
                  AND tablename IN (
                    'articles',
                    'category_articles',
                    'ingestion_runs'
                  )
                ORDER BY indexname
                """,
            ).splitlines()
            if line
        )
    }

    assert indexes["ix_articles_canonical_url_hash"]["unique"] is True
    assert "(canonical_url_hash)" in indexes["ix_articles_canonical_url_hash"]["definition"]
    assert indexes["ix_articles_normalized_title_hash"]["unique"] is False
    assert "(normalized_title_hash)" in indexes["ix_articles_normalized_title_hash"]["definition"]
    assert indexes["ix_category_articles_feed_cursor"]["unique"] is False
    assert (
        "(category_id, inserted_at DESC, article_id DESC)"
        in indexes["ix_category_articles_feed_cursor"]["definition"]
    )
    assert indexes["ix_category_articles_category_article"]["unique"] is True
    assert (
        "(category_id, article_id)"
        in indexes["ix_category_articles_category_article"]["definition"]
    )
    assert indexes["ix_ingestion_runs_idempotency_key"]["unique"] is True
    assert "(idempotency_key)" in indexes["ix_ingestion_runs_idempotency_key"]["definition"]

    trigger_and_status = _psql(
        container_name,
        """
        SELECT json_agg(column_name ORDER BY ordinal_position)
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'ingestion_runs'
          AND column_name IN ('trigger', 'idempotency_key', 'status')
        """,
    )
    assert json.loads(trigger_and_status) == ["trigger", "idempotency_key", "status"]
