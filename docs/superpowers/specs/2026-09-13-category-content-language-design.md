# Category Content Language Design

**Status:** approved for implementation
**Date:** 2026-09-13
**Scope:** add a Category-level content-language setting, initially limited to Traditional Chinese.

## Goal

Every Category declares the language of Articles it accepts. The only supported
value in this release is `zh-Hant` (Traditional Chinese), and it is the default
for newly created and pre-existing Categories.

## Data and contract

`categories` gains a non-null `content_language text` column with the database
default `zh-Hant`. The migration backfills existing rows through that default.
The column has a check constraint permitting only `zh-Hant`, so malformed or
unsupported language choices cannot enter the ingestion pipeline.

The checked-in OpenAPI artifact adds required `content_language` to
`CategoryResponse`, `CreateCategoryRequest`, and the corresponding update
request. It is a backward-compatible response addition. Requests that omit it
are accepted and interpreted as `zh-Hant`; requests that send another value
receive the existing 422 validation response. This preserves older Flutter
clients during rollout while all newly built clients explicitly send the value.

## Backend and ingestion flow

The Go Category `Request` and `Response` models carry `ContentLanguage`. Store
create and update normalise an omitted value to `zh-Hant`, validate the sole
supported value, and persist it atomically with the Category and Source
Settings. List reads it from PostgreSQL.

`WorkPlanner` reads `categories.content_language` and includes it in
`SourceWork`. Concrete search adapters use it in their search instructions.
Candidate acceptance is defensive: an adapter must reject a result when the
available language metadata says it is not Traditional Chinese. Missing language
metadata remains eligible only when the adapter's text-level Traditional Chinese
classifier accepts its title and available summary. The system never converts
Simplified Chinese source text to Traditional Chinese.

## Flutter flow

The Category data model, DTO, repository mapping and `CategoryDraft` carry the
same BCP-47 string. Category Settings displays a required language selector
labelled「內容語言」with only「繁體中文」available and preselected. Saving posts
`content_language: "zh-Hant"`. Existing API responses missing the field decode
to `zh-Hant` to make rolling deployment safe.

## Error handling and observability

An unsupported request value is a validation error (422), not a silent fallback.
Language-filtered candidates are counted as candidates but not written; their
per-source attempt error summary stays empty because filtering is expected,
successful behavior. Logs must not include complete prompts or article content.

## Verification

- PostgreSQL migration tests prove the default, backfill and constraint.
- Go handler/store tests prove omission defaults to `zh-Hant`, responses expose
  it, and unsupported values produce 422.
- Planner and adapter tests prove the language reaches the search request and
  non-Traditional candidates are rejected.
- Flutter DTO and widget tests prove absent legacy response defaults correctly,
  the selector displays the default, and saving sends `zh-Hant`.
- Run `gofmt`, `go vet ./...`, `go test ./...`, `flutter analyze`, focused and
  full Flutter tests, contract validation, and `git diff --check`.
