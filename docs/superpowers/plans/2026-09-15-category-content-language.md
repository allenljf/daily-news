# Category Content Language Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Let every Category select a content language, defaulting and initially restricting it to Traditional Chinese (`zh-Hant`).

**Architecture:** Persist `content_language` on Category with a database default and constraint; expose it through the checked-in OpenAPI contract and Go Category boundary. Carry it in `ingestion.SourceWork` so adapters make and enforce the same language decision. Flutter treats it as immutable feature data and posts its required default from the Category settings form.

**Tech Stack:** PostgreSQL/golang-migrate, Go `net/http`/`database/sql`, Flutter/Riverpod/Dio, OpenAPI JSON, Flutter widget tests.

**Spec:** `docs/superpowers/specs/2026-09-13-category-content-language-design.md`

## Global Constraints

- The only supported persisted content-language value is exactly `zh-Hant`.
- Omitted API request values and legacy response payloads default to `zh-Hant`; explicit unsupported values return `422 application/problem+json`.
- Do not convert Simplified Chinese source content to Traditional Chinese.
- Preserve existing Category, Source Setting, Article, Firebase and ingestion transaction semantics.
- Preserve unrelated uncommitted workspace changes.

---

### Task 1: Persist and expose Category content language

**Files:**

- Create: `backend/migrations/000004_category_content_language.up.sql`, `backend/migrations/000004_category_content_language.down.sql`
- Modify: `backend/internal/category/category.go`, `backend/internal/category/handler_test.go`, `backend/internal/category/store_integration_test.go`, `docs/contracts/daily-news.openapi.json`, `docs/requirements/daily-news.md`, `docs/tasks/daily-news.md`

**Interfaces:**

- Produces: `category.Request.ContentLanguage string`, `category.Response.ContentLanguage string`.
- Consumes: `POST`/`PATCH /v1/categories` JSON field `content_language`.

- [ ] **Step 1: Write the failing test** — prove an omitted request gives `zh-Hant`, explicit `en` gives 422, and the migrated table has a default/check constraint.
- [ ] **Step 2: Verify RED** — run `cd backend && go test ./internal/category -run 'ContentLanguage|StorePreserves' -count=1`; expect failure because the model, SQL and migration lack the field.
- [ ] **Step 3: Implement GREEN** — add migration 4 with `text NOT NULL DEFAULT 'zh-Hant' CHECK (content_language = 'zh-Hant')`; add required response and optional request fields in OpenAPI; normalise omitted values, validate explicit values, and select/insert/update the field in `Store`.
- [ ] **Step 4: Verify GREEN** — run `cd backend && gofmt -w internal/category && go test ./internal/category -count=1`; expect pass.
- [ ] **Step 5: Commit** — `git add backend/migrations/000004_category_content_language.* backend/internal/category docs/contracts/daily-news.openapi.json docs/requirements/daily-news.md docs/tasks/daily-news.md && git commit -m "feat: add category content language"`.

### Task 2: Carry language into ingestion and reject known non-Traditional candidates

**Files:**

- Modify: `backend/internal/ingestion/orchestrator.go`, `backend/internal/ingestion/planner.go`, `backend/internal/ingestion/planner_integration_test.go`, concrete web adapter and its tests.

**Interfaces:**

- Consumes: `categories.content_language`.
- Produces: `ingestion.SourceWork.ContentLanguage string`; adapters return only matching candidates.

- [ ] **Step 1: Write failing tests** — assert planner returns `zh-Hant`; explicit `zh-Hans` candidate metadata is rejected while `zh-Hant` is retained.
- [ ] **Step 2: Verify RED** — run `cd backend && go test ./internal/ingestion -run 'WorkPlanner|Traditional' -count=1`; expect failure because language is not selected or gated.
- [ ] **Step 3: Implement GREEN** — select/scan language into `SourceWork`, include it in search instructions, reject candidates with known non-`zh-Hant` metadata, and never transform source text.
- [ ] **Step 4: Verify GREEN** — run `cd backend && gofmt -w internal/ingestion && go test ./internal/ingestion -count=1`; expect pass.
- [ ] **Step 5: Commit** — `git add backend/internal/ingestion && git commit -m "feat: filter ingestion by category language"`.

### Task 3: Add the Flutter default selector and API mapping

**Files:**

- Modify: `apps/mobile/lib/features/categories/data/{category.dart,category_dto.dart,category_remote_service.dart,category_repository.dart}`, `apps/mobile/lib/features/categories/presentation/category_settings_sheet.dart`, `apps/mobile/lib/l10n/app_zh.arb`, generated localization files, `apps/mobile/test/features/categories/category_sheet_test.dart`.

**Interfaces:**

- Produces: `Category.contentLanguage`, `CategoryDraft.contentLanguage`, and POST body `content_language`.

- [ ] **Step 1: Write failing tests** — legacy DTO payload defaults to `zh-Hant`; form renders「內容語言」/「繁體中文」and passes `zh-Hant` through saved draft and POST body.
- [ ] **Step 2: Verify RED** — run `cd apps/mobile && flutter test test/features/categories/category_sheet_test.dart -r expanded`; expect failure because no language model or selector exists.
- [ ] **Step 3: Implement GREEN** — add immutable fields/defaults to data models and DTO mapping, include language in POST body, add localized copy and a one-option `DropdownButtonFormField<String>` defaulted to `zh-Hant`.
- [ ] **Step 4: Verify GREEN** — run `cd apps/mobile && flutter gen-l10n && dart format lib/features/categories test/features/categories && flutter analyze && flutter test test/features/categories/category_sheet_test.dart -r expanded`; expect pass.
- [ ] **Step 5: Commit** — `git add apps/mobile/lib/features/categories apps/mobile/lib/l10n apps/mobile/test/features/categories && git commit -m "feat: select category content language"`.

### Task 4: Full verification and completion record

**Files:**

- Modify: `docs/tasks/daily-news.md`

- [ ] **Step 1: Run backend verification** — `cd backend && go vet ./... && go test ./...`; expect pass.
- [ ] **Step 2: Run Flutter verification** — `cd apps/mobile && flutter analyze && flutter test && cd ../.. && python3 flutter-dev-guide/tools/check-rules.py --all`; expect pass.
- [ ] **Step 3: Validate contract and diff** — `python3 -m json.tool docs/contracts/daily-news.openapi.json >/dev/null && git diff --check`; expect pass.
- [ ] **Step 4: Record exact successful commands and check off the language task** in `docs/tasks/daily-news.md`; do not claim external staging verification.
- [ ] **Step 5: Commit** — `git add docs/tasks/daily-news.md && git commit -m "docs: record category language verification"`.
