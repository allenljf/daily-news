package ingestion

// blockedSourceDomains lists platforms that must never enter a news feed.
// YouTube results cannot be verified against the Category intent, so every
// adapter's YouTube candidates are dropped before persistence.
var blockedSourceDomains = []string{"youtube.com", "youtu.be"}

// isBlockedSourceURL reports whether a candidate belongs to a platform the
// ingestion pipeline no longer accepts, regardless of which adapter produced it.
func isBlockedSourceURL(raw string) bool {
	host := sourceHost(raw)
	if host == "" {
		return false
	}
	for _, domain := range blockedSourceDomains {
		if matchesHost(host, domain) {
			return true
		}
	}
	return false
}
