"""PostgreSQL-backed HTTP contracts for reading and managing Articles."""

from __future__ import annotations

import uuid
from collections.abc import AsyncIterator
from datetime import datetime, timedelta, timezone

import pytest
from fastapi.testclient import TestClient
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from app.db.engine import get_session
from app.identity.dependencies import require_allowed_identity
from app.identity.firebase import VerifiedIdentity
from app.main import create_app
from tests.postgres import psql, run_alembic_upgrade


@pytest.fixture
def migrated_database(postgres_container: tuple[str, str]) -> tuple[str, str]:
    """Apply production migrations to an isolated PostgreSQL database."""
    run_alembic_upgrade(postgres_container[1])
    return postgres_container


@pytest.fixture
def authenticated_client(migrated_database: tuple[str, str]) -> TestClient:
    """Use the production application with real PostgreSQL and a fake identity."""
    _, database_url = migrated_database
    session_factory = async_sessionmaker(create_async_engine(database_url), expire_on_commit=False)
    app = create_app()

    async def fake_identity() -> VerifiedIdentity:
        return VerifiedIdentity(uid="test-user", email="reader@example.com")

    async def override_session() -> AsyncIterator[AsyncSession]:
        async with session_factory() as session:
            yield session

    app.dependency_overrides[require_allowed_identity] = fake_identity
    app.dependency_overrides[get_session] = override_session
    with TestClient(app) as client:
        yield client


def _create_category(client: TestClient, name: str, source_labels: list[str]) -> dict[str, object]:
    response = client.post(
        "/v1/categories",
        json={
            "name": name,
            "source_settings": [
                {"label": label, "website_input": f"{label.lower()}.example", "kind": "website"}
                for label in source_labels
            ],
        },
    )
    assert response.status_code == 201
    return response.json()


def _insert_article(
    container_name: str,
    *,
    category_id: str,
    source_setting_id: str,
    title: str,
    position: int,
    expires_at: datetime | None = None,
    deleted_at: datetime | None = None,
) -> str:
    """Insert a hand-specified Article fixture and return its ID."""
    article_id = str(uuid.uuid4())
    inserted_at = datetime(2026, 8, 1, tzinfo=timezone.utc) + timedelta(minutes=position)
    expires_sql = "NULL" if expires_at is None else f"'{expires_at.isoformat()}'"
    deleted_sql = "NULL" if deleted_at is None else f"'{deleted_at.isoformat()}'"
    psql(
        container_name,
        f"""
        INSERT INTO articles (
          id, title, normalized_title_hash, canonical_url, canonical_url_hash,
          first_seen_at, expires_at, deleted_at
        ) VALUES (
          '{article_id}', '{title}', '{position:064x}',
          'https://example.com/{position}', '{(position + 1000):064x}',
          '{inserted_at.isoformat()}', {expires_sql}, {deleted_sql}
        );
        INSERT INTO category_articles (category_id, article_id, source_setting_id, inserted_at)
        VALUES (
          '{category_id}', '{article_id}', '{source_setting_id}', '{inserted_at.isoformat()}'
        );
        """,
    )
    return article_id


def test_news_page_returns_twenty_items_then_continues_from_opaque_cursor(
    authenticated_client: TestClient,
    migrated_database: tuple[str, str],
) -> None:
    """A regression: an offset, 21-item page, or unstable cursor must fail this contract."""
    container_name, _ = migrated_database
    category = _create_category(authenticated_client, "AI", ["Open web"])
    category_id = category["id"]
    source_id = category["source_settings"][0]["id"]
    expected_titles = [f"Article {position}" for position in range(21, 0, -1)]
    for position in range(1, 22):
        _insert_article(
            container_name,
            category_id=category_id,
            source_setting_id=source_id,
            title=f"Article {position}",
            position=position,
            expires_at=datetime(2030, 1, 1, tzinfo=timezone.utc),
        )

    first_page = authenticated_client.get(f"/v1/categories/{category_id}/news")

    assert first_page.status_code == 200
    first_body = first_page.json()
    assert [item["title"] for item in first_body["items"]] == expected_titles[:20]
    assert first_body["next_cursor"]
    second_page = authenticated_client.get(
        f"/v1/categories/{category_id}/news",
        params={"cursor": first_body["next_cursor"]},
    )
    assert second_page.status_code == 200
    assert [item["title"] for item in second_page.json()["items"]] == expected_titles[20:]
    assert second_page.json()["next_cursor"] is None


