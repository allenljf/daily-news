"""HTTP DTOs for Article list, detail, and lifecycle operations."""

from __future__ import annotations

import uuid
from datetime import datetime

from pydantic import BaseModel


class NewsListItem(BaseModel):
    """An Article as it appears in one Category's feed."""

    id: uuid.UUID
    title: str
    summary: str | None
    canonical_url: str
    published_at: datetime | None
    inserted_at: datetime
    expires_at: datetime | None
    source_tag_id: uuid.UUID
    source_tag_label: str


class NewsPage(BaseModel):
    """A stable, cursor-paginated Category Article feed page."""

    items: list[NewsListItem]
    next_cursor: str | None


class NewsDetail(NewsListItem):
    """The full available Article representation for a Category."""

    first_seen_at: datetime


class ArticleResponse(BaseModel):
    """An Article-global response that intentionally has no Category source tag."""

    id: uuid.UUID
    title: str
    summary: str | None
    canonical_url: str
    published_at: datetime | None
    first_seen_at: datetime
    expires_at: datetime | None


class UpdateArticleRequest(BaseModel):
    """Set an Article's global permanence state."""

    permanent: bool
