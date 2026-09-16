package ingestion

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

const (
	maxResponseBytes int64 = 2 << 20
	maxRedirects           = 5
)

// HTTPDoer is the narrow outbound HTTP boundary used by ingestion adapters.
type HTTPDoer interface {
	Do(request *http.Request) (*http.Response, error)
}

// FetchResult is a bounded public HTTP response.
type FetchResult struct {
	Body        []byte
	ContentType string
	FinalURL    string
}

// Fetcher performs bounded, SSRF-safe outbound requests.
type Fetcher struct {
	client   *http.Client
	maxBytes int64
}

// NewSafeFetcher builds a Fetcher that refuses to dial non-public addresses.
func NewSafeFetcher() *Fetcher { return newFetcher(isPublicIP) }

func newFetcher(allow func(net.IP) bool) *Fetcher {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				return safeDial(ctx, dialer, allow, network, address)
			},
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: time.Second,
			IdleConnTimeout:       30 * time.Second,
			DisableKeepAlives:     true,
		},
		Timeout: 15 * time.Second,
		CheckRedirect: func(_ *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("source redirected too many times")
			}
			return nil
		},
	}
	return &Fetcher{client: client, maxBytes: maxResponseBytes}
}

// Do exposes the bounded client so platform adapters reuse the same transport.
func (fetcher *Fetcher) Do(request *http.Request) (*http.Response, error) {
	return fetcher.client.Do(request)
}

// Fetch retrieves a public HTTP(S) document within the response-size limit.
func (fetcher *Fetcher) Fetch(ctx context.Context, rawURL string) (FetchResult, error) {
	if !isPublicHTTPURL(rawURL) {
		return FetchResult{}, errors.New("not a public http url")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return FetchResult{}, fmt.Errorf("build source request: %w", err)
	}
	request.Header.Set("User-Agent", "DailyNews/1.0")
	request.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/html;q=0.8, */*;q=0.1")
	response, err := fetcher.client.Do(request)
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetch source: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return FetchResult{}, fmt.Errorf("source returned %s", response.Status)
	}
	body, err := readBounded(response.Body, fetcher.maxBytes)
	if err != nil {
		return FetchResult{}, err
	}
	finalURL := rawURL
	if response.Request != nil && response.Request.URL != nil {
		finalURL = response.Request.URL.String()
	}
	return FetchResult{Body: body, ContentType: response.Header.Get("Content-Type"), FinalURL: finalURL}, nil
}

func safeDial(ctx context.Context, dialer *net.Dialer, allow func(net.IP) bool, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", host, err)
	}
	for _, address := range resolved {
		if allow(address.IP) {
			return dialer.DialContext(ctx, network, net.JoinHostPort(address.IP.String(), port))
		}
	}
	return nil, fmt.Errorf("refusing to connect to non-public address for %s", host)
}

func isPublicIP(ip net.IP) bool {
	if ip == nil || !ip.IsGlobalUnicast() {
		return false
	}
	return !ip.IsUnspecified() && !ip.IsLoopback() && !ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() && !ip.IsMulticast() && !ip.IsInterfaceLocalMulticast() && !ip.IsPrivate()
}

func readBounded(body io.Reader, limit int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	if err != nil {
		return nil, fmt.Errorf("read source: %w", err)
	}
	if int64(len(data)) > limit {
		return nil, errors.New("source response exceeds limit")
	}
	return data, nil
}

func isPublicHTTPURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != ""
}
