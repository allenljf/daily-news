package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGoogleSearchAdapterBuildsQueryAndKeepsSameHostCandidates(t *testing.T) {
	var baseURL string
	var captured map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		captured = map[string]string{
			"key":          query.Get("key"),
			"cx":           query.Get("cx"),
			"q":            query.Get("q"),
			"num":          query.Get("num"),
			"dateRestrict": query.Get("dateRestrict"),
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"items": []map[string]any{
				{
					"title":   "Same host",
					"link":    baseURL + "/article?utm_source=search",
					"snippet": "Summary",
					"pagemap": map[string]any{"metatags": []map[string]string{{"article:published_time": "2026-09-01T00:00:00Z"}}},
				},
				{"title": "Off host", "link": "https://other.example/article", "snippet": "Off"},
				{"title": "", "link": baseURL + "/untitled", "snippet": "Untitled"},
			},
		})
	}))
	defer server.Close()
	baseURL = server.URL

	keywords := "ai safety"
	special := "prefer primary sources"
	adapter := NewGoogleSearchAdapter(server.Client(), "test-key", "test-cx", "m1")
	adapter.endpoint = server.URL
	result, err := adapter.Search(context.Background(), SourceWork{
		CategoryName:        "Technology",
		SearchKeywords:      &keywords,
		SpecialRequirements: &special,
		WebsiteInput:        server.URL,
	})
	if err != nil {
		t.Fatal(err)
	}
	if captured["key"] != "test-key" || captured["cx"] != "test-cx" || captured["num"] != "10" || captured["dateRestrict"] != "m1" {
		t.Fatalf("query = %#v", captured)
	}
	for _, want := range []string{"Technology", "ai safety", "prefer primary sources"} {
		if !strings.Contains(captured["q"], want) {
			t.Fatalf("q = %q does not contain %q", captured["q"], want)
		}
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
	candidate := result.Candidates[0]
	if candidate.CanonicalURL != server.URL+"/article" {
		t.Fatalf("canonical = %q", candidate.CanonicalURL)
	}
	if candidate.PublishedAt == nil {
		t.Fatal("published_at was not extracted from the CSE metatags")
	}
}

func TestGoogleSearchAdapterCapsAtTen(t *testing.T) {
	var baseURL string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		items := make([]map[string]any, 0, 12)
		for index := 0; index < 12; index++ {
			items = append(items, map[string]any{"title": "Article", "link": fmt.Sprintf("%s/article-%d", baseURL, index)})
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{"items": items})
	}))
	defer server.Close()
	baseURL = server.URL

	adapter := NewGoogleSearchAdapter(server.Client(), "test-key", "test-cx", "")
	adapter.endpoint = server.URL
	result, err := adapter.Search(context.Background(), SourceWork{CategoryName: "Technology", WebsiteInput: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != maxCandidatesPerSource || result.CandidateCount != maxCandidatesPerSource {
		t.Fatalf("candidates = %d, count = %d", len(result.Candidates), result.CandidateCount)
	}
}

func TestGoogleSearchAdapterWithoutCredentialsIsNotConfigured(t *testing.T) {
	for _, adapter := range []*GoogleSearchAdapter{
		NewGoogleSearchAdapter(http.DefaultClient, "", "test-cx", ""),
		NewGoogleSearchAdapter(http.DefaultClient, "test-key", "", ""),
	} {
		if _, err := adapter.Search(context.Background(), SourceWork{CategoryName: "Technology"}); !errors.Is(err, ErrWebSearchNotConfigured) {
			t.Fatalf("err = %v, want ErrWebSearchNotConfigured", err)
		}
	}
}

func TestGoogleSearchAdapterErrorDoesNotLeakAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	adapter := NewGoogleSearchAdapter(server.Client(), "super-secret-key", "test-cx", "")
	adapter.endpoint = server.URL
	_, err := adapter.Search(context.Background(), SourceWork{CategoryName: "Technology", WebsiteInput: server.URL})
	if err == nil || strings.Contains(err.Error(), "super-secret-key") {
		t.Fatalf("err = %v", err)
	}
}
