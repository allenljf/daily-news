package ingestion

import (
	"context"
	"errors"
	"testing"
)

type recordingAdapter struct {
	name   string
	called *string
}

func (adapter recordingAdapter) Search(context.Context, SourceWork) (SourceSearchResult, error) {
	if adapter.called != nil {
		*adapter.called = adapter.name
	}
	return SourceSearchResult{}, nil
}

func TestHostRouterRoutesPlatformHostsToOfficialAdapters(t *testing.T) {
	called := ""
	router := NewHostRouter(
		recordingAdapter{name: "web", called: &called},
		recordingAdapter{name: "youtube", called: &called},
		recordingAdapter{name: "facebook", called: &called},
		recordingAdapter{name: "instagram", called: &called},
		recordingAdapter{name: "threads", called: &called},
	)

	cases := map[string]string{
		"https://www.youtube.com/@channel":   "youtube",
		"https://youtu.be/abc":               "youtube",
		"https://www.facebook.com/some-page": "facebook",
		"https://www.instagram.com/someone":  "instagram",
		"https://www.threads.net/@user":      "threads",
		"https://threads.com/@user":          "threads",
		"https://news.example/home":          "web",
		"not-a-url":                          "web",
	}
	for website, want := range cases {
		called = ""
		if _, err := router.Search(context.Background(), SourceWork{WebsiteInput: website}); err != nil {
			t.Fatalf("Search(%q): %v", website, err)
		}
		if called != want {
			t.Fatalf("Search(%q) routed to %q, want %q", website, called, want)
		}
	}
}

func TestMetaAdaptersRejectArbitraryKeywordSearch(t *testing.T) {
	for _, platform := range []string{"Facebook", "Instagram", "Threads"} {
		if _, err := NewMetaAdapter(platform).Search(context.Background(), SourceWork{WebsiteInput: "https://example.invalid"}); !errors.Is(err, ErrMetaAdapterUnavailable) {
			t.Fatalf("%s err = %v, want ErrMetaAdapterUnavailable", platform, err)
		}
	}
}
