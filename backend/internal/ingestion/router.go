package ingestion

import (
	"context"
	"errors"
	"strings"
)

// HostRouter dispatches each Source Setting to its platform adapter or the
// composite web adapter. Meta hosts never reach the general web adapter.
type HostRouter struct {
	web       SourceAdapter
	youtube   SourceAdapter
	threads   SourceAdapter
	facebook  SourceAdapter
	instagram SourceAdapter
}

func NewHostRouter(web, youtube, facebook, instagram, threads SourceAdapter) *HostRouter {
	return &HostRouter{web: web, youtube: youtube, facebook: facebook, instagram: instagram, threads: threads}
}

func (router *HostRouter) Search(ctx context.Context, work SourceWork) (SourceSearchResult, error) {
	host := sourceHost(work.WebsiteInput)
	switch {
	case matchesHost(host, "youtube.com") || host == "youtu.be":
		return route(ctx, router.youtube, work)
	case matchesHost(host, "threads.net") || matchesHost(host, "threads.com"):
		return route(ctx, router.threads, work)
	case matchesHost(host, "facebook.com"):
		return route(ctx, router.facebook, work)
	case matchesHost(host, "instagram.com"):
		return route(ctx, router.instagram, work)
	default:
		return route(ctx, router.web, work)
	}
}

func route(ctx context.Context, adapter SourceAdapter, work SourceWork) (SourceSearchResult, error) {
	if adapter == nil {
		return SourceSearchResult{}, errors.New("source adapter is not configured")
	}
	return adapter.Search(ctx, work)
}

func matchesHost(host, domain string) bool {
	if host == "" {
		return false
	}
	return host == domain || strings.HasSuffix(host, "."+domain)
}
