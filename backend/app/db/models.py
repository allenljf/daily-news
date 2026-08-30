from __future__ import annotations

import uuid
from datetime import date, datetime

from sqlalchemy import Date, DateTime, ForeignKey, Index, Integer, String, Text, func, text
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import DeclarativeBase, Mapped, mapped_column


class Base(DeclarativeBase):
    """Base metadata for the Daily News schema."""


class Category(Base):
    __tablename__ = "categories"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name: Mapped[str] = mapped_column(String(length=255))
    search_keywords: Mapped[str | None] = mapped_column(Text(), nullable=True)
    special_requirements: Mapped[str | None] = mapped_column(Text(), nullable=True)
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
    )
    updated_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
    )
    deleted_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)


class SourceSetting(Base):
    __tablename__ = "source_settings"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    category_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("categories.id"),
    )
    label: Mapped[str] = mapped_column(String(length=255))
    website_input: Mapped[str] = mapped_column(Text())
    normalized_host: Mapped[str | None] = mapped_column(String(length=255), nullable=True)
    kind: Mapped[str] = mapped_column(String(length=50))
    position: Mapped[int] = mapped_column(Integer())
    created_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
    )
    deleted_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)


class Article(Base):
    __tablename__ = "articles"
    __table_args__ = (
        Index("ix_articles_canonical_url_hash", "canonical_url_hash", unique=True),
        Index("ix_articles_normalized_title_hash", "normalized_title_hash"),
    )

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    title: Mapped[str] = mapped_column(Text())
    normalized_title_hash: Mapped[str] = mapped_column(String(length=64))
    canonical_url: Mapped[str] = mapped_column(Text())
    canonical_url_hash: Mapped[str] = mapped_column(String(length=64))
    summary: Mapped[str | None] = mapped_column(Text(), nullable=True)
    published_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    first_seen_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
    )
    expires_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    deleted_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)


class CategoryArticle(Base):
    __tablename__ = "category_articles"
    __table_args__ = (
        Index(
            "ix_category_articles_category_article",
            "category_id",
            "article_id",
            unique=True,
        ),
    )

    category_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("categories.id"),
        primary_key=True,
    )
    article_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("articles.id"),
        primary_key=True,
    )
    source_setting_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("source_settings.id"),
    )
    inserted_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
    )


class IngestionRun(Base):
    __tablename__ = "ingestion_runs"
    __table_args__ = (Index("ix_ingestion_runs_idempotency_key", "idempotency_key", unique=True),)

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    trigger: Mapped[str] = mapped_column(String(length=32))
    idempotency_key: Mapped[str] = mapped_column(String(length=255))
    taipei_date: Mapped[date] = mapped_column(Date())
    started_at: Mapped[datetime] = mapped_column(
        DateTime(timezone=True),
        server_default=func.now(),
    )
    finished_at: Mapped[datetime | None] = mapped_column(DateTime(timezone=True), nullable=True)
    status: Mapped[str] = mapped_column(String(length=32))
    candidate_count: Mapped[int] = mapped_column(Integer(), server_default="0")
    inserted_count: Mapped[int] = mapped_column(Integer(), server_default="0")
    duplicate_count: Mapped[int] = mapped_column(Integer(), server_default="0")
    error_count: Mapped[int] = mapped_column(Integer(), server_default="0")


class IngestionAttempt(Base):
    __tablename__ = "ingestion_attempts"

    id: Mapped[uuid.UUID] = mapped_column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    run_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("ingestion_runs.id"),
    )
    category_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("categories.id"),
    )
    source_setting_id: Mapped[uuid.UUID] = mapped_column(
        UUID(as_uuid=True),
        ForeignKey("source_settings.id"),
    )
    status: Mapped[str] = mapped_column(String(length=32))
    candidate_count: Mapped[int] = mapped_column(Integer(), server_default="0")
    inserted_count: Mapped[int] = mapped_column(Integer(), server_default="0")
    duplicate_count: Mapped[int] = mapped_column(Integer(), server_default="0")
    error_summary: Mapped[str | None] = mapped_column(Text(), nullable=True)


Index(
    "ix_category_articles_feed_cursor",
    CategoryArticle.category_id,
    text("inserted_at DESC"),
    text("article_id DESC"),
)
