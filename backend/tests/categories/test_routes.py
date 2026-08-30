"""PostgreSQL-backed contract tests for category settings routes."""

from __future__ import annotations

import uuid
from collections.abc import AsyncIterator

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
    """Apply the real Alembic schema to an isolated PostgreSQL database."""
    run_alembic_upgrade(postgres_container[1])
    return postgres_container


@pytest.fixture
def authenticated_client(migrated_database: tuple[str, str]) -> TestClient:
    """Use the production app with a fake identity and real async PostgreSQL session."""
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


def category_payload() -> dict[str, object]:
    return {
        "name": "AI research",
        "search_keywords": "agents",
        "special_requirements": "prefer primary sources",
        "source_settings": [
            {"label": "Open web", "website_input": "", "kind": "unspecified"},
            {"label": "OpenAI", "website_input": "openai.com", "kind": "website"},
            {"label": "GitHub", "website_input": "https://github.com", "kind": "website"},
        ],
    }


def test_create_category_returns_ordered_source_settings(authenticated_client: TestClient) -> None:
    """A create regression: removing source persistence or ordering fails this contract."""
    response = authenticated_client.post("/v1/categories", json=category_payload())

    assert response.status_code == 201
    body = response.json()
    assert body["name"] == "AI research"
    assert [item["label"] for item in body["source_settings"]] == [
        "Open web",
        "OpenAI",
        "GitHub",
    ]
    assert [item["position"] for item in body["source_settings"]] == [0, 1, 2]


def test_create_category_rejects_blank_name(authenticated_client: TestClient) -> None:
    """A validation regression: accepting an all-whitespace category name is invalid."""
    payload = category_payload() | {"name": "   "}

    response = authenticated_client.post("/v1/categories", json=payload)

    assert response.status_code == 422


def test_update_category_replaces_and_reorders_source_settings(
    authenticated_client: TestClient,
) -> None:
    """A replacement regression: stale active settings or wrong positions fail this contract."""
    created = authenticated_client.post("/v1/categories", json=category_payload()).json()
    update_payload = {
        "name": "AI research updates",
        "search_keywords": "models",
        "special_requirements": None,
        "source_settings": [
            {"label": "GitHub", "website_input": "https://github.com", "kind": "website"},
            {"label": "Open web", "website_input": "", "kind": "unspecified"},
        ],
    }

    response = authenticated_client.patch(
        f"/v1/categories/{created['id']}", json=update_payload
    )

    assert response.status_code == 200
    assert [item["label"] for item in response.json()["source_settings"]] == ["GitHub", "Open web"]
    listed = authenticated_client.get("/v1/categories")
    assert listed.status_code == 200
    assert listed.json()[0]["name"] == "AI research updates"
    assert [item["position"] for item in listed.json()[0]["source_settings"]] == [0, 1]


def test_delete_category_soft_deletes_links_without_deleting_article(
    authenticated_client: TestClient,
    migrated_database: tuple[str, str],
) -> None:
    """A deletion regression: category cleanup must never remove an Article row."""
    container_name, _ = migrated_database
    created = authenticated_client.post("/v1/categories", json=category_payload()).json()
    article_id = uuid.uuid4()
    source_setting_id = created["source_settings"][0]["id"]
    category_id = created["id"]
    psql(
        container_name,
        "INSERT INTO articles "
        "(id, title, normalized_title_hash, canonical_url, canonical_url_hash) "
        f"VALUES ('{article_id}', 'Kept article', '{'a' * 64}', 'https://example.com/a', "
        f"'{'b' * 64}'); "
        "INSERT INTO category_articles (category_id, article_id, source_setting_id) "
        f"VALUES ('{category_id}', '{article_id}', '{source_setting_id}');",
    )

    response = authenticated_client.delete(f"/v1/categories/{category_id}")

    assert response.status_code == 204
    deletion_state = psql(
        container_name,
        f"""
        SELECT
          (SELECT deleted_at IS NOT NULL FROM categories WHERE id = '{category_id}'),
          (SELECT count(*) = 3 AND bool_and(deleted_at IS NOT NULL)
             FROM source_settings WHERE category_id = '{category_id}'),
          (SELECT deleted_at IS NOT NULL FROM category_articles
             WHERE category_id = '{category_id}' AND article_id = '{article_id}'),
          (SELECT count(*) FROM articles WHERE id = '{article_id}')
        """,
    )
    assert deletion_state == "t,t,t,1"
    second_delete = authenticated_client.delete(f"/v1/categories/{category_id}")
    assert second_delete.status_code == 404
    assert second_delete.headers["content-type"] == "application/problem+json"
    assert authenticated_client.get("/v1/categories").json() == []
    patch_after_delete = authenticated_client.patch(
        f"/v1/categories/{category_id}", json=category_payload()
    )
    assert patch_after_delete.status_code == 404
    assert patch_after_delete.headers["content-type"] == "application/problem+json"


def test_categories_require_authorization(migrated_database: tuple[str, str]) -> None:
    """An auth regression: category routes cannot be public."""
    app = create_app()
    with TestClient(app) as client:
        response = client.get("/v1/categories")

    assert response.status_code == 401
