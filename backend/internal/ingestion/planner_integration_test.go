package ingestion

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
)

func TestWorkPlannerLoadsOnlyActiveHTTPSourceSettings(t *testing.T) {
	database := ingestionIntegrationDatabase(t)
	defer database.Close()
	categoryID := uuid.New()
	if _, err := database.Exec(`INSERT INTO categories (id,name) VALUES ($1,'Active')`, categoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO source_settings (id,category_id,label,website_input,kind,position) VALUES ($1,$2,'Feed','https://feed.example/rss','website',0),($3,$2,'Unsupported','ftp://feed.example/rss','website',1),($4,$2,'Blank','','unspecified',2)`, uuid.New(), categoryID, uuid.New(), uuid.New()); err != nil {
		t.Fatal(err)
	}
	deletedCategoryID := uuid.New()
	if _, err := database.Exec(`INSERT INTO categories (id,name,deleted_at) VALUES ($1,'Deleted',now())`, deletedCategoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO source_settings (id,category_id,label,website_input,kind,position) VALUES ($1,$2,'Old','https://old.example/rss','website',0)`, uuid.New(), deletedCategoryID); err != nil {
		t.Fatal(err)
	}

	work, err := NewWorkPlanner(database, fakeAdapter{}).Load(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(work) != 1 {
		t.Fatalf("work count = %d", len(work))
	}
	if work[0].WebsiteInput != "https://feed.example/rss" {
		t.Fatalf("website input = %q", work[0].WebsiteInput)
	}
	if work[0].ContentLanguage != "zh-Hant" {
		t.Fatalf("content language = %q, want zh-Hant", work[0].ContentLanguage)
	}
}

var _ *sql.DB
