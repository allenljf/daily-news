package ingestion

import "strings"

// searchQuery composes the Category name, keywords and special requirements
// into one keyword query for a Source Setting.
func searchQuery(work SourceWork) string {
	parts := make([]string, 0, 3)
	if name := strings.TrimSpace(work.CategoryName); name != "" {
		parts = append(parts, name)
	}
	if work.SearchKeywords != nil && strings.TrimSpace(*work.SearchKeywords) != "" {
		parts = append(parts, strings.TrimSpace(*work.SearchKeywords))
	}
	if work.SpecialRequirements != nil && strings.TrimSpace(*work.SpecialRequirements) != "" {
		parts = append(parts, strings.TrimSpace(*work.SpecialRequirements))
	}
	return strings.Join(parts, " ")
}
