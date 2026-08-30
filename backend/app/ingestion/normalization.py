"""Stable canonical URL and title fingerprints for Article deduplication."""

from __future__ import annotations

import hashlib
from urllib.parse import parse_qsl, urlencode, urlsplit, urlunsplit

_TRACKING_PARAMETER_NAMES = {"fbclid", "gclid", "mc_cid", "mc_eid"}


def canonicalize_url(url: str) -> str:
    """Remove fragments and tracking parameters while retaining meaningful URL identity."""
    parsed = urlsplit(url.strip())
    hostname = (parsed.hostname or "").lower()
    if not hostname:
        return url.strip()
    scheme = parsed.scheme.lower()
    port = parsed.port
    netloc = hostname
    if port is not None and (scheme, port) not in {("http", 80), ("https", 443)}:
        netloc = f"{hostname}:{port}"
    query = urlencode(
        sorted(
            (key, value)
            for key, value in parse_qsl(parsed.query, keep_blank_values=True)
            if not _is_tracking_parameter(key)
        ),
        doseq=True,
    )
    return urlunsplit((scheme, netloc, parsed.path or "/", query, ""))


def url_hash(url: str) -> str:
    """Return the SHA-256 fingerprint of a canonical URL."""
    return _sha256(canonicalize_url(url))


def normalize_title(title: str) -> str:
    """Collapse Unicode whitespace and case-fold a human-visible Article title."""
    return " ".join(title.split()).casefold()


def normalized_title_hash(title: str) -> str:
    """Return the SHA-256 fingerprint of a normalized Article title."""
    return _sha256(normalize_title(title))


def _is_tracking_parameter(key: str) -> bool:
    normalized_key = key.casefold()
    return normalized_key.startswith("utm_") or normalized_key in _TRACKING_PARAMETER_NAMES


def _sha256(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()
