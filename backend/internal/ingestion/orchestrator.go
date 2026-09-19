package ingestion

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

const maxCandidatesPerSource = 10

// CandidateArticle is an adapter result before database deduplication.
type CandidateArticle struct {
	Title        string
	CanonicalURL string
	CitationURL  string
	Summary      *string
	PublishedAt  *time.Time
}

// SourceSearchResult separates source candidate accounting from eligible Articles.
// CandidateCount is capped source candidates considered before eligibility filtering.
type SourceSearchResult struct {
	Candidates     []CandidateArticle
	CandidateCount int
}

// SourceAdapter retrieves candidates for a single Source Setting.
type SourceAdapter interface {
	Search(context.Context, SourceWork) (SourceSearchResult, error)
}

// SourceWork binds one adapter to the Category and Source Setting it serves.
type SourceWork struct {
	CategoryID          uuid.UUID
	SourceID            uuid.UUID
	CategoryName        string
	ContentLanguage     string
	SearchKeywords      *string
	SpecialRequirements *string
	SourceLabel         string
	WebsiteInput        string
	Adapter             SourceAdapter
}

// Result is the final aggregate accounting for one Ingestion Run.
type Result struct {
	CandidateCount int
	InsertedCount  int
	DuplicateCount int
	ErrorCount     int
}

// Orchestrator isolates adapter failures while persisting each source atomically.
type Orchestrator struct {
	db  *sql.DB
	now func() time.Time
}

func NewOrchestrator(db *sql.DB) *Orchestrator {
	return &Orchestrator{db: db, now: time.Now}
}

func (orchestrator *Orchestrator) Run(ctx context.Context, runID uuid.UUID, work []SourceWork) (Result, error) {
	if err := orchestrator.markRunning(ctx, runID); err != nil {
		return Result{}, err
	}
	var result Result
	for _, item := range work {
		searchResult, err := item.Adapter.Search(ctx, item)
		if err != nil {
			if recordErr := orchestrator.recordFailure(ctx, runID, item, err); recordErr != nil {
				return Result{}, recordErr
			}
			result.ErrorCount++
			continue
		}
		searchResult = searchResult.capped()
		counts, err := orchestrator.recordSource(ctx, runID, item, searchResult)
		if err != nil {
			return Result{}, err
		}
		result.CandidateCount += counts.CandidateCount
		result.InsertedCount += counts.InsertedCount
		result.DuplicateCount += counts.DuplicateCount
	}
	if err := orchestrator.finish(ctx, runID); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (orchestrator *Orchestrator) markRunning(ctx context.Context, runID uuid.UUID) error {
	result, err := orchestrator.db.ExecContext(ctx, `UPDATE ingestion_runs SET status='running' WHERE id=$1 AND status='queued'`, runID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return errors.New("ingestion run is not queued")
	}
	return nil
}

func (result SourceSearchResult) capped() SourceSearchResult {
	eligible := make([]CandidateArticle, 0, len(result.Candidates))
	for _, candidate := range result.Candidates {
		if isBlockedSourceURL(candidate.CanonicalURL) {
			continue
		}
		eligible = append(eligible, candidate)
	}
	result.Candidates = eligible
	if result.CandidateCount > maxCandidatesPerSource {
		result.CandidateCount = maxCandidatesPerSource
	}
	if len(result.Candidates) > result.CandidateCount {
		result.Candidates = result.Candidates[:result.CandidateCount]
	}
	return result
}

func (orchestrator *Orchestrator) recordSource(ctx context.Context, runID uuid.UUID, item SourceWork, searchResult SourceSearchResult) (Result, error) {
	tx, err := orchestrator.db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, err
	}
	defer tx.Rollback()
	result := Result{CandidateCount: searchResult.CandidateCount}
	for _, candidate := range searchResult.Candidates {
		created, err := orchestrator.recordCandidate(ctx, tx, item, candidate)
		if err != nil {
			return Result{}, err
		}
		if created {
			result.InsertedCount++
		} else {
			result.DuplicateCount++
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO ingestion_attempts (id,run_id,category_id,source_setting_id,status,candidate_count,inserted_count,duplicate_count) VALUES ($1,$2,$3,$4,'succeeded',$5,$6,$7)`, uuid.New(), runID, item.CategoryID, item.SourceID, result.CandidateCount, result.InsertedCount, result.DuplicateCount); err != nil {
		return Result{}, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE ingestion_runs SET candidate_count=candidate_count+$2,inserted_count=inserted_count+$3,duplicate_count=duplicate_count+$4 WHERE id=$1`, runID, result.CandidateCount, result.InsertedCount, result.DuplicateCount); err != nil {
		return Result{}, err
	}
	if err = tx.Commit(); err != nil {
		return Result{}, err
	}
	return result, nil
}

func (orchestrator *Orchestrator) recordCandidate(ctx context.Context, tx *sql.Tx, item SourceWork, candidate CandidateArticle) (bool, error) {
	canonicalHash := fingerprint(candidate.CanonicalURL)
	titleHash := fingerprint(strings.ToLower(strings.TrimSpace(candidate.Title)))
	var articleID uuid.UUID
	var deletedAt *time.Time
	err := tx.QueryRowContext(ctx, `SELECT id,deleted_at FROM articles WHERE canonical_url_hash=$1`, canonicalHash).Scan(&articleID, &deletedAt)
	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `SELECT id,deleted_at FROM articles WHERE normalized_title_hash=$1`, titleHash).Scan(&articleID, &deletedAt)
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	if deletedAt != nil {
		return false, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		articleID = uuid.New()
		now := orchestrator.now().UTC()
		expiresAt := now.AddDate(0, 0, 30)
		if _, err = tx.ExecContext(ctx, `INSERT INTO articles (id,title,normalized_title_hash,canonical_url,canonical_url_hash,citation_url,summary,published_at,first_seen_at,expires_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, articleID, candidate.Title, titleHash, candidate.CanonicalURL, canonicalHash, candidate.CitationURL, candidate.Summary, candidate.PublishedAt, now, expiresAt); err != nil {
			return false, err
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO category_articles (category_id,article_id,source_setting_id) VALUES ($1,$2,$3) ON CONFLICT (category_id,article_id) DO NOTHING`, item.CategoryID, articleID, item.SourceID); err != nil {
			return false, err
		}
		return true, nil
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO category_articles (category_id,article_id,source_setting_id) VALUES ($1,$2,$3) ON CONFLICT (category_id,article_id) DO NOTHING`, item.CategoryID, articleID, item.SourceID); err != nil {
		return false, err
	}
	return false, nil
}

func (orchestrator *Orchestrator) recordFailure(ctx context.Context, runID uuid.UUID, item SourceWork, sourceErr error) error {
	tx, err := orchestrator.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO ingestion_attempts (id,run_id,category_id,source_setting_id,status,error_summary) VALUES ($1,$2,$3,$4,'failed',$5)`, uuid.New(), runID, item.CategoryID, item.SourceID, sourceErr.Error()); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE ingestion_runs SET error_count=error_count+1 WHERE id=$1`, runID); err != nil {
		return err
	}
	return tx.Commit()
}

func (orchestrator *Orchestrator) finish(ctx context.Context, runID uuid.UUID) error {
	_, err := orchestrator.db.ExecContext(ctx, `UPDATE ingestion_runs SET status='succeeded',finished_at=$2 WHERE id=$1 AND status='running'`, runID, orchestrator.now().UTC())
	return err
}

func fingerprint(value string) string {
	digest := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(digest[:])
}
