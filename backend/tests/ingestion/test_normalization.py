"""Contracts for ingestion URL and title normalization."""

from app.ingestion.normalization import canonicalize_url, normalized_title_hash, url_hash


def test_tracking_parameters_do_not_change_the_canonical_url_hash() -> None:
    """Removing tracking parameters must retain the URL's meaningful query parameters."""
    tracked_url = "HTTPS://Example.COM/news?ref=homepage&utm_source=newsletter#top"
    clean_url = "https://example.com/news?ref=homepage"

    assert canonicalize_url(tracked_url) == clean_url
    assert url_hash(tracked_url) == url_hash(clean_url)


def test_title_hash_ignores_case_and_extra_whitespace() -> None:
    """The same human-visible title must deduplicate despite cosmetic formatting changes."""
    compact_title = "OpenAI releases a model"
    noisy_title = "  OPENAI\u3000 releases\n a   model  "

    assert normalized_title_hash(compact_title) == normalized_title_hash(noisy_title)
