package ingestion

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RSSAdapter reads direct public RSS feeds configured as Source Settings.
type RSSAdapter struct{ client *http.Client }

func NewRSSAdapter(client *http.Client) *RSSAdapter { return &RSSAdapter{client: client} }

func (adapter *RSSAdapter) Search(ctx context.Context, work SourceWork) ([]CandidateArticle, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, work.WebsiteInput, nil)
	if err != nil {
		return nil, fmt.Errorf("build feed request: %w", err)
	}
	response, err := adapter.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned %s", response.Status)
	}
	var feed struct {
		Channel struct {
			Items []feedEntry `xml:"item"`
		} `xml:"channel"`
		Entries []feedEntry `xml:"entry"`
	}
	if err := xml.NewDecoder(io.LimitReader(response.Body, 2<<20)).Decode(&feed); err != nil {
		return nil, fmt.Errorf("parse rss: %w", err)
	}
	entries := append(feed.Channel.Items, feed.Entries...)
	articles := make([]CandidateArticle, 0, len(entries))
	for _, item := range entries {
		citation := strings.TrimSpace(item.Link.Href)
		if citation == "" {
			citation = strings.TrimSpace(item.Link.Text)
		}
		canonical, err := canonicalizeURL(citation)
		if err != nil || strings.TrimSpace(item.Title) == "" {
			continue
		}
		summary := strings.TrimSpace(item.Description)
		if summary == "" {
			summary = strings.TrimSpace(item.Summary)
		}
		articles = append(articles, CandidateArticle{Title: strings.TrimSpace(item.Title), CanonicalURL: canonical, CitationURL: citation, Summary: &summary})
		if len(articles) == maxCandidatesPerSource {
			break
		}
	}
	return articles, nil
}

type feedEntry struct {
	Title string `xml:"title"`
	Link  struct {
		Href string `xml:"href,attr"`
		Text string `xml:",chardata"`
	} `xml:"link"`
	Description string `xml:"description"`
	Summary     string `xml:"summary"`
}

func canonicalizeURL(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return "", fmt.Errorf("not a public http url")
	}
	query := parsed.Query()
	for key := range query {
		if strings.HasPrefix(strings.ToLower(key), "utm_") {
			query.Del(key)
		}
	}
	parsed.RawQuery = query.Encode()
	parsed.Fragment = ""
	return parsed.String(), nil
}
