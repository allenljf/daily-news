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

## Fix round 1 — reviewer findings

### Scope and changes

- Strengthened the B4 delete route integration test to query the same real
  PostgreSQL database after deletion and assert `deleted_at IS NOT NULL` for
  the Category, all three Source Settings, and the Category Article link while
  the Article count remains exactly one.
- Added direct route assertions that a deleted Category is absent from
  `GET /v1/categories` and that `PATCH /v1/categories/{id}` returns 404 with
  `application/problem+json`.
- Added `backend/tests/postgres.py` as the one reusable disposable
  PostgreSQL/Docker/Alembic helper. Both `tests/categories/test_routes.py` and
  `tests/integration/test_schema.py` use its function-scoped container fixture,
  migration runner, and `psql` helper; `tests/conftest.py` re-exports the
  fixture for pytest discovery. The schema test is kept in its intended file.

### Red and Green evidence

- Red for the extracted helper: after B4 route tests were changed to import
  `tests.postgres` but before that shared helper existed,
  `cd backend && uv run pytest tests/categories/test_routes.py tests/integration/test_schema.py -q`
  stopped during collection with `ModuleNotFoundError: No module named
  'tests.postgres'`.
- Green after adding the helper and test assertions: the same command reports
  **6 passed**. The new delete assertion returns `t,t,t,1`, proving all three
  Category-owned rows are soft-deleted while the Article is retained.

### Fix-round verification

- `cd backend && uv run pytest tests/categories/test_routes.py tests/integration/test_schema.py -q`
  — **6 passed**.
- `cd backend && uv run ruff check .` — **All checks passed!**.
- `cd backend && uv run pytest -q` — **12 passed**.
- `git diff --check` — no whitespace errors.

The same pre-existing Starlette/httpx `TestClient` deprecation warning appears;
there are no test or lint failures. `docs/tasks/daily-news.md` and the
untracked `android-dev-guide/` directory remain untouched.
