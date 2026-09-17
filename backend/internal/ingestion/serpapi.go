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

// SerpAPIAdapter performs a Google News keyword search through SerpApi. It is
// the whole-web fallback for a Source Setting that exposes no RSS/Atom feed;
// an explicit website is restricted with the `site:` operator.
type SerpAPIAdapter struct {
	client   HTTPDoer
	apiKey   string
	when     string
	endpoint string
}

func NewSerpAPIAdapter(client HTTPDoer, apiKey, when string) *SerpAPIAdapter {
	when = strings.TrimSpace(when)
	if when == "" {
		when = "7d"
	}
	return &SerpAPIAdapter{
		client:   client,
		apiKey:   strings.TrimSpace(apiKey),
		when:     when,
		endpoint: "https://serpapi.com/search",
	}
}

func (adapter *SerpAPIAdapter) configured() bool {
	return adapter != nil && adapter.apiKey != ""
}

func (adapter *SerpAPIAdapter) Search(ctx context.Context, work SourceWork) (SourceSearchResult, error) {
	if !adapter.configured() {
		return SourceSearchResult{}, ErrWebSearchNotConfigured
	}
	query := searchQuery(work)
	if query == "" {
		return SourceSearchResult{}, errors.New("web search has no query terms")
	}
	if host := sourceHost(work.WebsiteInput); host != "" {
		query += " site:" + host
	}
	if adapter.when != "" {
		query += " when:" + adapter.when
	}
	language, country := serpAPILocale(work.ContentLanguage)
	values := url.Values{}
	values.Set("engine", "google_news")
	values.Set("q", query)
	values.Set("hl", language)
	values.Set("gl", country)
	values.Set("api_key", adapter.apiKey)
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
		Error       string `json:"error"`
		NewsResults []struct {
			Title   string `json:"title"`
			Link    string `json:"link"`
			Date    string `json:"date"`
			ISODate string `json:"iso_date"`
			Snippet string `json:"snippet"`
			Source  struct {
				Name string `json:"name"`
			} `json:"source"`
		} `json:"news_results"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return SourceSearchResult{}, fmt.Errorf("parse web search response: %w", err)
	}
	if strings.TrimSpace(payload.Error) != "" {
		return SourceSearchResult{}, errors.New("web search rejected the request")
	}
	result := SourceSearchResult{}
	for _, item := range payload.NewsResults {
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
		if err != nil {
			continue
		}
		summary := strings.TrimSpace(item.Snippet)
		var publishedAt *time.Time
		if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(item.ISODate)); err == nil {
			publishedAt = &parsed
		}
		result.Candidates = append(result.Candidates, CandidateArticle{
			Title:        title,
			CanonicalURL: canonical,
			CitationURL:  citation,
			Summary:      &summary,
			PublishedAt:  publishedAt,
		})
	}
	return result, nil
}

func serpAPILocale(contentLanguage string) (string, string) {
	if normalizeContentLanguage(strings.TrimSpace(contentLanguage)) == "en" {
		return "en", "us"
	}
	return "zh-tw", "tw"
}
