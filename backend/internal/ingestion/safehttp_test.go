package ingestion

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func allowAllIPs(net.IP) bool { return true }

func TestSafeFetcherRejectsNonPublicTargets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		t.Fatal("a non-public target must never be dialed")
	}))
	defer server.Close()

	targets := []string{
		server.URL,
		"http://127.0.0.1/",
		"http://10.0.0.1/",
		"http://192.168.1.1/",
		"http://169.254.169.254/latest/meta-data/",
		"http://[::1]/",
	}
	for _, target := range targets {
		if _, err := NewSafeFetcher().Fetch(context.Background(), target); err == nil {
			t.Fatalf("Fetch(%q) succeeded for a non-public target", target)
		}
	}
}

func TestFetcherRejectsOversizedResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(strings.Repeat("a", 200)))
	}))
	defer server.Close()

	fetcher := newFetcher(allowAllIPs)
	fetcher.maxBytes = 64
	if _, err := fetcher.Fetch(context.Background(), server.URL); err == nil {
		t.Fatal("Fetch succeeded for an oversized response")
	}
}

func TestFetcherCapsRedirects(t *testing.T) {
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, serverURL+"/next", http.StatusFound)
	}))
	serverURL = server.URL
	defer server.Close()

	if _, err := newFetcher(allowAllIPs).Fetch(context.Background(), server.URL); err == nil {
		t.Fatal("Fetch followed an unbounded redirect loop")
	}
}
