"""HTTP DTOs for category settings."""

from __future__ import annotations

import uuid

from pydantic import BaseModel, ConfigDict, field_validator


class SourceSettingInput(BaseModel):
    """A user-supplied source setting before persistence."""

    label: str = ""
    website_input: str = ""
    kind: str = "unspecified"

    @field_validator("label", "website_input", "kind", mode="before")
    @classmethod
    def strip_text(cls, value: str | None) -> str:
        return value.strip() if isinstance(value, str) else ""

    def is_empty(self) -> bool:
        """Whether this UI placeholder should be omitted from persistence."""
        return not self.label and not self.website_input


class CreateCategoryRequest(BaseModel):
    """The complete category settings payload used for creation."""

    name: str
    search_keywords: str | None = None
    special_requirements: str | None = None
    source_settings: list[SourceSettingInput]

    @field_validator("name")
    @classmethod
    def require_name(cls, value: str) -> str:
        value = value.strip()
        if not value:
            raise ValueError("Category name must not be blank.")
        return value


class UpdateCategoryRequest(CreateCategoryRequest):
    """The complete replacement payload used for updates."""


class SourceSettingResponse(BaseModel):
    """A persisted source setting returned by the API."""

    model_config = ConfigDict(from_attributes=True)

    id: uuid.UUID
    label: str
    website_input: str
    normalized_host: str | None
    kind: str
    position: int


class CategoryResponse(BaseModel):
    """A category settings response, independent from ORM entities."""

    id: uuid.UUID
    name: str
    search_keywords: str | None
    special_requirements: str | None
    source_settings: list[SourceSettingResponse]
