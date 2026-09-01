# Go backend task entrypoint

Before backend work, read `../CONTEXT.md`, `../docs/requirements/daily-news.md`,
`../docs/tasks/daily-news.md`,
`../docs/contracts/daily-news.openapi.json`, and
`../docs/superpowers/specs/2026-09-01-go-backend-migration-design.md`.

For Go replacement tasks, implement only the uncompleted task whose dependencies
are complete. Preserve the `/v1` OpenAPI artifact, PostgreSQL schema semantics,
Firebase email allowlist, and Cloud Run/GitHub WIF/Secret Manager names. Update
the artifact only through its explicit contract-baseline task, with a Flutter
compatibility review.

Use a context-first call chain: obtain `ctx` from `r.Context()`, pass it as the
first argument to application, repository, Firebase, database, and outbound HTTP
operations, and derive bounded operation contexts with a cancelled timeout.
Do not retain a request context in a long-lived struct.

Map validation and domain errors explicitly to the checked-in
`application/problem+json` status and body contract. Wrap internal errors for
diagnostics, but never expose them in a response.

Set transaction boundaries at application operations that must be atomic. Use
`BeginTx`, run every query for that operation through the resulting `*sql.Tx`,
defer rollback, and commit only after all invariants hold. In particular, active
Ingestion Run acquisition and Article/Category Article/Attempt writes must not
split their atomic boundary.

Before recording a Go task as complete, run `gofmt` on changed Go files,
`go vet ./...`, and `go test ./...`, plus the task-specific verification. Record
the exact command and result in `docs/tasks/daily-news.md`. Commit only after
verification passes.
