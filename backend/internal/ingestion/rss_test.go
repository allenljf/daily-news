package ingestion

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRSSAdapterReturnsPublicCitationAndCanonicalURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><rss><channel><item><title>Daily News</title><link>https://news.example/article?utm_source=feed</link><description>Summary</description></item></channel></rss>`))
	}))
	defer server.Close()

	result, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("article count = %d", len(result.Candidates))
	}
	if result.Candidates[0].CitationURL != "https://news.example/article?utm_source=feed" {
		t.Fatalf("citation = %q", result.Candidates[0].CitationURL)
	}
	if result.Candidates[0].CanonicalURL != "https://news.example/article" {
		t.Fatalf("canonical = %q", result.Candidates[0].CanonicalURL)
	}
}

func TestRSSAdapterReadsAtomEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/atom+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Atom News</title><link href="https://news.example/atom?utm_campaign=test"/><summary>Atom summary</summary></entry></feed>`))
	}))
	defer server.Close()

	result, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].CanonicalURL != "https://news.example/atom" {
		t.Fatalf("articles = %#v", result.Candidates)
	}
}

func TestRSSAdapterRejectsKnownNonTraditionalLanguageMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><rss><channel><item><title>Traditional News</title><link>https://news.example/traditional</link><language>zh-Hant</language></item><item><title>Simplified News</title><link>https://news.example/simplified</link><language>zh-Hans</language></item></channel></rss>`))
	}))
	defer server.Close()

	result, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL, ContentLanguage: "zh-Hant"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("article count = %d, want 1", len(result.Candidates))
	}
	if result.Candidates[0].Title != "Traditional News" {
		t.Fatalf("retained title = %q, want Traditional News", result.Candidates[0].Title)
	}
}

func TestRSSAdapterAcceptsTraditionalChineseVariants(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><rss><channel><language>zh-TW</language><item><title>台灣新聞</title><link>https://news.example/tw</link></item></channel></rss>`))
	}))
	defer server.Close()

	result, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL, ContentLanguage: "zh-Hant"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("article count = %d, want 1", len(result.Candidates))
	}
}

func TestRSSAdapterRetainsOnlyEnglishEntriesForEnglishCategory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/rss+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><rss><channel><item><title>English News</title><link>https://news.example/english</link><language>en</language></item><item><title>Traditional News</title><link>https://news.example/traditional</link><language>zh-Hant</language></item></channel></rss>`))
	}))
	defer server.Close()

	result, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL, ContentLanguage: "en"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("article count = %d, want 1", len(result.Candidates))
	}
	if result.Candidates[0].Title != "English News" {
		t.Fatalf("retained title = %q, want English News", result.Candidates[0].Title)
	}
}

func TestRSSAdapterRejectsNonTraditionalAtomEntryLanguage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/atom+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry xml:lang="zh-Hant"><title>Traditional Atom News</title><link href="https://news.example/traditional-atom"/></entry><entry xml:lang="zh-Hans"><title>Simplified Atom News</title><link href="https://news.example/simplified-atom"/></entry><entry><title>Unlabelled Atom News</title><link href="https://news.example/unlabelled-atom"/></entry></feed>`))
	}))
	defer server.Close()

	result, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL, ContentLanguage: "zh-Hant"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("article count = %d, want 2", len(result.Candidates))
	}
	if result.Candidates[0].Title != "Traditional Atom News" || result.Candidates[1].Title != "Unlabelled Atom News" {
		t.Fatalf("retained titles = %q, %q", result.Candidates[0].Title, result.Candidates[1].Title)
	}
}
