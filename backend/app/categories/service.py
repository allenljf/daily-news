"""Category settings application service."""

from __future__ import annotations

import uuid
from datetime import datetime, timezone

from sqlalchemy.ext.asyncio import AsyncSession

from app.categories.repository import CategoryRepository
from app.categories.schemas import (
    CategoryResponse,
    CreateCategoryRequest,
    SourceSettingResponse,
    UpdateCategoryRequest,
)
from app.db.models import Category, SourceSetting


class CategoryService:
    """Coordinates transactional Category and Source Setting changes."""

    def __init__(self, session: AsyncSession) -> None:
        self._session = session
        self._repository = CategoryRepository(session)

    async def list_categories(self) -> list[CategoryResponse]:
        categories = await self._repository.list_active()
        return [
            self._response(category, await self._repository.active_source_settings(category.id))
            for category in categories
        ]

    async def create_category(self, request: CreateCategoryRequest) -> CategoryResponse:
        async with self._session.begin():
            category = self._repository.add_category(
                name=request.name,
                search_keywords=request.search_keywords,
                special_requirements=request.special_requirements,
            )
            await self._session.flush()
            source_settings = self._repository.add_source_settings(
                category.id,
                request.source_settings,
            )
            await self._session.flush()
            return self._response(category, source_settings)

    async def update_category(
        self,
        category_id: uuid.UUID,
        request: UpdateCategoryRequest,
    ) -> CategoryResponse | None:
        async with self._session.begin():
            category = await self._repository.get_active(category_id)
            if category is None:
                return None
            category.name = request.name
            category.search_keywords = request.search_keywords
            category.special_requirements = request.special_requirements
            category.updated_at = datetime.now(timezone.utc)
            now = datetime.now(timezone.utc)
            await self._repository.soft_delete_source_settings(category_id, now)
            source_settings = self._repository.add_source_settings(
                category_id,
                request.source_settings,
            )
            await self._session.flush()
            return self._response(category, source_settings)

    async def delete_category(self, category_id: uuid.UUID) -> bool:
        async with self._session.begin():
            category = await self._repository.get_active(category_id)
            if category is None:
                return False
            deleted_at = datetime.now(timezone.utc)
            category.deleted_at = deleted_at
            await self._repository.soft_delete_source_settings(category_id, deleted_at)
            await self._repository.soft_delete_category_articles(category_id, deleted_at)
        return True

    @staticmethod
    def _response(category: Category, source_settings: list[SourceSetting]) -> CategoryResponse:
        return CategoryResponse(
            id=category.id,
            name=category.name,
            search_keywords=category.search_keywords,
            special_requirements=category.special_requirements,
            source_settings=[
                SourceSettingResponse.model_validate(setting) for setting in source_settings
            ],
        )
