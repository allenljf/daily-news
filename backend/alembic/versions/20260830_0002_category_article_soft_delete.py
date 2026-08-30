"""Allow Category Article links to be soft-deleted with their Category."""

from typing import Sequence

import sqlalchemy as sa

from alembic import op

revision: str = "20260830_0002"
down_revision: str | None = "20260830_0001"
branch_labels: Sequence[str] | None = None
depends_on: Sequence[str] | None = None


def upgrade() -> None:
    op.add_column("category_articles", sa.Column("deleted_at", sa.DateTime(timezone=True)))


def downgrade() -> None:
    op.drop_column("category_articles", "deleted_at")
