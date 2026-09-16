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

// ErrWebSearchNotConfigured isolates a missing search credential to one source.
var ErrWebSearchNotConfigured = errors.New("web search is not configured")

// GoogleSearchAdapter uses the Google Custom Search JSON API as the bounded
// fallback for a public website that exposes no discoverable RSS/Atom feed.
type GoogleSearchAdapter struct {
	client       HTTPDoer
	apiKey       string
	engineID     string
	dateRestrict string
	endpoint     string
}

func NewGoogleSearchAdapter(client HTTPDoer, apiKey, engineID, dateRestrict string) *GoogleSearchAdapter {
	return &GoogleSearchAdapter{
		client:       client,
		apiKey:       strings.TrimSpace(apiKey),
		engineID:     strings.TrimSpace(engineID),
		dateRestrict: strings.TrimSpace(dateRestrict),
		endpoint:     "https://www.googleapis.com/customsearch/v1",
	}
}

func (adapter *GoogleSearchAdapter) configured() bool {
	return adapter != nil && adapter.apiKey != "" && adapter.engineID != ""
}

func (adapter *GoogleSearchAdapter) Search(ctx context.Context, work SourceWork) (SourceSearchResult, error) {
	if !adapter.configured() {
		return SourceSearchResult{}, ErrWebSearchNotConfigured
	}
	query := searchQuery(work)
	if query == "" {
		return SourceSearchResult{}, errors.New("web search has no query terms")
	}
	values := url.Values{}
	values.Set("key", adapter.apiKey)
	values.Set("cx", adapter.engineID)
	values.Set("q", query)
	values.Set("num", "10")
	if adapter.dateRestrict != "" {
		values.Set("dateRestrict", adapter.dateRestrict)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, adapter.endpoint+"?"+values.Encode(), nil)
	if err != nil {
		return SourceSearchResult{}, fmt.Errorf("build web search request: %w", err)
	}
	response, err := adapter.client.Do(request)
	if err != nil {
		return SourceSearchResult{}, fmt.Errorf("web search request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return SourceSearchResult{}, fmt.Errorf("web search returned %s", response.Status)
	}
	data, err := readBounded(response.Body, maxResponseBytes)
	if err != nil {
		return SourceSearchResult{}, err
	}
	var payload struct {
		Items []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Snippet string `json:"snippet"`
			PageMap struct {
				MetaTags []map[string]string `json:"metatags"`
			} `json:"pagemap"`
		} `json:"items"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return SourceSearchResult{}, fmt.Errorf("parse web search response: %w", err)
	}
	result := SourceSearchResult{}
	for _, item := range payload.Items {
		if result.CandidateCount == maxCandidatesPerSource {
			break
		}
		result.CandidateCount++
		title := strings.TrimSpace(item.Title)
		citation := strings.TrimSpace(item.Link)
		if title == "" || citation == "" {
			continue
		}
		canonical, err := canonicalizeURL(citation)
		if err != nil || !withinSourceHost(work.WebsiteInput, canonical) {
			continue
		}
		summary := strings.TrimSpace(item.Snippet)
		result.Candidates = append(result.Candidates, CandidateArticle{
			Title:        title,
			CanonicalURL: canonical,
			CitationURL:  citation,
			Summary:      &summary,
			PublishedAt:  publishedTime(item.PageMap.MetaTags),
		})
	}
	return result, nil
}

func publishedTime(metaTags []map[string]string) *time.Time {
	for _, meta := range metaTags {
		for _, key := range []string{"article:published_time", "og:updated_time", "datePublished"} {
			value := strings.TrimSpace(meta[key])
			if value == "" {
				continue
			}
			if parsed, err := time.Parse(time.RFC3339, value); err == nil {
				return &parsed
			}
		}
	}
	return nil
}
