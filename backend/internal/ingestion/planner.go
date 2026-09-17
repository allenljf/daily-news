package ingestion

import (
	"context"
	"database/sql"
)

// WorkPlanner binds each active, explicitly configured Source Setting to one adapter.
type WorkPlanner struct {
	db      *sql.DB
	adapter SourceAdapter
}

func NewWorkPlanner(db *sql.DB, adapter SourceAdapter) *WorkPlanner {
	return &WorkPlanner{db: db, adapter: adapter}
}
func (planner *WorkPlanner) Load(ctx context.Context) ([]SourceWork, error) {
	rows, err := planner.db.QueryContext(ctx, `SELECT c.id,ss.id,c.name,c.content_language,c.search_keywords,c.special_requirements,ss.label,ss.website_input FROM categories c JOIN source_settings ss ON ss.category_id=c.id WHERE c.deleted_at IS NULL AND ss.deleted_at IS NULL AND (btrim(ss.website_input) ~* '^https?://' OR btrim(ss.website_input) = '') ORDER BY c.created_at,ss.position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var work []SourceWork
	for rows.Next() {
		var item SourceWork
		if err := rows.Scan(&item.CategoryID, &item.SourceID, &item.CategoryName, &item.ContentLanguage, &item.SearchKeywords, &item.SpecialRequirements, &item.SourceLabel, &item.WebsiteInput); err != nil {
			return nil, err
		}
		item.Adapter = planner.adapter
		work = append(work, item)
	}
	return work, rows.Err()
}
