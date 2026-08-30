"""Authenticated HTTP routes for Category-scoped news and global Article mutations."""

from __future__ import annotations

import uuid
from typing import Annotated

from fastapi import APIRouter, Depends, Query, Response, status
from sqlalchemy.ext.asyncio import AsyncSession

from app.core.errors import bad_request_error, not_found_error
from app.db.engine import get_session
from app.identity.dependencies import require_allowed_identity
from app.identity.firebase import VerifiedIdentity
from app.news.schemas import ArticleResponse, NewsDetail, NewsPage, UpdateArticleRequest
from app.news.service import PAGE_SIZE, InvalidCursorError, NewsService

router = APIRouter(prefix="/v1", tags=["news"])


@router.get("/categories/{category_id}/news", response_model=NewsPage)
async def list_news(
    category_id: uuid.UUID,
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
    cursor: str | None = None,
    limit: int = Query(default=PAGE_SIZE, ge=1, le=PAGE_SIZE),
    source_tag_id: uuid.UUID | None = Query(default=None, alias="sourceTagId"),
) -> NewsPage:
    del limit
    try:
        page = await NewsService(session).list_news(
            category_id,
            cursor=cursor,
            source_tag_id=source_tag_id,
        )
    except InvalidCursorError as error:
        raise bad_request_error("Invalid cursor") from error
    if page is None:
        raise not_found_error("Category not found")
    return page


@router.get("/categories/{category_id}/news/{article_id}", response_model=NewsDetail)
async def get_news_detail(
    category_id: uuid.UUID,
    article_id: uuid.UUID,
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> NewsDetail:
    detail = await NewsService(session).get_news_detail(category_id, article_id)
    if detail is None:
        raise not_found_error("Article not found")
    return detail


@router.patch("/news/{article_id}", response_model=ArticleResponse)
async def update_news(
    article_id: uuid.UUID,
    request: UpdateArticleRequest,
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> ArticleResponse:
    article = await NewsService(session).set_permanent(article_id, request.permanent)
    if article is None:
        raise not_found_error("Article not found")
    return ArticleResponse(
        id=article.id,
        title=article.title,
        summary=article.summary,
        canonical_url=article.canonical_url,
        published_at=article.published_at,
        expires_at=article.expires_at,
        first_seen_at=article.first_seen_at,
    )


@router.delete("/news/{article_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_news(
    article_id: uuid.UUID,
    _: Annotated[VerifiedIdentity, Depends(require_allowed_identity)],
    session: Annotated[AsyncSession, Depends(get_session)],
) -> Response:
    if not await NewsService(session).delete_article(article_id):
        raise not_found_error("Article not found")
    return Response(status_code=status.HTTP_204_NO_CONTENT)
