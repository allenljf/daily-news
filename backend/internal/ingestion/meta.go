package ingestion

import (
	"context"
	"errors"
	"fmt"
)

// ErrMetaAdapterUnavailable documents the restricted Meta adapter boundary.
var ErrMetaAdapterUnavailable = errors.New("Meta search requires an authorized integration and is not enabled")

// MetaAdapter is the future boundary for authorized Facebook Page, Instagram
// Professional/hashtag, or Threads content only. It never performs arbitrary
// full-text search and never falls back to the general web search adapter.
type MetaAdapter struct{ platform string }

func NewMetaAdapter(platform string) MetaAdapter { return MetaAdapter{platform: platform} }

func (adapter MetaAdapter) Search(context.Context, SourceWork) (SourceSearchResult, error) {
	return SourceSearchResult{}, fmt.Errorf("%w: %s", ErrMetaAdapterUnavailable, adapter.platform)
}
