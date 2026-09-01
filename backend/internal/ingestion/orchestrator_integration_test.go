package ingestion

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestOrchestratorPersistsDedupeExpiryAttemptsAndTerminalRun(t *testing.T) {
	database := ingestionIntegrationDatabase(t)
	defer database.Close()

	runID, firstCategoryID, secondCategoryID, firstSourceID, secondSourceID, failedSourceID := seedIngestionWork(t, database)
	seedSuppressedArticle(t, database)
	firstCandidates := make([]CandidateArticle, 0, 12)
	firstCandidates = append(firstCandidates,
		CandidateArticle{Title: "Original", CanonicalURL: "https://example.test/original"},
		CandidateArticle{Title: "Different title", CanonicalURL: "https://example.test/original"},
		CandidateArticle{Title: "Original", CanonicalURL: "https://elsewhere.test/original"},
	)
	for index := 4; index <= 12; index++ {
		firstCandidates = append(firstCandidates, CandidateArticle{Title: fmt.Sprintf("Article %d", index), CanonicalURL: fmt.Sprintf("https://example.test/%d", index)})
	}
	work := []SourceWork{
		{CategoryID: firstCategoryID, SourceID: firstSourceID, Adapter: fakeAdapter{candidates: firstCandidates}},
		{CategoryID: secondCategoryID, SourceID: secondSourceID, Adapter: fakeAdapter{candidates: []CandidateArticle{{Title: "Original", CanonicalURL: "https://example.test/original"}, {Title: "Suppressed", CanonicalURL: "https://example.test/suppressed"}}}},
		{CategoryID: firstCategoryID, SourceID: failedSourceID, Adapter: fakeAdapter{err: errors.New("source unavailable")}},
	}

	result, err := NewOrchestrator(database).Run(context.Background(), runID, work)
	if err != nil {
		t.Fatal(err)
	}
	if result != (Result{CandidateCount: 12, InsertedCount: 8, DuplicateCount: 4, ErrorCount: 1}) {
		t.Fatalf("result = %#v", result)
	}

	var status string
	var candidates, inserted, duplicates, failures int
	if err := database.QueryRow(`SELECT status,candidate_count,inserted_count,duplicate_count,error_count FROM ingestion_runs WHERE id=$1`, runID).Scan(&status, &candidates, &inserted, &duplicates, &failures); err != nil {
		t.Fatal(err)
	}
	if status != "succeeded" || candidates != 12 || inserted != 8 || duplicates != 4 || failures != 1 {
		t.Fatalf("terminal run = %q %d %d %d %d", status, candidates, inserted, duplicates, failures)
	}
	var attempts int
	if err := database.QueryRow(`SELECT count(*) FROM ingestion_attempts WHERE run_id=$1`, runID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if attempts != 3 {
		t.Fatalf("attempt count = %d", attempts)
	}
	var linkedCategories int
	if err := database.QueryRow(`SELECT count(*) FROM category_articles ca JOIN articles a ON a.id=ca.article_id WHERE a.canonical_url='https://example.test/original'`).Scan(&linkedCategories); err != nil {
		t.Fatal(err)
	}
	if linkedCategories != 2 {
		t.Fatalf("shared Article links = %d", linkedCategories)
	}
	var suppressedLinks int
	if err := database.QueryRow(`SELECT count(*) FROM category_articles ca JOIN articles a ON a.id=ca.article_id WHERE a.canonical_url='https://example.test/suppressed'`).Scan(&suppressedLinks); err != nil {
		t.Fatal(err)
	}
	if suppressedLinks != 0 {
		t.Fatalf("suppressed Article links = %d", suppressedLinks)
	}
	var expiresAt time.Time
	if err := database.QueryRow(`SELECT expires_at FROM articles WHERE canonical_url='https://example.test/original'`).Scan(&expiresAt); err != nil {
		t.Fatal(err)
	}
	if delta := time.Until(expiresAt); delta < 29*24*time.Hour || delta > 31*24*time.Hour {
		t.Fatalf("expires at = %s", expiresAt)
	}
}

type fakeAdapter struct {
	candidates []CandidateArticle
	err        error
}

func (adapter fakeAdapter) Search(context.Context, SourceWork) ([]CandidateArticle, error) {
	return adapter.candidates, adapter.err
}

func seedIngestionWork(t *testing.T, database *sql.DB) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	runID, firstCategoryID, secondCategoryID := uuid.New(), uuid.New(), uuid.New()
	firstSourceID, secondSourceID, failedSourceID := uuid.New(), uuid.New(), uuid.New()
	if _, err := database.Exec(`INSERT INTO ingestion_runs (id,trigger,idempotency_key,taipei_date,status) VALUES ($1,'manual',$2,'2026-09-01','queued')`, runID, "manual:"+runID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO categories (id,name) VALUES ($1,'First'),($2,'Second')`, firstCategoryID, secondCategoryID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO source_settings (id,category_id,label,website_input,kind,position) VALUES ($1,$2,'First','','unspecified',0),($3,$4,'Second','','unspecified',0),($5,$2,'Failed','','unspecified',1)`, firstSourceID, firstCategoryID, secondSourceID, secondCategoryID, failedSourceID); err != nil {
		t.Fatal(err)
	}
	return runID, firstCategoryID, secondCategoryID, firstSourceID, secondSourceID, failedSourceID
}

func seedSuppressedArticle(t *testing.T, database *sql.DB) {
	t.Helper()
	if _, err := database.Exec(`INSERT INTO articles (id,title,normalized_title_hash,canonical_url,canonical_url_hash,first_seen_at,expires_at,deleted_at) VALUES ($1,'Suppressed',$2,'https://example.test/suppressed',$3,now(),now()+interval '30 days',now())`, uuid.New(), fingerprint("suppressed"), fingerprint("https://example.test/suppressed")); err != nil {
		t.Fatal(err)
	}
}
