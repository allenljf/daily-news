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

	articles, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 {
		t.Fatalf("article count = %d", len(articles))
	}
	if articles[0].CitationURL != "https://news.example/article?utm_source=feed" {
		t.Fatalf("citation = %q", articles[0].CitationURL)
	}
	if articles[0].CanonicalURL != "https://news.example/article" {
		t.Fatalf("canonical = %q", articles[0].CanonicalURL)
	}
}

func TestRSSAdapterReadsAtomEntries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/atom+xml")
		_, _ = writer.Write([]byte(`<?xml version="1.0"?><feed xmlns="http://www.w3.org/2005/Atom"><entry><title>Atom News</title><link href="https://news.example/atom?utm_campaign=test"/><summary>Atom summary</summary></entry></feed>`))
	}))
	defer server.Close()

	articles, err := NewRSSAdapter(server.Client()).Search(context.Background(), SourceWork{WebsiteInput: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 1 || articles[0].CanonicalURL != "https://news.example/atom" {
		t.Fatalf("articles = %#v", articles)
	}
}
