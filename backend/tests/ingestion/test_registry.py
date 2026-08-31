"""Contracts for platform Source Setting adapter routing."""

from app.ingestion.registry import AdapterRegistry


class FakeAdapter:
    async def search(self, request):
        return []


def test_registry_routes_youtube_and_github_hosts_to_official_adapters() -> None:
    """A wrong route could send platform content through an unauthorized crawler."""
    youtube = FakeAdapter()
    github = FakeAdapter()
    registry = AdapterRegistry(youtube_adapter=youtube, github_adapter=github, meta_adapter=None)

    assert registry.for_website("https://youtube.com/@daily") is youtube
    assert registry.for_website("https://github.com/openai") is github


def test_registry_returns_disabled_meta_adapter_without_fallback() -> None:
    """Missing Meta authorization must be explicit instead of using a general web adapter."""
    registry = AdapterRegistry(
        youtube_adapter=None,
        github_adapter=FakeAdapter(),
        meta_adapter=None,
    )

    result = registry.for_website("https://www.instagram.com/example")

    assert result.is_disabled is True
    assert result.reason == "Meta adapter is not enabled."
