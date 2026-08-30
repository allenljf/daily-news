"""Database operations for Category settings."""

from __future__ import annotations

import uuid
from datetime import datetime
from urllib.parse import urlparse

from sqlalchemy import select, update
from sqlalchemy.ext.asyncio import AsyncSession

from app.categories.schemas import SourceSettingInput
from app.db.models import Category, CategoryArticle, SourceSetting


class CategoryRepository:
    """Persistence adapter for active categories and their source settings."""

    def __init__(self, session: AsyncSession) -> None:
        self._session = session

    async def list_active(self) -> list[Category]:
        result = await self._session.scalars(
            select(Category).where(Category.deleted_at.is_(None)).order_by(Category.created_at)
        )
        return list(result)

    async def get_active(self, category_id: uuid.UUID) -> Category | None:
        return await self._session.scalar(
            select(Category).where(Category.id == category_id, Category.deleted_at.is_(None))
        )

    async def active_source_settings(self, category_id: uuid.UUID) -> list[SourceSetting]:
        result = await self._session.scalars(
            select(SourceSetting)
            .where(
                SourceSetting.category_id == category_id,
                SourceSetting.deleted_at.is_(None),
            )
            .order_by(SourceSetting.position)
        )
        return list(result)

    def add_category(
        self,
        *,
        name: str,
        search_keywords: str | None,
        special_requirements: str | None,
    ) -> Category:
        category = Category(
            name=name,
            search_keywords=search_keywords,
            special_requirements=special_requirements,
        )
        self._session.add(category)
        return category

    def add_source_settings(
        self,
        category_id: uuid.UUID,
        source_settings: list[SourceSettingInput],
    ) -> list[SourceSetting]:
        persisted: list[SourceSetting] = []
        for position, source_setting in enumerate(
            setting for setting in source_settings if not setting.is_empty()
        ):
            setting = SourceSetting(
                category_id=category_id,
                label=source_setting.label,
                website_input=source_setting.website_input,
                normalized_host=self._normalized_host(source_setting.website_input),
                kind=source_setting.kind,
                position=position,
            )
            self._session.add(setting)
            persisted.append(setting)
        return persisted

    async def soft_delete_source_settings(
        self,
        category_id: uuid.UUID,
        deleted_at: datetime,
    ) -> None:
        await self._session.execute(
            update(SourceSetting)
            .where(SourceSetting.category_id == category_id, SourceSetting.deleted_at.is_(None))
            .values(deleted_at=deleted_at)
        )

    async def soft_delete_category_articles(
        self,
        category_id: uuid.UUID,
        deleted_at: datetime,
    ) -> None:
        await self._session.execute(
            update(CategoryArticle)
            .where(CategoryArticle.category_id == category_id, CategoryArticle.deleted_at.is_(None))
            .values(deleted_at=deleted_at)
        )

    @staticmethod
    def _normalized_host(website_input: str) -> str | None:
        if not website_input:
            return None
        parsed = urlparse(website_input if "://" in website_input else f"//{website_input}")
        return parsed.hostname