def test_news_filter_excludes_expired_and_deleted_articles_and_matches_source_tag(
    authenticated_client: TestClient,
    migrated_database: tuple[str, str],
) -> None:
    """A regression: leaking an expired, deleted, or wrong-source Article must fail."""
    container_name, _ = migrated_database
    category = _create_category(authenticated_client, "AI", ["Open web", "GitHub"])
    category_id = category["id"]
    open_web_id = category["source_settings"][0]["id"]
    github_id = category["source_settings"][1]["id"]
    _insert_article(
        container_name,
        category_id=category_id,
        source_setting_id=open_web_id,
        title="Open web article",
        position=1,
        expires_at=datetime(2030, 1, 1, tzinfo=timezone.utc),
    )
    _insert_article(
        container_name,
        category_id=category_id,
        source_setting_id=github_id,
        title="GitHub article",
        position=2,
        expires_at=datetime(2030, 1, 1, tzinfo=timezone.utc),
    )
    _insert_article(
        container_name,
        category_id=category_id,
        source_setting_id=github_id,
        title="Expired article",
        position=3,
        expires_at=datetime(2020, 1, 1, tzinfo=timezone.utc),
    )
    _insert_article(
        container_name,
        category_id=category_id,
        source_setting_id=github_id,
        title="Deleted article",
        position=4,
        expires_at=datetime(2030, 1, 1, tzinfo=timezone.utc),
        deleted_at=datetime(2026, 8, 1, tzinfo=timezone.utc),
    )

    response = authenticated_client.get(
        f"/v1/categories/{category_id}/news",
        params={"sourceTagId": github_id},
    )

    assert response.status_code == 200
    assert [item["title"] for item in response.json()["items"]] == ["GitHub article"]


def test_article_permanence_and_deletion_apply_to_every_category(
    authenticated_client: TestClient,
    migrated_database: tuple[str, str],
) -> None:
    """A regression: Article mutations must not be limited to one Category Article link."""
    container_name, _ = migrated_database
    first_category = _create_category(authenticated_client, "AI", ["Open web"])
    second_category = _create_category(authenticated_client, "Security", ["GitHub"])
    article_id = _insert_article(
        container_name,
        category_id=first_category["id"],
        source_setting_id=first_category["source_settings"][0]["id"],
        title="Shared article",
        position=1,
        expires_at=datetime(2030, 1, 1, tzinfo=timezone.utc),
    )
    psql(
        container_name,
        "INSERT INTO category_articles (category_id, article_id, source_setting_id, inserted_at) "
        f"VALUES ('{second_category['id']}', '{article_id}', "
        f"'{second_category['source_settings'][0]['id']}', '2026-08-01T00:02:00+00:00');",
    )

    permanent = authenticated_client.patch(f"/v1/news/{article_id}", json={"permanent": True})

    assert permanent.status_code == 200
    assert permanent.json()["expires_at"] is None
    deleted = authenticated_client.delete(f"/v1/news/{article_id}")
    assert deleted.status_code == 204
    first_feed = authenticated_client.get(f"/v1/categories/{first_category['id']}/news")
    second_feed = authenticated_client.get(f"/v1/categories/{second_category['id']}/news")
    assert first_feed.json()["items"] == []
    assert second_feed.json()["items"] == []


def test_news_routes_reject_invalid_cursor_and_unknown_article(
    authenticated_client: TestClient,
) -> None:
    """A regression: malformed cursors need 400 and missing Article mutations need 404."""
    category = _create_category(authenticated_client, "AI", ["Open web"])

    invalid_cursor = authenticated_client.get(
        f"/v1/categories/{category['id']}/news",
        params={"cursor": "not-a-valid-cursor"},
    )
    missing_article = authenticated_client.patch(
        f"/v1/news/{uuid.uuid4()}",
        json={"permanent": True},
    )

    assert invalid_cursor.status_code == 400
    assert invalid_cursor.headers["content-type"] == "application/problem+json"
    assert missing_article.status_code == 404
    assert missing_article.headers["content-type"] == "application/problem+json"
