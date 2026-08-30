# B4 — Category Settings API report

## Scope

Implemented only B4: authenticated `GET`, `POST`, `PATCH`, and `DELETE`
`/v1/categories` with Router → Service → Repository separation.  Creation and
update receive complete ordered source-setting payloads; blank source-setting
placeholders are omitted; update soft-retires active settings before replacing
them.  Category deletion soft-deletes the Category, its Source Settings, and
its Category Article links, never its Articles.  A small follow-on migration
adds the `category_articles.deleted_at` column required to fulfill that
soft-delete contract; B2's initial migration is unchanged.

## Changed files

- `backend/app/categories/{__init__.py,schemas.py,repository.py,service.py,router.py}`
- `backend/tests/categories/test_routes.py`
- `backend/app/main.py`
- `backend/app/core/errors.py`
- `backend/app/db/models.py`
- `backend/alembic/versions/20260830_0002_category_article_soft_delete.py`

`docs/tasks/daily-news.md` and the untracked `android-dev-guide/` directory
were not modified.

## TDD evidence

- Initial test harness attempt: `cd backend && uv run pytest tests/categories/test_routes.py -q`
  stopped at test setup because `postgres_container` was local to the existing
  schema test rather than a shared fixture. The B4 test now owns its equivalent
  disposable PostgreSQL fixture.
- Red (before any B4 production modules/routes): the same command completed
  with **5 failures**. The create, validation, and unauthenticated contracts
  received `404 Not Found`, demonstrating that `/v1/categories` was absent;
  two later tests had expected cascading missing-response-key failures.
- Green: after the B4 implementation, the focused suite reports **5 passed**.

## Verification commands and results

- `cd backend && uv run pytest tests/categories/test_routes.py -q` — **5 passed**.
- `cd backend && uv run ruff check . --fix` — fixed one import-order issue;
  rerunning `cd backend && uv run ruff check .` reported **All checks passed!**.
- `cd backend && uv run pytest -q` — **12 passed**.
- Final fresh combined verification:
  `cd backend && uv run pytest tests/categories/test_routes.py -q && uv run ruff check . && uv run pytest -q`
  — **5 passed**, **All checks passed!**, **12 passed**.
- `git diff --check` — no whitespace errors.

The test commands emit the pre-existing Starlette/httpx `TestClient`
deprecation warning; no test or lint failure remains.

## PostgreSQL proof

Every B4 route test starts a fresh disposable `postgres:16-alpine` container,
runs `uv run alembic upgrade head` against it, and injects the application's
real async SQLAlchemy session (`postgresql+asyncpg`). The delete contract also
uses `psql` to insert an Article plus Category Article link, then queries the
real database after `DELETE /v1/categories/{id}` and asserts exactly one
matching Article remains. The passing 5-test focused suite therefore verifies
the migration, real PostgreSQL writes, order replacement, and no-Article-delete
behavior without SQLite or Firebase.

## Self-review

- All four category routes carry `require_allowed_identity`; the no-token route
  test confirms 401.
- Request/response schemas are DTOs, not ORM response models. Blank names are
  validation errors (422); response source order is persisted in `position`.
- PATCH uses one transaction, soft-deactivates all active settings, writes only
  the supplied ordered nonblank settings, and updates `updated_at`.
- DELETE is transactional, marks all three Category-owned record types deleted,
  and has no Article delete statement. A second delete returns a 404
  `application/problem+json` response.
- No credentials, tokens, or user data are logged.

## Concerns

No B4 blockers. The only runtime noise is the existing Starlette/httpx
deprecation warning noted above. Future news-list work should explicitly filter
`CategoryArticle.deleted_at IS NULL` and preserve historical source labels when
displaying previously ingested articles.
