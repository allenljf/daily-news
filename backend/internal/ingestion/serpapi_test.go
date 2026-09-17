package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSerpAPIAdapterBuildsNewsQueryAndMapsResults(t *testing.T) {
	var captured map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		captured = map[string]string{
			"engine":  query.Get("engine"),
			"q":       query.Get("q"),
			"hl":      query.Get("hl"),
			"gl":      query.Get("gl"),
			"api_key": query.Get("api_key"),
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"news_results": []map[string]any{{
				"title":    "台灣 AI 新聞",
				"link":     "https://news.example/article?utm_source=news",
				"iso_date": "2026-09-16T01:00:00Z",
				"snippet":  "摘要",
				"source":   map[string]any{"name": "範例新聞"},
			}, {
				"title": "",
				"link":  "https://news.example/untitled",
			}},
		})
	}))
	defer server.Close()

	keywords := "人工智慧"
	adapter := NewSerpAPIAdapter(server.Client(), "test-key", "7d")
	adapter.endpoint = server.URL
	result, err := adapter.Search(context.Background(), SourceWork{
		CategoryName:    "AI",
		SearchKeywords:  &keywords,
		ContentLanguage: "zh-Hant",
		WebsiteInput:    "https://news.example",
	})
	if err != nil {
		t.Fatal(err)
	}
	if captured["engine"] != "google_news" || captured["hl"] != "zh-tw" || captured["gl"] != "tw" || captured["api_key"] != "test-key" {
		t.Fatalf("query = %#v", captured)
	}
	for _, want := range []string{"AI", "人工智慧", "site:news.example", "when:7d"} {
		if !strings.Contains(captured["q"], want) {
			t.Fatalf("q = %q does not contain %q", captured["q"], want)
		}
	}
	if len(result.Candidates) != 1 || result.Candidates[0].CanonicalURL != "https://news.example/article" {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
	if result.Candidates[0].PublishedAt == nil {
		t.Fatal("published_at was not parsed from iso_date")
	}
}

func TestSerpAPIAdapterSearchOmitsSiteForUnspecifiedSource(t *testing.T) {
	var captured string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		captured = request.URL.Query().Get("q")
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"news_results":[]}`))
	}))
	defer server.Close()

	adapter := NewSerpAPIAdapter(server.Client(), "test-key", "7d")
	adapter.endpoint = server.URL
	if _, err := adapter.Search(context.Background(), SourceWork{CategoryName: "AI", WebsiteInput: ""}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(captured, "site:") {
		t.Fatalf("q = %q must not restrict to a site", captured)
	}
}

func TestSerpAPIAdapterWithoutKeyIsNotConfigured(t *testing.T) {
	adapter := NewSerpAPIAdapter(http.DefaultClient, "", "")
	if _, err := adapter.Search(context.Background(), SourceWork{CategoryName: "AI"}); !errors.Is(err, ErrWebSearchNotConfigured) {
		t.Fatalf("err = %v, want ErrWebSearchNotConfigured", err)
	}
}

func TestSerpAPIAdapterErrorDoesNotLeakAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	adapter := NewSerpAPIAdapter(server.Client(), "super-secret-key", "7d")
	adapter.endpoint = server.URL
	_, err := adapter.Search(context.Background(), SourceWork{CategoryName: "AI", WebsiteInput: "https://news.example"})
	if err == nil || strings.Contains(err.Error(), "super-secret-key") {
		t.Fatalf("err = %v", err)
	}
}
