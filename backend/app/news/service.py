"""Application service for Article feed reads and global lifecycle changes."""

from __future__ import annotations

import base64
import json
import uuid
from datetime import datetime, timedelta, timezone

from sqlalchemy.ext.asyncio import AsyncSession

from app.db.models import Article, CategoryArticle, SourceSetting
from app.news.repository import NewsRepository
from app.news.schemas import NewsDetail, NewsListItem, NewsPage

PAGE_SIZE = 20


class InvalidCursorError(ValueError):
    """Raised when a client-supplied feed cursor cannot be decoded safely."""


class NewsService:
    """Coordinates visible feed reads and Article-global mutations."""

    def __init__(self, session: AsyncSession) -> None:
        self._session = session
        self._repository = NewsRepository(session)

    async def list_news(
        self,
        category_id: uuid.UUID,
        *,
        cursor: str | None,
        source_tag_id: uuid.UUID | None,
    ) -> NewsPage | None:
        if not await self._repository.category_exists(category_id):
            return None
        before = self._decode_cursor(cursor) if cursor else None
        rows = await self._repository.list_visible(
            category_id=category_id,
            now=datetime.now(timezone.utc),
            limit=PAGE_SIZE + 1,
            source_tag_id=source_tag_id,
            before=before,
        )
        has_next_page = len(rows) > PAGE_SIZE
        items = rows[:PAGE_SIZE]
        next_cursor = self._encode_cursor(items[-1][1]) if has_next_page else None
        return NewsPage(items=[self._list_item(*row) for row in items], next_cursor=next_cursor)

    async def get_news_detail(
        self,
        category_id: uuid.UUID,
        article_id: uuid.UUID,
    ) -> NewsDetail | None:
        row = await self._repository.get_visible(
            category_id=category_id,
            article_id=article_id,
            now=datetime.now(timezone.utc),
        )
        if row is None:
            return None
        article, category_article, source_setting = row
        return NewsDetail(
            **self._list_item(article, category_article, source_setting).model_dump(),
            first_seen_at=article.first_seen_at,
        )

    async def set_permanent(self, article_id: uuid.UUID, permanent: bool) -> Article | None:
        async with self._session.begin():
            article = await self._repository.get_active_article(article_id)
            if article is None:
                return None
            article.expires_at = None if permanent else article.first_seen_at + timedelta(days=30)
        return article

    async def delete_article(self, article_id: uuid.UUID) -> bool:
        async with self._session.begin():
            article = await self._repository.get_active_article(article_id)
            if article is None:
                return False
            article.deleted_at = datetime.now(timezone.utc)
        return True

    @staticmethod
    def _list_item(
        article: Article,
        category_article: CategoryArticle,
        source_setting: SourceSetting,
    ) -> NewsListItem:
        return NewsListItem(
            id=article.id,
            title=article.title,
            summary=article.summary,
            canonical_url=article.canonical_url,
            published_at=article.published_at,
            inserted_at=category_article.inserted_at,
            expires_at=article.expires_at,
            source_tag_id=source_setting.id,
            source_tag_label=source_setting.label,
        )

    @staticmethod
    def _encode_cursor(category_article: CategoryArticle) -> str:
        payload = json.dumps(
            {
                "inserted_at": category_article.inserted_at.isoformat(),
                "article_id": str(category_article.article_id),
            },
            separators=(",", ":"),
        ).encode()
        return base64.urlsafe_b64encode(payload).decode().rstrip("=")

    @staticmethod
    def _decode_cursor(cursor: str) -> tuple[datetime, uuid.UUID]:
        try:
            padded = cursor + "=" * (-len(cursor) % 4)
            payload = json.loads(base64.urlsafe_b64decode(padded).decode())
            inserted_at = datetime.fromisoformat(payload["inserted_at"])
            article_id = uuid.UUID(payload["article_id"])
            if inserted_at.tzinfo is None:
                raise ValueError("Cursor timestamp must include a timezone.")
        except (KeyError, TypeError, UnicodeDecodeError, ValueError, json.JSONDecodeError) as error:
            raise InvalidCursorError("Invalid cursor") from error
        return inserted_at, article_id
