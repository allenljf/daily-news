package ingestion

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWebSourceAdapterDiscoversRelativeSameHostFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/":
			writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = writer.Write([]byte(`<html><head><link rel="alternate" type="application/rss+xml" title="Feed" href="/feed.xml"></head><body>Home</body></html>`))
		case "/feed.xml":
			writer.Header().Set("Content-Type", "application/rss+xml")
			_, _ = writer.Write([]byte(`<?xml version="1.0"?><rss><channel><item><title>Discovered</title><link>https://news.example/discovered</link></item></channel></rss>`))
		default:
			t.Fatalf("unexpected path %q", request.URL.Path)
		}
	}))
	defer server.Close()

	adapter := NewWebSourceAdapter(newFetcher(allowAllIPs), nil)
	result, err := adapter.Search(context.Background(), SourceWork{WebsiteInput: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].CanonicalURL != "https://news.example/discovered" {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
}

func TestWebSourceAdapterReturnsDirectFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><rss><channel><item><title>Direct</title><link>https://news.example/direct</link></item></channel></rss>`))
	}))
	defer server.Close()

	adapter := NewWebSourceAdapter(newFetcher(allowAllIPs), nil)
	result, err := adapter.Search(context.Background(), SourceWork{WebsiteInput: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].CanonicalURL != "https://news.example/direct" {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
}

func TestWebSourceAdapterRejectsCrossHostDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte(`<html><head><link rel="alternate" type="application/rss+xml" href="https://other.example/feed.xml"></head></html>`))
	}))
	defer server.Close()

	adapter := NewWebSourceAdapter(newFetcher(allowAllIPs), nil)
	if _, err := adapter.Search(context.Background(), SourceWork{WebsiteInput: server.URL}); !errors.Is(err, ErrWebSearchNotConfigured) {
		t.Fatalf("err = %v, want ErrWebSearchNotConfigured", err)
	}
}

func TestWebSourceAdapterRejectsInvalidDiscoveryAndOversizedHomepage(t *testing.T) {
	t.Run("invalid discovery link", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/html")
			_, _ = writer.Write([]byte(`<html><head><link rel="alternate" type="application/rss+xml" href="javascript:alert(1)"></head></html>`))
		}))
		defer server.Close()

		adapter := NewWebSourceAdapter(newFetcher(allowAllIPs), nil)
		if _, err := adapter.Search(context.Background(), SourceWork{WebsiteInput: server.URL}); err == nil {
			t.Fatal("expected an error for an invalid discovery link")
		}
	})

	t.Run("oversized homepage", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "text/html")
			_, _ = writer.Write([]byte(`<html>` + strings.Repeat("x", 200) + `</html>`))
		}))
		defer server.Close()

		fetcher := newFetcher(allowAllIPs)
		fetcher.maxBytes = 64
		adapter := NewWebSourceAdapter(fetcher, nil)
		if _, err := adapter.Search(context.Background(), SourceWork{WebsiteInput: server.URL}); err == nil {
			t.Fatal("expected an error for an oversized homepage")
		}
	})
}

func TestWebSourceAdapterFallsBackToWebSearchWhenNoFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte(`<html><body>No feed here</body></html>`))
	}))
	defer server.Close()

	search := recordingAdapter{name: "search"}
	adapter := NewWebSourceAdapter(newFetcher(allowAllIPs), search)
	if _, err := adapter.Search(context.Background(), SourceWork{WebsiteInput: server.URL}); err != nil {
		t.Fatal(err)
	}
}
