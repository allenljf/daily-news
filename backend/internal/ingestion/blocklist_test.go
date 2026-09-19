package ingestion

import "testing"

func TestBlockedSourceURLRejectsOnlyBlockedPlatformHosts(t *testing.T) {
	blocked := []string{
		"https://www.youtube.com/watch?v=abc",
		"https://youtube.com/watch?v=abc",
		"https://youtu.be/abc",
		"https://m.youtube.com/watch?v=abc",
	}
	for _, raw := range blocked {
		if !isBlockedSourceURL(raw) {
			t.Fatalf("isBlockedSourceURL(%q) = false, want true", raw)
		}
	}

	allowed := []string{
		"",
		"not-a-url",
		"https://news.example/youtube.com-story",
		"https://news.example/watch?v=abc",
		"https://youtube.example.com/watch",
	}
	for _, raw := range allowed {
		if isBlockedSourceURL(raw) {
			t.Fatalf("isBlockedSourceURL(%q) = true, want false", raw)
		}
	}
}

func TestCappedDropsBlockedPlatformCandidatesButKeepsThemConsidered(t *testing.T) {
	result := SourceSearchResult{
		CandidateCount: 3,
		Candidates: []CandidateArticle{
			{Title: "Video", CanonicalURL: "https://www.youtube.com/watch?v=abc"},
			{Title: "Article", CanonicalURL: "https://news.example/a"},
			{Title: "Short", CanonicalURL: "https://youtu.be/abc"},
		},
	}.capped()

	if result.CandidateCount != 3 {
		t.Fatalf("CandidateCount = %d, want 3 considered candidates", result.CandidateCount)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].CanonicalURL != "https://news.example/a" {
		t.Fatalf("candidates = %#v, want only the non-YouTube Article", result.Candidates)
	}
}
