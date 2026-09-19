# Web Source Adapter Design

## Purpose

Allow a Category Source Setting containing a public website homepage to produce
Articles. The current Job treats every HTTP(S) input as RSS/Atom and therefore
records a failed attempt when the URL returns HTML. This design retains direct
feed support and adds bounded fallback paths without changing the mobile or
HTTP API contract.

## Source flow

For each active Source Setting, the Job uses one composite adapter. An explicit
HTTP(S) website:

1. Request the configured URL with a bounded context and response-size limit.
2. If it is a valid RSS or Atom document, parse it using the existing feed
   parser.
3. If it is HTML, inspect only `link` elements whose `rel` contains
   `alternate` and whose type is `application/rss+xml` or `application/atom+xml`.
   Resolve a relative `href` against the configured URL, require public HTTP(S),
   and fetch the first eligible discovered feed through the existing parser.
4. If no eligible feed is found, or the feed fetch/parsing fails, call the
   SerpApi Google News search (`engine=google_news`). The query includes Category
   name, keywords and special requirements, restricted with `site:<host>`, and
   `SERPAPI_WHEN` bounds recency. An unspecified Source Setting has no host to
   fetch and goes straight to a whole-web Google News search without `site:`.
5. Accept at most ten candidates. Each must have a non-empty title and a public
   HTTP(S) URL. The adapter maps the result link to
   `CandidateArticle.CitationURL`; the orchestrator remains the only writer and
   dedupe authority.

The composite adapter returns a source error only after all eligible fallbacks
fail. Existing per-source failure isolation, attempt accounting, 30-day expiry,
global suppression, and URL-then-title dedupe are unchanged.

YouTube is no longer ingested. `youtube.com` and `youtu.be` Source Settings are
routed to an explicit unsupported error instead of any keyword search, and every
adapter's YouTube candidates are dropped at the write boundary so no YouTube URL
can be stored. Threads keyword search is out of scope for this change: Threads,
Facebook and Instagram hosts are routed to a restricted Meta boundary that
returns an explicit error and never reaches the general search fallback.

## Boundaries and safety

- Direct feed parsing and discovery use a shared HTTP client with request
  timeout, 2 MiB response limit, redirect cap, and a transport that rejects
  loopback, link-local, private, and otherwise non-public target addresses.
- Discovered feed URLs must resolve to the configured host or a subdomain of it.
  Search results for an explicit website are restricted with `site:<host>`.
- No page body, prompt, API key, or Firebase token is logged. Attempts retain
  a short, redacted error summary only.
- The SerpApi Google News fallback is an optional composition dependency. With
  no `SERPAPI_API_KEY`, the adapter returns the explicit
  `web search is not configured` source error; it does not silently make an
  unrestricted web request or fail the whole Run.
- The Cloud Run Job gets `SERPAPI_API_KEY` from Secret Manager only. It is never
  available to Flutter, checked into the repository, or emitted in tests.

## SerpApi Google News contract

Use `GET https://serpapi.com/search?engine=google_news` with `q`, `api_key`,
`hl` and `gl` derived from the Category content language, and `when:<window>`
for recency. Parse and validate the JSON in Go. Accept only results whose link
is a public HTTP(S) URL, cap the candidates at ten, and record
`iso_date` as `published_at` when present.

`SERPAPI_WHEN` is configuration with a safe documented default (`7d`); the
adapter is injected behind the existing `SourceAdapter` interface so unit and
integration tests never call the API.

## Components

- `FeedDiscoveryAdapter`: identifies direct XML vs HTML, extracts safe feed
  discovery links, and delegates parsing to `RSSAdapter`.
- `SerpAPIAdapter`: owns Google News request/response DTOs, `site:` restriction,
  candidate validation, and bounded error translation.
- `WebSourceAdapter`: composes discovery then SerpApi search and exposes the
  existing `SourceAdapter.Search` method to `WorkPlanner`.
- `HostRouter`: rejects Youtube Source Settings with an explicit unsupported
  error and routes Facebook/Instagram/Threads to the restricted Meta boundary;
  everything else uses the composite web adapter.
- The orchestrator write boundary drops candidates whose canonical URL is a
  blocked platform host (currently `youtube.com`/`youtu.be`) before persistence.
- Job composition builds the shared safe HTTP client and composite adapter. The
  Cloud Run manifest injects only the SerpApi secret for production.

## Verification

Tests first demonstrate: an HTML homepage discovers a relative feed; unsafe or
cross-host discovery links are rejected; a homepage without a feed delegates to
a fake search adapter; off-host, malformed, or over-limit search candidates are
rejected; a missing search credential creates one failed attempt without
preventing another source from succeeding; a YouTube Source Setting returns the
explicit unsupported error; and YouTube candidates from any adapter are dropped
before persistence. A Job composition test confirms active homepage Source
Settings use the composite adapter.

Run `gofmt`, focused adapter and Job tests, `go vet ./...`, `go test ./...`,
container build verification, and an authenticated staging Run using a
homepage-only disposable Source Setting. Record only IDs and aggregate counts.
