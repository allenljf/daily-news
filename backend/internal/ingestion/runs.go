// Package ingestion implements Ingestion Run state and launch coordination.
package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/allenljf/daily-news/backend/internal/httpapi"
	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/google/uuid"
)

const (
	TriggerManual    = "manual"
	TriggerScheduled = "scheduled"
	StatusQueued     = "queued"
	activeRunLockID  = int64(824706321)
)

// RunResponse is the public representation of an asynchronous Ingestion Run.
type RunResponse struct {
	ID         uuid.UUID  `json:"id"`
	Trigger    string     `json:"trigger"`
	Status     string     `json:"status"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
}

// LatestResponse is the homepage refresh metadata and active background Run.
type LatestResponse struct {
	LastSuccessfulAt *time.Time   `json:"last_successful_at"`
	ActiveRun        *RunResponse `json:"active_run"`
}

// JobLauncher triggers the Cloud Run Job after a Run is committed.
type JobLauncher interface {
	Launch(context.Context, uuid.UUID) error
}

// NoopJobLauncher preserves local development behavior until a Cloud Run
// launcher adapter is configured at deployment composition.
type NoopJobLauncher struct{}

func (NoopJobLauncher) Launch(context.Context, uuid.UUID) error { return nil }

// RunStore persists Ingestion Runs through explicit SQL transaction boundaries.
type RunStore struct {
	db  *sql.DB
	now func() time.Time
}

func NewRunStore(db *sql.DB) *RunStore {
	return &RunStore{db: db, now: time.Now}
}

// Latest returns the active Run and timestamp of the latest successful Run.
func (store *RunStore) Latest(ctx context.Context) (LatestResponse, error) {
	var result LatestResponse
	if err := store.db.QueryRowContext(ctx, `SELECT finished_at FROM ingestion_runs WHERE status='succeeded' ORDER BY finished_at DESC LIMIT 1`).Scan(&result.LastSuccessfulAt); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return LatestResponse{}, err
	}
	active, err := findActive(ctx, store.db)
	if err != nil {
		return LatestResponse{}, err
	}
	result.ActiveRun = active
	return result, nil
}

// AcquireManual atomically reuses an active Run or creates a new queued manual Run.
func (store *RunStore) AcquireManual(ctx context.Context) (RunResponse, bool, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return RunResponse{}, false, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, activeRunLockID); err != nil {
		return RunResponse{}, false, err
	}
	active, err := findActive(ctx, tx)
	if err != nil {
		return RunResponse{}, false, err
	}
	if active != nil {
		if err = tx.Commit(); err != nil {
			return RunResponse{}, false, err
		}
		return *active, false, nil
	}

	result := RunResponse{ID: uuid.New(), Trigger: TriggerManual, Status: StatusQueued}
	taipeiDate := store.now().In(time.FixedZone("Asia/Taipei", 8*60*60)).Format("2006-01-02")
	err = tx.QueryRowContext(ctx, `INSERT INTO ingestion_runs (id,trigger,idempotency_key,taipei_date,status) VALUES ($1,$2,$3,$4,$5) RETURNING started_at,finished_at`, result.ID, result.Trigger, "manual:"+result.ID.String(), taipeiDate, result.Status).Scan(&result.StartedAt, &result.FinishedAt)
	if err != nil {
		return RunResponse{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return RunResponse{}, false, err
	}
	return result, true, nil
}

// AcquireScheduled atomically reuses today's Taipei scheduled Run or creates one.
func (store *RunStore) AcquireScheduled(ctx context.Context, requestedID uuid.UUID) (RunResponse, bool, error) {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return RunResponse{}, false, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, activeRunLockID); err != nil {
		return RunResponse{}, false, err
	}
	taipeiDate := store.now().In(time.FixedZone("Asia/Taipei", 8*60*60)).Format("2006-01-02")
	var existing RunResponse
	err = tx.QueryRowContext(ctx, `SELECT id,trigger,status,started_at,finished_at FROM ingestion_runs WHERE idempotency_key=$1`, "scheduled:"+taipeiDate).Scan(&existing.ID, &existing.Trigger, &existing.Status, &existing.StartedAt, &existing.FinishedAt)
	if err == nil {
		if err = tx.Commit(); err != nil {
			return RunResponse{}, false, err
		}
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return RunResponse{}, false, err
	}
	result := RunResponse{ID: requestedID, Trigger: TriggerScheduled, Status: StatusQueued}
	err = tx.QueryRowContext(ctx, `INSERT INTO ingestion_runs (id,trigger,idempotency_key,taipei_date,status) VALUES ($1,$2,$3,$4,$5) RETURNING started_at,finished_at`, result.ID, result.Trigger, "scheduled:"+taipeiDate, taipeiDate, result.Status).Scan(&result.StartedAt, &result.FinishedAt)
	if err != nil {
		return RunResponse{}, false, err
	}
	if err = tx.Commit(); err != nil {
		return RunResponse{}, false, err
	}
	return result, true, nil
}

// EnsureJobRun returns an existing API-created Run, or creates today's scheduled Run.
func (store *RunStore) EnsureJobRun(ctx context.Context, requestedID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := store.db.QueryRowContext(ctx, `SELECT id FROM ingestion_runs WHERE id=$1`, requestedID).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, err
	}
	run, _, err := store.AcquireScheduled(ctx, requestedID)
	return run.ID, err
}

type rowQueryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func findActive(ctx context.Context, source rowQueryer) (*RunResponse, error) {
	var result RunResponse
	err := source.QueryRowContext(ctx, `SELECT id,trigger,status,started_at,finished_at FROM ingestion_runs WHERE status IN ('queued','running') ORDER BY started_at DESC LIMIT 1`).Scan(&result.ID, &result.Trigger, &result.Status, &result.StartedAt, &result.FinishedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// RunService coordinates database acquisition with launch after commit.
type RunService struct {
	store    *RunStore
	launcher JobLauncher
}

func NewRunService(store *RunStore, launcher JobLauncher) *RunService {
	return &RunService{store: store, launcher: launcher}
}

func (service *RunService) Latest(ctx context.Context) (LatestResponse, error) {
	return service.store.Latest(ctx)
}

func (service *RunService) RequestManual(ctx context.Context) (RunResponse, error) {
	run, created, err := service.store.AcquireManual(ctx)
	if err != nil || !created {
		return run, err
	}
	if err = service.launcher.Launch(ctx, run.ID); err != nil {
		return RunResponse{}, err
	}
	return run, nil
}

// NewHandler exposes the authenticated Ingestion Run HTTP contract.
func NewHandler(store *RunStore, launcher JobLauncher, verifier identity.TokenVerifier, allowedEmail string) *http.ServeMux {
	mux := http.NewServeMux()
	service := NewRunService(store, launcher)
	require := func(next http.Handler) http.Handler {
		return httpapi.RequireAllowed(verifier, allowedEmail, next)
	}
	mux.Handle("GET /v1/ingestion-runs/latest", require(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		result, err := service.Latest(request.Context())
		if err != nil {
			httpapi.WriteProblem(writer, httpapi.Problem{Title: "Database unavailable", Status: http.StatusServiceUnavailable})
			return
		}
		writeJSON(writer, http.StatusOK, result)
	})))
	mux.Handle("POST /v1/ingestion-runs", require(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		result, err := service.RequestManual(request.Context())
		if err != nil {
			httpapi.WriteProblem(writer, httpapi.Problem{Title: "Ingestion launch unavailable", Status: http.StatusServiceUnavailable})
			return
		}
		writeJSON(writer, http.StatusAccepted, result)
	})))
	return mux
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
