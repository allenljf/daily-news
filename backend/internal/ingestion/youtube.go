package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrYouTubeNotConfigured isolates a missing YouTube API key to one source.
var ErrYouTubeNotConfigured = errors.New("YouTube search is not configured")

// YouTubeAdapter searches YouTube Data API v3 for public videos.
type YouTubeAdapter struct {
	client   HTTPDoer
	apiKey   string
	endpoint string
}

func NewYouTubeAdapter(client HTTPDoer, apiKey string) *YouTubeAdapter {
	return &YouTubeAdapter{
		client:   client,
		apiKey:   strings.TrimSpace(apiKey),
		endpoint: "https://www.googleapis.com/youtube/v3/search",
	}
}

func (adapter *YouTubeAdapter) Search(ctx context.Context, work SourceWork) (SourceSearchResult, error) {
	if adapter == nil || adapter.apiKey == "" {
		return SourceSearchResult{}, ErrYouTubeNotConfigured
	}
	query := searchQuery(work)
	if query == "" {
		return SourceSearchResult{}, errors.New("YouTube search has no query terms")
	}
	values := url.Values{}
	values.Set("part", "snippet")
	values.Set("type", "video")
	values.Set("maxResults", "10")
	values.Set("q", query)
	if language := relevanceLanguage(work.ContentLanguage); language != "" {
		values.Set("relevanceLanguage", language)
	}
	values.Set("key", adapter.apiKey)
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, adapter.endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return SourceSearchResult{}, fmt.Errorf("build youtube request: %w", err)
	}
	response, err := adapter.client.Do(request)
	if err != nil {
		return SourceSearchResult{}, fmt.Errorf("youtube request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return SourceSearchResult{}, fmt.Errorf("youtube returned %s", response.Status)
	}
	data, err := readBounded(response.Body, maxResponseBytes)
	if err != nil {
		return SourceSearchResult{}, err
	}
	var payload struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
				PublishedAt string `json:"publishedAt"`
			} `json:"snippet"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return SourceSearchResult{}, fmt.Errorf("parse youtube response: %w", err)
	}
	result := SourceSearchResult{}
	for _, item := range payload.Items {
		if result.CandidateCount == maxCandidatesPerSource {
			break
		}
		result.CandidateCount++
		videoID := strings.TrimSpace(item.ID.VideoID)
		title := strings.TrimSpace(item.Snippet.Title)
		if videoID == "" || title == "" {
			continue
		}
		canonical := "https://www.youtube.com/watch?v=" + videoID
		summary := strings.TrimSpace(item.Snippet.Description)
		var publishedAt *time.Time
		if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(item.Snippet.PublishedAt)); err == nil {
			publishedAt = &parsed
		}
		result.Candidates = append(result.Candidates, CandidateArticle{
			Title:        title,
			CanonicalURL: canonical,
			CitationURL:  canonical,
			Summary:      &summary,
			PublishedAt:  publishedAt,
		})
	}
	return result, nil
}

// searchQuery composes the Category name, keywords and special requirements.
func searchQuery(work SourceWork) string {
	parts := make([]string, 0, 3)
	if name := strings.TrimSpace(work.CategoryName); name != "" {
		parts = append(parts, name)
	}
	if work.SearchKeywords != nil && strings.TrimSpace(*work.SearchKeywords) != "" {
		parts = append(parts, strings.TrimSpace(*work.SearchKeywords))
	}
	if work.SpecialRequirements != nil && strings.TrimSpace(*work.SpecialRequirements) != "" {
		parts = append(parts, strings.TrimSpace(*work.SpecialRequirements))
	}
	return strings.Join(parts, " ")
}

func relevanceLanguage(contentLanguage string) string {
	return strings.TrimSpace(contentLanguage)
}
