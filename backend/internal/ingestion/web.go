package ingestion

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var (
	htmlLinkTagPattern   = regexp.MustCompile(`(?is)<link\b[^>]*>`)
	htmlAttributePattern = regexp.MustCompile(`(?is)([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*=\s*("([^"]*)"|'([^']*)'|([^\s"'>]+))`)
)

// WebSourceAdapter composes direct feed parsing, same-host feed discovery and
// the Google Custom Search fallback behind the SourceAdapter contract.
type WebSourceAdapter struct {
	fetcher *Fetcher
	search  SourceAdapter
}

func NewWebSourceAdapter(fetcher *Fetcher, search SourceAdapter) *WebSourceAdapter {
	return &WebSourceAdapter{fetcher: fetcher, search: search}
}

func (adapter *WebSourceAdapter) Search(ctx context.Context, work SourceWork) (SourceSearchResult, error) {
	fetched, fetchErr := adapter.fetcher.Fetch(ctx, work.WebsiteInput)
	if fetchErr == nil {
		if result, err := parseFeed(fetched.Body, work); err == nil {
			return result, nil
		}
		for _, href := range discoverFeedLinks(fetched.Body, fetched.FinalURL) {
			if !withinSourceHost(work.WebsiteInput, href) {
				continue
			}
			discovered, err := adapter.fetcher.Fetch(ctx, href)
			if err != nil {
				continue
			}
			if result, err := parseFeed(discovered.Body, work); err == nil {
				return result, nil
			}
		}
	}
	if adapter.search == nil {
		if fetchErr != nil {
			return SourceSearchResult{}, fmt.Errorf("fetch source homepage: %w", fetchErr)
		}
		return SourceSearchResult{}, ErrWebSearchNotConfigured
	}
	return adapter.search.Search(ctx, work)
}

// discoverFeedLinks extracts RSS/Atom alternate link URLs and resolves them
// against the fetched document URL.
func discoverFeedLinks(body []byte, baseURL string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}
	var hrefs []string
	for _, tag := range htmlLinkTagPattern.FindAll(body, -1) {
		attributes := htmlAttributes(tag)
		if !containsToken(attributes["rel"], "alternate") || !isFeedLinkType(attributes["type"]) {
			continue
		}
		href := strings.TrimSpace(attributes["href"])
		if href == "" {
			continue
		}
		reference, err := url.Parse(href)
		if err != nil {
			continue
		}
		resolved := base.ResolveReference(reference)
		if (resolved.Scheme != "http" && resolved.Scheme != "https") || resolved.Hostname() == "" {
			continue
		}
		hrefs = append(hrefs, resolved.String())
	}
	return hrefs
}

func htmlAttributes(tag []byte) map[string]string {
	attributes := map[string]string{}
	for _, match := range htmlAttributePattern.FindAllSubmatch(tag, -1) {
		value := ""
		for _, candidate := range [][]byte{match[3], match[4], match[5]} {
			if len(candidate) > 0 {
				value = string(candidate)
				break
			}
		}
		attributes[strings.ToLower(string(match[1]))] = value
	}
	return attributes
}

func containsToken(value, token string) bool {
	for _, field := range strings.Fields(strings.ToLower(value)) {
		if field == token {
			return true
		}
	}
	return false
}

func isFeedLinkType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return strings.Contains(contentType, "rss") || strings.Contains(contentType, "atom")
}

// withinSourceHost reports whether candidate belongs to the Source Setting host
// or one of its subdomains. An empty Source Setting host imposes no restriction.
func withinSourceHost(source, candidate string) bool {
	baseHost := sourceHost(source)
	if baseHost == "" {
		return true
	}
	candidateHost := sourceHost(candidate)
	if candidateHost == "" {
		return false
	}
	return candidateHost == baseHost || strings.HasSuffix(candidateHost, "."+baseHost)
}

func sourceHost(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	if scheme := strings.ToLower(parsed.Scheme); scheme != "http" && scheme != "https" {
		return ""
	}
	return strings.TrimPrefix(strings.ToLower(parsed.Hostname()), "www.")
}
