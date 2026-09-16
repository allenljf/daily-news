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

func TestYouTubeAdapterBuildsQueryAndMapsPublicVideoURLs(t *testing.T) {
	var capturedQuery map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		capturedQuery = map[string]string{
			"part":              query.Get("part"),
			"type":              query.Get("type"),
			"maxResults":        query.Get("maxResults"),
			"q":                 query.Get("q"),
			"relevanceLanguage": query.Get("relevanceLanguage"),
			"key":               query.Get("key"),
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"items": []map[string]any{{
				"id": map[string]any{"kind": "youtube#video", "videoId": "abc123"},
				"snippet": map[string]any{
					"title":       "Video title",
					"description": "Video description",
					"publishedAt": "2026-09-01T00:00:00Z",
				},
			}},
		})
	}))
	defer server.Close()

	keywords := "kubernetes"
	special := "conference talks"
	adapter := NewYouTubeAdapter(server.Client(), "test-key")
	adapter.endpoint = server.URL
	result, err := adapter.Search(context.Background(), SourceWork{
		CategoryName:        "Cloud",
		ContentLanguage:     "zh-Hant",
		SearchKeywords:      &keywords,
		SpecialRequirements: &special,
		WebsiteInput:        "https://www.youtube.com/@channel",
	})
	if err != nil {
		t.Fatal(err)
	}
	if capturedQuery["part"] != "snippet" || capturedQuery["type"] != "video" || capturedQuery["maxResults"] != "10" {
		t.Fatalf("query = %#v", capturedQuery)
	}
	if capturedQuery["relevanceLanguage"] != "zh-Hant" {
		t.Fatalf("relevanceLanguage = %q", capturedQuery["relevanceLanguage"])
	}
	for _, want := range []string{"Cloud", "kubernetes", "conference talks"} {
		if !strings.Contains(capturedQuery["q"], want) {
			t.Fatalf("q = %q does not contain %q", capturedQuery["q"], want)
		}
	}
	if len(result.Candidates) != 1 {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
	candidate := result.Candidates[0]
	if candidate.CanonicalURL != "https://www.youtube.com/watch?v=abc123" || candidate.CitationURL != candidate.CanonicalURL {
		t.Fatalf("candidate = %#v", candidate)
	}
}

func TestYouTubeAdapterCapsAtTen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		items := make([]map[string]any, 0, 12)
		for index := 0; index < 12; index++ {
			items = append(items, map[string]any{
				"id":      map[string]any{"kind": "youtube#video", "videoId": fmt.Sprintf("video-%d", index)},
				"snippet": map[string]any{"title": fmt.Sprintf("Video %d", index), "description": ""},
			})
		}
		writer.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(writer).Encode(map[string]any{"items": items})
	}))
	defer server.Close()

	adapter := NewYouTubeAdapter(server.Client(), "test-key")
	adapter.endpoint = server.URL
	result, err := adapter.Search(context.Background(), SourceWork{CategoryName: "Cloud", WebsiteInput: "https://www.youtube.com"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != maxCandidatesPerSource || result.CandidateCount != maxCandidatesPerSource {
		t.Fatalf("candidates = %d, count = %d", len(result.Candidates), result.CandidateCount)
	}
}

func TestYouTubeAdapterWithoutKeyIsNotConfigured(t *testing.T) {
	adapter := NewYouTubeAdapter(http.DefaultClient, "")
	if _, err := adapter.Search(context.Background(), SourceWork{CategoryName: "Cloud"}); !errors.Is(err, ErrYouTubeNotConfigured) {
		t.Fatalf("err = %v, want ErrYouTubeNotConfigured", err)
	}
}

func TestYouTubeAdapterErrorDoesNotLeakAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "forbidden", http.StatusForbidden)
	}))
	defer server.Close()

	adapter := NewYouTubeAdapter(server.Client(), "super-secret-key")
	adapter.endpoint = server.URL
	_, err := adapter.Search(context.Background(), SourceWork{CategoryName: "Cloud", WebsiteInput: "https://www.youtube.com"})
	if err == nil {
		t.Fatal("expected an error")
	}
	if strings.Contains(err.Error(), "super-secret-key") {
		t.Fatalf("error leaked the API key: %v", err)
	}
}
