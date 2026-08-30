"""Database operations for visible Category Articles and global Article mutations."""

from __future__ import annotations

import uuid
from datetime import datetime

from sqlalchemy import and_, or_, select
from sqlalchemy.ext.asyncio import AsyncSession

from app.db.models import Article, Category, CategoryArticle, SourceSetting


class NewsRepository:
    """Persistence adapter for Category-scoped Article reads."""

    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def category_exists(self, category_id: uuid.UUID) -> bool:
        return (
            await self._session.scalar(
                select(Category.id).where(Category.id == category_id, Category.deleted_at.is_(None))
            )
            is not None
        )

    async def list_visible(
        self,
        *,
        category_id: uuid.UUID,
        now: datetime,
        limit: int,
        source_tag_id: uuid.UUID | None,
        before: tuple[datetime, uuid.UUID] | None,
    ) -> list[tuple[Article, CategoryArticle, SourceSetting]]:
        conditions = [
            CategoryArticle.category_id == category_id,
            CategoryArticle.deleted_at.is_(None),
            Article.deleted_at.is_(None),
            or_(Article.expires_at.is_(None), Article.expires_at > now),
            SourceSetting.deleted_at.is_(None),
        ]
        if source_tag_id is not None:
            conditions.append(CategoryArticle.source_setting_id == source_tag_id)
        if before is not None:
            inserted_at, article_id = before
            conditions.append(
                or_(
                    CategoryArticle.inserted_at < inserted_at,
                    and_(
                        CategoryArticle.inserted_at == inserted_at,
                        CategoryArticle.article_id < article_id,
                    ),
                )
            )
        result = await self._session.execute(
            select(Article, CategoryArticle, SourceSetting)
            .join(CategoryArticle, CategoryArticle.article_id == Article.id)
            .join(SourceSetting, SourceSetting.id == CategoryArticle.source_setting_id)
            .where(*conditions)
            .order_by(CategoryArticle.inserted_at.desc(), CategoryArticle.article_id.desc())
            .limit(limit)
        )
        return list(result.tuples())

    async def get_visible(
        self,
        *,
        category_id: uuid.UUID,
        article_id: uuid.UUID,
        now: datetime,
    ) -> tuple[Article, CategoryArticle, SourceSetting] | None:
        result = await self._session.execute(
            select(Article, CategoryArticle, SourceSetting)
            .join(CategoryArticle, CategoryArticle.article_id == Article.id)
            .join(SourceSetting, SourceSetting.id == CategoryArticle.source_setting_id)
            .where(
                CategoryArticle.category_id == category_id,
                CategoryArticle.article_id == article_id,
                CategoryArticle.deleted_at.is_(None),
                Article.deleted_at.is_(None),
                or_(Article.expires_at.is_(None), Article.expires_at > now),
                SourceSetting.deleted_at.is_(None),
            )
        )
        return result.tuples().one_or_none()

    async def get_active_article(self, article_id: uuid.UUID) -> Article | None:
        return await self._session.scalar(
            select(Article).where(Article.id == article_id, Article.deleted_at.is_(None))
        )
