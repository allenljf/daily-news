"""Authenticated HTTP routes for Category settings."""

from __future__ import annotations

import uuid
from typing import Annotated

from fastapi import APIRouter, Depends, Response, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.categories.schemas import CategoryResponse, CreateCategoryRequest, UpdateCategoryRequest
from app.categories.service import CategoryService
from app.core.errors import not_found_error
from app.db.engine import get_session
from app.identity.dependencies import require_allowed_identity
from app.identity.firebase import VerifiedIdentity

router = APIRouter(prefix="/v1/categories", tags=["categories"])


@router.get("", response_model=list[CategoryResponse])
async def list_categories(
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> list[CategoryResponse]:
    return await CategoryService(session).list_categories()


@router.post("", response_model=CategoryResponse, status_code=status.HTTP_201_CREATED)
async def create_category(
    request: CreateCategoryRequest,
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> CategoryResponse:
    return await CategoryService(session).create_category(request)


@router.patch("/{category_id}", response_model=CategoryResponse)
async def update_category(
    category_id: uuid.UUID,
    request: UpdateCategoryRequest,
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> CategoryResponse:
    category = await CategoryService(session).update_category(category_id, request)
    if category is None:
        raise not_found_error("Category not found")
    return category


@router.delete("/{category_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_category(
    category_id: uuid.UUID,
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> Response:
    if not await CategoryService(session).delete_category(category_id):
        raise not_found_error("Category not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)
