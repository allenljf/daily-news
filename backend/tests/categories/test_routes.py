"""PostgreSQL-backed contract tests for category settings routes."""

from __future__ import annotations

import os
import secrets
import subprocess
import time
import uuid
from collections.abc import AsyncIterator, Iterator
from pathlib import Path

import pytest
from fastapi.testclient import TestClient
from sqlalchemy.ext.asyncio import AsyncSession, async_sessionmaker, create_async_engine

from app.db.engine import get_session
from app.identity.dependencies import require_allowed_identity
from app.identity.firebase import VerifiedIdentity
from app.main import create_app

BACKEND_ROOT = Path(__file__).resolve().parents[2]


def _docker(*args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        ["docker", *args],
        cwd=BACKEND_ROOT,
        text=True,
        capture_output=True,
        check=check,
    )


@pytest.fixture
def postgres_container() -> Iterator[tuple[str, str]]:
    container_name = f"daily-news-categories-test-{secrets.token_hex(4)}"
    _docker(
        "run",
        "--rm",
        "--detach",
        "--publish",
        "127.0.0.1::5432",
        "--name",
        container_name,
        "-e",
        "POSTGRES_USER=daily_news",
        "-e",
        "POSTGRES_PASSWORD=daily_news",
        "-e",
        "POSTGRES_DB=daily_news_test",
        "postgres:16-alpine",
    )
    try:
        deadline = time.time() + 30
        while time.time() < deadline:
            ready = _docker(
                "exec",
                container_name,
                "pg_isready",
                "-U",
                "daily_news",
                "-d",
                "daily_news_test",
                check=False,
            )
            if ready.returncode == 0:
                break
            time.sleep(1)
        else:
            pytest.fail("Timed out waiting for PostgreSQL test container to become ready.")
        port = _docker("port", container_name, "5432/tcp").stdout.strip().rsplit(":", 1)[-1]
        yield (
            container_name,
            "postgresql+asyncpg://daily_news:daily_news@127.0.0.1:"
            f"{port}/daily_news_test",
        )
    finally:
        _docker("rm", "-f", container_name, check=False)


@pytest.fixture
def migrated_database(postgres_container: tuple[str, str]) -> tuple[str, str]:
    """Apply the real Alembic schema to an isolated PostgreSQL database."""
    container_name, database_url = postgres_container
    subprocess.run(
        ["uv", "run", "alembic", "upgrade", "head"],
        cwd=".",
        env=os.environ | {"DATABASE_URL": database_url},
        text=True,
        capture_output=True,
        check=True,
    )
    return container_name, database_url


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
    sql = (
        "INSERT INTO articles "
        "(id, title, normalized_title_hash, canonical_url, canonical_url_hash) "
        f"VALUES ('{article_id}', 'Kept article', '{'a' * 64}', 'https://example.com/a', "
        f"'{'b' * 64}'); "
        "INSERT INTO category_articles (category_id, article_id, source_setting_id) "
        f"VALUES ('{category_id}', '{article_id}', '{source_setting_id}');"
    )
    subprocess.run(
        [
            "docker",
            "exec",
            "-e",
            "PGPASSWORD=daily_news",
            container_name,
            "psql",
            "-U",
            "daily_news",
            "-d",
            "daily_news_test",
            "-v",
            "ON_ERROR_STOP=1",
            "-c",
            sql,
        ],
        check=True,
        text=True,
        capture_output=True,
    )

    response = authenticated_client.delete(f"/v1/categories/{category_id}")

    assert response.status_code == 204
    article_count = subprocess.run(
        [
            "docker",
            "exec",
            "-e",
            "PGPASSWORD=daily_news",
            container_name,
            "psql",
            "-U",
            "daily_news",
            "-d",
            "daily_news_test",
            "-At",
            "-c",
            f"SELECT count(*) FROM articles WHERE id = '{article_id}'",
        ],
        check=True,
        text=True,
        capture_output=True,
    )
    assert article_count.stdout.strip() == "1"
    second_delete = authenticated_client.delete(f"/v1/categories/{category_id}")
    assert second_delete.status_code == 404
    assert second_delete.headers["content-type"] == "application/problem+json"


def test_categories_require_authorization(migrated_database: tuple[str, str]) -> None:
    """An auth regression: category routes cannot be public."""
    app = create_app()
    with TestClient(app) as client:
        response = client.get("/v1/categories")

    assert response.status_code == 401
