# Go backend architecture for Daily News

**Research date:** 2026-09-01  
**Scope:** Replace the current FastAPI/SQLAlchemy backend while preserving the
existing HTTP contract, PostgreSQL schema, Cloud Run topology, and ingestion
behaviour. Sources below are Go, Google Cloud, Firebase, or the selected
library maintainers' official documentation.

## Recommendation

Use a deliberately small, standard-library-first Go service:

```text
cmd/api                 composition root; HTTP server and graceful shutdown
cmd/daily-news-job      composition root; RUN_ID validation and job exit code
internal/httpapi        routes, auth middleware, request/response DTOs
internal/category       service + repository interfaces/implementations
internal/news
internal/ingestion      orchestration, adapters, dedupe, run state
internal/identity       Firebase verifier and allowlist policy
internal/platform       config, database pool, logging, HTTP clients
migrations              versioned SQL migrations
```

`net/http` is sufficient for this API: make one explicit `http.ServeMux`, keep
handlers thin, and inject services into a small `Server` type. Keep business
logic and SQL outside handlers so the API and Cloud Run Job share the same
services. This is a project selection, not a claim that Go mandates this exact
layout; Go's official module guidance describes layouts as context-dependent.
[Organizing a Go module](https://go.dev/doc/modules/layout).

Use `database/sql` plus pgx's `database/sql` adapter for a conventional,
portable repository boundary. This lets the code demonstrate the Go standard
library while retaining PostgreSQL support. Choose pgx's native API only if the
implementation genuinely needs PostgreSQL-specific capability or performance;
pgx officially provides both interfaces.
[pgx](https://github.com/jackc/pgx), [pgx stdlib adapter](https://pkg.go.dev/github.com/jackc/pgx/v5/stdlib).

Selected non-standard libraries (not official Go recommendations):

- `firebase.google.com/go/v4` — Firebase Admin Go SDK.
- `github.com/jackc/pgx/v5` — PostgreSQL driver and `database/sql` adapter.
- `github.com/golang-migrate/migrate/v4` — versioned SQL migrations.

Avoid adding a web framework, ORM, dependency-injection container, or generic
repository framework unless a later requirement proves it useful.

## HTTP, context, JSON, and errors

- Register routes on an explicit `http.NewServeMux`; use `http.Server` for
  timeouts and graceful shutdown. A handler must finish using the request body
  and response writer before returning. [net/http](https://pkg.go.dev/net/http)
- Receive `ctx := r.Context()` in each handler and pass it as the first argument
  through application, repository, Firebase, and outbound HTTP calls. Incoming
  request contexts are cancelled when the client disconnects or the handler
  returns; derive shorter operation limits with `context.WithTimeout` and call
  its `cancel`. Do not store a request context in a long-lived struct.
  [context](https://pkg.go.dev/context), [Contexts and structs](https://go.dev/blog/context-and-structs)
- Decode a single request payload with `json.Decoder`, calling
  `DisallowUnknownFields` for strict request DTOs; encode responses with
  `json.Encoder`. Explicitly map validation/domain errors to the existing
  `application/problem+json` contract (401/403/404/409/422/400); never return
  raw internal errors. [encoding/json](https://pkg.go.dev/encoding/json)
- Wrap lower-level errors with `%w` for diagnostics, then use `errors.Is` /
  `errors.As` in the HTTP error mapper to classify sentinel or typed errors.
  [errors](https://pkg.go.dev/errors)

## PostgreSQL and migrations

- Open one long-lived `*sql.DB`: it manages a connection pool rather than being
  one connection. Set pool limits based on Cloud Run concurrency and Cloud SQL
  capacity, then call `PingContext` during startup/readiness only when it fits
  the service policy. [Go database access overview](https://go.dev/doc/database/)
- Use `QueryContext`, `QueryRowContext`, and `ExecContext`; bind values as SQL
  arguments, never interpolate them. `Context` lets client cancellation and
  operation deadlines release database work. [Cancel database operations](https://go.dev/doc/database/cancel-operations), [Avoid SQL injection](https://go.dev/doc/database/sql-injection)
- For atomic operations such as active-run acquisition or inserting an Article
  plus Category Article plus Attempt, call `BeginTx(ctx, ...)`, do all queries
  through that `*sql.Tx`, defer `Rollback`, and finish with `Commit`. Do not mix
  `sql.Tx` operations with literal `BEGIN`/`COMMIT` statements.
  [Execute transactions](https://go.dev/doc/database/execute-transactions)
- Parameterization is the default protection. Prepare only demonstrably hot,
  repeated statements; close statements and use `Tx.PrepareContext` or
  `Tx.StmtContext` when a statement is used in a transaction.
  [Prepared statements](https://go.dev/doc/database/prepared-statements)
- Preserve the current schema and migration history as the contract. For new Go
  migrations, select `golang-migrate` and checked-in `{version}_{name}.up.sql` /
  `.down.sql` files. Its PostgreSQL driver documents transactional behaviour and
  calls out that operations such as `CREATE INDEX CONCURRENTLY` must be isolated.
  [Migration file convention](https://github.com/golang-migrate/migrate/blob/master/MIGRATIONS.md), [PostgreSQL driver](https://github.com/golang-migrate/migrate/blob/master/database/postgres/README.md)

## Firebase authentication

Initialize the Firebase Admin SDK with Application Default Credentials on Cloud
Run—Firebase strongly recommends ADC in Google environments, rather than
shipping a service-account key. For each Bearer token call
`app.Auth(ctx).VerifyIDToken(ctx, token)` and then apply the existing exact
email allowlist. `VerifyIDToken` validates format, expiry, and signature; use
`VerifyIDTokenAndCheckRevoked` only if the product adds revocation enforcement.
Tokens must arrive over HTTPS.
[Firebase Admin setup](https://firebase.google.com/docs/admin/setup), [Verify ID tokens](https://firebase.google.com/docs/auth/admin/verify-id-tokens).

## Tests

Use the standard `testing` package and `go test ./...`; keep unit tests beside
the package they exercise, use `t.Cleanup`/`t.TempDir` for cleanup, and test
handlers with `net/http/httptest`. Keep a PostgreSQL integration suite that
applies migrations to an empty disposable database, retaining every current
schema/unique-index assertion and API behaviour test.
[testing](https://pkg.go.dev/testing), [net/http/httptest](https://pkg.go.dev/net/http/httptest), [Go test tutorial](https://go.dev/doc/tutorial/add-a-test).

## Cloud Run implementation constraints

Build one Go image with two commands: an API binary and a Job binary. The
service binary must listen on `0.0.0.0:$PORT` (Cloud Run supplies `PORT`, default
8080); Cloud Run terminates TLS. The Job must not start an HTTP server: it runs
to completion, exits `0` on success and non-zero on failure. Preserve `RUN_ID`
as its explicit application input, honour `SIGTERM`/context cancellation, and
keep the Job single-task unless the workload is deliberately partitioned.
[Cloud Run container runtime contract](https://cloud.google.com/run/docs/container-contract), [Create Cloud Run jobs](https://cloud.google.com/run/docs/create-jobs).

Cloud Run Services are for the stateless HTTP API, whereas Jobs are for manual
or scheduled work that completes. This matches the existing topology without
changing the GitHub OIDC/WIF deployment model or Secret Manager ownership.
[Cloud Run overview](https://cloud.google.com/run/docs/overview/what-is-cloud-run).

## Migration guardrails

1. Treat the existing OpenAPI response/error shapes and PostgreSQL schema as
   compatibility contracts; generate or maintain an equivalent OpenAPI document
   before the Flutter client is changed.
2. Port one vertical slice at a time with parity tests: identity, categories,
   news, ingestion-run launching, then ingestion adapters/orchestrator.
3. Keep the existing Cloud Run manifests, GitHub workflows, secret names, and
   Firebase/Cloud SQL resources; only replace image build paths and commands.
4. Do not delete `backend/` until Go migration/schema/API/Job parity tests pass
   and staging smoke tests are complete.
