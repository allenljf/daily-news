package ingestion

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// RSSAdapter reads direct public RSS feeds configured as Source Settings.
type RSSAdapter struct{ client HTTPDoer }

func NewRSSAdapter(client HTTPDoer) *RSSAdapter { return &RSSAdapter{client: client} }

func (adapter *RSSAdapter) Search(ctx context.Context, work SourceWork) (SourceSearchResult, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, work.WebsiteInput, nil)
	if err != nil {
		return SourceSearchResult{}, fmt.Errorf("build feed request: %w", err)
	}
	response, err := adapter.client.Do(request)
	if err != nil {
		return SourceSearchResult{}, fmt.Errorf("fetch feed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return SourceSearchResult{}, fmt.Errorf("feed returned %s", response.Status)
	}
	body, err := readBounded(response.Body, maxResponseBytes)
	if err != nil {
		return SourceSearchResult{}, err
	}
	return parseFeed(body, work)
}

// parseFeed accepts only documents whose root element is an RSS, Atom or RDF feed.
func parseFeed(data []byte, work SourceWork) (SourceSearchResult, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		token, err := decoder.Token()
		if err != nil {
			return SourceSearchResult{}, fmt.Errorf("parse feed: %w", err)
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		switch strings.ToLower(start.Name.Local) {
		case "rss", "feed", "rdf":
			return decodeFeed(decoder, start, work)
		default:
			return SourceSearchResult{}, errors.New("not a feed document")
		}
	}
}

func decodeFeed(decoder *xml.Decoder, start xml.StartElement, work SourceWork) (SourceSearchResult, error) {
	var feed struct {
		Language string `xml:"lang,attr"`
		Channel  struct {
			Language string      `xml:"language"`
			Items    []feedEntry `xml:"item"`
		} `xml:"channel"`
		Entries []feedEntry `xml:"entry"`
	}
	if err := decoder.DecodeElement(&feed, &start); err != nil {
		return SourceSearchResult{}, fmt.Errorf("parse feed: %w", err)
	}
	entries := append(feed.Channel.Items, feed.Entries...)
	result := SourceSearchResult{Candidates: make([]CandidateArticle, 0, len(entries))}
	for _, item := range entries {
		citation := strings.TrimSpace(item.Link.Href)
		if citation == "" {
			citation = strings.TrimSpace(item.Link.Text)
		}
		canonical, err := canonicalizeURL(citation)
		if err != nil || strings.TrimSpace(item.Title) == "" {
			continue
		}
		if result.CandidateCount == maxCandidatesPerSource {
			break
		}
		result.CandidateCount++
		if !matchesContentLanguage(work.ContentLanguage, item.LanguageAttribute, item.Language, feed.Channel.Language, feed.Language) {
			continue
		}
		summary := strings.TrimSpace(item.Description)
		if summary == "" {
			summary = strings.TrimSpace(item.Summary)
		}
		result.Candidates = append(result.Candidates, CandidateArticle{Title: strings.TrimSpace(item.Title), CanonicalURL: canonical, CitationURL: citation, Summary: &summary})
	}
	return result, nil
}

type feedEntry struct {
	Title string `xml:"title"`
	Link  struct {
		Href string `xml:"href,attr"`
		Text string `xml:",chardata"`
	} `xml:"link"`
	Description       string `xml:"description"`
	Summary           string `xml:"summary"`
	LanguageAttribute string `xml:"lang,attr"`
	Language          string `xml:"language"`
}

func matchesContentLanguage(requested string, values ...string) bool {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return strings.EqualFold(value, strings.TrimSpace(requested))
		}
	}
	return true
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
