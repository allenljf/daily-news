# Daily News Go Backend Migration Design

**Status:** proposed; implementation requires user confirmation
**Date:** 2026-09-01
**Scope:** replace Python/FastAPI with Go while retaining the checked-in HTTP contract, PostgreSQL semantics, Firebase allowlist, and Cloud Run service/Job topology.

## Compatibility boundary

This is an implementation-only migration. Flutter keeps the same `/v1` paths, JSON request/response shapes, status codes, cursor behaviour, and `application/problem+json` errors. [`docs/contracts/daily-news.openapi.json`](../../contracts/daily-news.openapi.json) is the authoritative machine-readable contract. PostgreSQL tables, data meanings, dedupe fingerprints, soft-deleted rows, and migration history stay in place. This repository work never creates or operates real GCP, GitHub, WIF, secret, or service-account resources.

The following remain true: Firebase verifies tokens and exact-matches `ALLOWED_USER_EMAIL`; canonical URL dedupe precedes normalized-title dedupe; deleted Articles remain dedupe-visible; feeds use the existing opaque keyset cursor and maximum 20 items; non-permanent Articles expire at 30 days; and one `queued` or `running` Ingestion Run exists globally at most once.

## Target module and structure

`backend/` becomes Go module `github.com/allenljf/daily-news/backend`. It uses only `net/http`, `context`, `encoding/json`, `errors`, `database/sql`, and `testing/httptest` from the standard library, plus `github.com/jackc/pgx/v5` through its `database/sql` adapter, `firebase.google.com/go/v4`, and `github.com/golang-migrate/migrate/v4`. No web framework, ORM, or DI container is used; explicit constructors and narrow I/O interfaces supply dependencies.

```text
cmd/api/                    HTTP composition root and graceful shutdown
cmd/daily-news-job/         Job composition root; RUN_ID and exit status
internal/httpapi/           ServeMux, middleware, DTOs, problem mapping
internal/identity/          Firebase verifier and exact-email policy
internal/category/          Category and Source Setting service and SQL
internal/news/              Article query, cursor, permanence, deletion
internal/ingestion/         runs, adapters, normalizer, dedupe, orchestrator
internal/platform/          config, sql.DB pool, Firebase, outbound HTTP
migrations/                 golang-migrate SQL representation
contract/                   OpenAPI and HTTP golden parity tests
```

One image contains both commands. The API binary listens on `0.0.0.0:$PORT`; the Job receives `RUN_ID`, performs one bounded Run, and exits zero only after a successful terminal Run. Cloud Run service/Job names, service accounts, Secret Manager names, GitHub OIDC/WIF, and `RUN_ID` remain unchanged.

## Request execution and error boundary

An explicit `http.NewServeMux` registers every route. Handlers obtain `ctx := r.Context()` and pass it as the first argument through application services, repositories, Firebase, SQL, and outbound HTTP. Bounded downstream work derives and cancels `context.WithTimeout`; request contexts never live in long-lived structs.

Handlers use a strict `json.Decoder` (unknown and trailing input rejected), then map typed validation/domain errors to the existing 400/401/403/404/409/422 Problem Details response. Unexpected wrapped errors are logged for diagnostics and never exposed. Every error response sets `application/problem+json` and preserves the contract artifact's body, field names, nullable values, and timestamps.

## SQL and transaction design

`platform.OpenDB` owns one process-lifetime pgx-backed `*sql.DB`; repositories use `QueryContext`, `QueryRowContext`, and `ExecContext` with bound parameters. Transactions are explicit application-operation boundaries: `AcquireManualRun` atomically performs active-run lookup, idempotency selection, and insert; `PersistCandidate` atomically writes Article, Category Article, Attempt, and Run counters. Each uses `BeginTx`, a deferred rollback, only `*sql.Tx` calls inside the boundary, then commit after every invariant holds.

Existing Alembic history is first applied to a disposable PostgreSQL database as a schema baseline. Go receives equivalent golang-migrate SQL without changing the resulting schema; a schema/index diff gate must pass before Go becomes deployable.

## Identity and ingestion

Firebase Admin Go SDK initializes with ADC. Its verifier receives the request context and is behind a small interface for offline tests. The allowlist is exact and server-side.

The ingestion core keeps `CandidateArticle`, source-search requests, and fingerprint logic free of SQL/HTTP. Adapters accept context and return candidates plus typed disabled/failure outcomes. The orchestrator caps candidates per Source Setting at ten, records each Attempt, retains per-source failure isolation, and preserves Article/Category Article dedupe semantics. The Job starts no HTTP server.

## Contract, parity, and cutover

R1 freezes FastAPI's current `/openapi.json` as a deterministic checked-in artifact and compares it with Flutter DTO use. The artifact is not regenerated by Go; any later change requires a deliberate Flutter-compatible contract review. Parity evidence is: OpenAPI structural check, `httptest` goldens for success and Problem Details responses, PostgreSQL schema/index checks, and product-invariant integration tests. Python remains until every parity gate passes.

The final cutover changes only Docker, CI, and Cloud Run command wiring. It leaves all resource names and identities unchanged. O4 staging smoke test starts only after Go parity. If staging fails, redeploy the prior image digest through the existing Cloud Run configuration; this plan makes no database rollback or external-resource mutation.

## Verification

Each Go task runs its focused test, `gofmt` on changed files, `go vet ./...`, `go test ./...`, the applicable PostgreSQL integration suite, and `git diff --check`. Record exact commands/results in the task completion record, then commit.
