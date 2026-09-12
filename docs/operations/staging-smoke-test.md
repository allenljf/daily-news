# Daily News staging smoke test

This runbook verifies the Go API and ingestion Job against the real staging
resources. Never paste Firebase ID tokens, database URLs, secret values, or
URL query secrets into this file, terminal transcripts, issues, or logs.

## Failure criteria

The staging smoke test fails immediately if any of these conditions occurs:

- GitHub Actions cannot exchange its OIDC token through WIF or cannot deploy
  the service and Job with the documented deployer identity.
- `GET /healthz` does not reach the Go handler and return `200`,
  `application/json`, and `{"status":"ok"}` through the public Cloud Run URL.
- A missing Firebase token does not return `401 application/problem+json`, or
  an authenticated account other than `ALLOWED_USER_EMAIL` can read or mutate
  Daily News data instead of receiving `403 application/problem+json`.
- Two concurrent manual Run requests create different active Ingestion Runs,
  or more than one `queued`/`running` Run exists at the same time.
- A requested Run never reaches `succeeded` or `failed` within the bounded
  observation window, or its Job execution and Attempt records cannot be
  correlated by Run ID.
- A configured public Source Setting produces no inspectable Attempt, or a
  persisted Article lacks a public original URL and its required citation.
- Category creation, Article listing/detail, permanent-state mutation, global
  soft delete, or the mobile integration journey diverges from the checked-in
  OpenAPI contract.

## Evidence rules

- Record commit SHA, workflow run URL, Cloud Run revision, Run ID, timestamps,
  HTTP status/content type, aggregate counters, and pass/fail results.
- Redact bearer tokens, database connection strings, secret values, prompt
  content, and personal email addresses.
- Keep the allowlisted and non-allowlisted Firebase sessions separate. Do not
  save either token to the repository.
- Stop before data mutation while the public health preflight is failing.

## Execution order

1. Confirm CI and deploy workflows succeeded for the intended commit. Confirm
   the service and Job are Ready, the service uses the Go image, ingress is
   `all`, public invocation is enabled, and traffic is 100% on the intended
   revision.
2. Call public `/healthz`. Then call one protected endpoint without a token and
   verify the API's `401 application/problem+json` boundary.
3. With a non-allowlisted Firebase ID token, verify `403` and no data exposure.
4. With the allowlisted account, create a disposable Category with at least one
   public Source Setting. Record its Category and Source Setting IDs.
5. Send two manual Run requests concurrently. Verify both responses identify
   the same active Run and the database has at most one active Run.
6. Poll latest status until that Run reaches a terminal state. Correlate the
   Cloud Run Job execution, structured logs, `ingestion_runs`, and
   `ingestion_attempts`; record only identifiers and aggregate counters.
7. Verify the resulting Article list/detail contains a public original URL and
   citation, set the Article permanent, then globally soft-delete it and verify
   it no longer appears in any Category.
8. Run the mobile integration journey against the verified staging contract,
   remove the disposable Category if safe, and record final pass/fail evidence.

## 2026-09-12 preflight result

- Commit `5f464ca5f3ff5d9066e3c783253d34d1a67d1514` is on `origin/main`.
  CI run `34669133586` and deploy run `34669133578` both succeeded.
- Revision `daily-news-api-00004-fqd` is Ready with 100% traffic and the Go
  image digest `sha256:d1ed333a6c0bc8a5d02d24216f065eb98aa28d60460c2c4e4642a87db0520b12`.
  The service reports ingress `all`, `invokerIamDisabled: true`, and an enabled
  default URI.
- Both Cloud Run-provided URLs return Google edge HTML `404` for `/healthz`.
  The result is identical through all eight resolved IPv4 edge addresses, with
  a Google identity token, and through `gcloud run services proxy`. No matching
  `run.googleapis.com/requests` or VPC Service Controls denial entry exists, so
  the request does not reach the revision. Explicitly enabling the default URL,
  then disabling and re-enabling it, advanced the observed service generation
  from 4 to 7 but did not change the result.
- The deployed Go Job currently invokes `Orchestrator.Run(ctx, id, nil)`.
  Production code has no Source Setting loader or concrete source adapter, and
  the Go `CandidateArticle`/PostgreSQL schema has no citation field. Therefore a
  Run can only terminate with zero source work and cannot satisfy the public
  source/citation acceptance criterion.

Result: **blocked before authenticated or data-mutating smoke steps**. The
Cloud Run hostname registration needs platform repair or service recreation,
and the ingestion/citation gap must be returned to implementation planning
before O4 can pass.
