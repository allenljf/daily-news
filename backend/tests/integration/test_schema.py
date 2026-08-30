"""PostgreSQL schema migration contract tests."""

import json

from tests.postgres import psql, run_alembic_upgrade


def test_initial_migration_creates_daily_news_schema(
    postgres_container: tuple[str, str],
) -> None:
    container_name, database_url = postgres_container
    run_alembic_upgrade(database_url)

    tables = {
        row.split(",")[0]
        for row in psql(
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
            for line in psql(
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

    trigger_and_status = psql(
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
