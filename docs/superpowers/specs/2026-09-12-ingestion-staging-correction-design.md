# Ingestion staging correction design

## Goal

Make the Go Cloud Run Job turn active Category/Source Setting rows into at
least one real, citation-backed public Article without changing the checked-in
HTTP contract.

## Design

`WorkPlanner` is the ingestion module's deep module: it reads active Category
and Source Setting rows, chooses the single injected public-feed adapter, and
returns complete `SourceWork` values. The Job uses it instead of an empty work
slice. The first production adapter accepts an explicit public RSS or Atom URL
in `website_input`; it has a bounded context, rejects non-HTTP(S) and private
literal hosts, parses RSS/Atom with `encoding/xml`, and returns no more than ten
candidates. Its entry URL is both the citation URL and the input to canonical
URL normalization.

`articles.citation_url` is added by golang-migrate migration `000003`. New
Articles store a non-empty public citation. The existing HTTP JSON stays
unchanged: `canonical_url` remains the public article link, while citation is
verified from PostgreSQL during O4. The Job applies checked-in migrations from
the image before reading/writing; this is safe for the currently empty Neon
database and makes schema upgrade observable through the Job terminal status.

## Verification

Tests use a disposable PostgreSQL instance and fake HTTP transport to prove
planner filtering, RSS/Atom parsing, URL/citation persistence, and Job work
wiring. Full Go tests, container build, and the O4 public smoke test follow.
