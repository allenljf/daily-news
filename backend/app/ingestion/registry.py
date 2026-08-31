"""Host-based routing for official platform source adapters."""

from __future__ import annotations

from dataclasses import dataclass
from urllib.parse import urlsplit


@dataclass(frozen=True)
class DisabledAdapter:
    """A deliberate disabled result that lets one Source Setting fail independently."""

    reason: str
    is_disabled: bool = True


class AdapterRegistry:
    """Routes platform hosts without ever selecting the general-web adapter for Meta."""

    def __init__(
        self,
        *,
        youtube_adapter: object | None,
        github_adapter: object | None,
        meta_adapter: object | None,
    ) -> None:
        self._youtube_adapter = youtube_adapter
        self._github_adapter = github_adapter
        self._meta_adapter = meta_adapter

    def for_website(self, website_input: str) -> object:
        host = (urlsplit(website_input).hostname or "").lower().removeprefix("www.")
        if host == "youtube.com" or host.endswith(".youtube.com"):
            return self._youtube_adapter or DisabledAdapter("YouTube adapter is not enabled.")
        if host == "github.com" or host.endswith(".github.com"):
            return self._github_adapter or DisabledAdapter("GitHub adapter is not enabled.")
        if host in {"facebook.com", "instagram.com", "threads.net"} or host.endswith(
            (".facebook.com", ".instagram.com", ".threads.net")
        ):
            return self._meta_adapter or DisabledAdapter("Meta adapter is not enabled.")
        return DisabledAdapter("No platform adapter is registered for this host.")
