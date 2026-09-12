package ingestion

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/allenljf/daily-news/backend/internal/identity"
	"github.com/allenljf/daily-news/backend/internal/platform"
	"github.com/google/uuid"
)

func TestRequestManualMergesConcurrentActiveRun(t *testing.T) {
	database := ingestionIntegrationDatabase(t)
	defer database.Close()

	service := NewRunService(NewRunStore(database), &recordingLauncher{})
	var responses [2]RunResponse
	var errs [2]error
	var group sync.WaitGroup
	group.Add(2)
	for index := range responses {
		go func(index int) {
			defer group.Done()
			responses[index], errs[index] = service.RequestManual(context.Background())
		}(index)
	}
	group.Wait()
	if errs[0] != nil || errs[1] != nil {
		t.Fatalf("RequestManual() errors = %v, %v", errs[0], errs[1])
	}
	if responses[0].ID != responses[1].ID || responses[0].Status != StatusQueued || responses[1].Trigger != TriggerManual {
		t.Fatalf("concurrent responses = %#v, %#v", responses[0], responses[1])
	}

	var count int
	if err := database.QueryRow(`SELECT count(*) FROM ingestion_runs WHERE status IN ('queued', 'running')`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("active runs = %d, want 1", count)
	}
}

func TestRunServiceLatestAndCompletedScheduledRun(t *testing.T) {
	database := ingestionIntegrationDatabase(t)
	defer database.Close()

	finishedAt := time.Date(2026, 8, 30, 0, 1, 0, 0, time.UTC)
	if _, err := database.Exec(`INSERT INTO ingestion_runs (id,trigger,idempotency_key,taipei_date,started_at,finished_at,status) VALUES ($1,'scheduled','scheduled:2026-08-30','2026-08-30',$2,$3,'succeeded')`, uuid.New(), finishedAt.Add(-time.Minute), finishedAt); err != nil {
		t.Fatal(err)
	}
	launcher := &recordingLauncher{}
	service := NewRunService(NewRunStore(database), launcher)

	latest, err := service.Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if latest.LastSuccessfulAt == nil || !latest.LastSuccessfulAt.Equal(finishedAt) || latest.ActiveRun != nil {
		t.Fatalf("Latest() = %#v", latest)
	}
	manual, err := service.RequestManual(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if manual.Trigger != TriggerManual || manual.Status != StatusQueued || manual.FinishedAt != nil {
		t.Fatalf("manual run = %#v", manual)
	}
	if len(launcher.RunIDs) != 1 || launcher.RunIDs[0] != manual.ID {
		t.Fatalf("launches = %#v", launcher.RunIDs)
	}
}

func TestAcquireScheduledUsesTaipeiDayIdempotency(t *testing.T) {
	database := ingestionIntegrationDatabase(t)
	defer database.Close()
	store := NewRunStore(database)
	first, created, err := store.AcquireScheduled(context.Background(), uuid.New())
	if err != nil || !created {
		t.Fatalf("first scheduled run = %#v, %t, %v", first, created, err)
	}
	second, created, err := store.AcquireScheduled(context.Background(), uuid.New())
	if err != nil || created || second.ID != first.ID || second.Trigger != "scheduled" {
		t.Fatalf("second scheduled run = %#v, %t, %v", second, created, err)
	}
}

func TestRunHandlerPreservesAcceptedAndProblemContracts(t *testing.T) {
	database := ingestionIntegrationDatabase(t)
	defer database.Close()

	launcher := &recordingLauncher{}
	handler := NewHandler(NewRunStore(database), launcher, ingestionVerifier{}, "allowed@example.com")
	request := httptest.NewRequest(http.MethodPost, "/v1/ingestion-runs", nil)
	request.Header.Set("Authorization", "Bearer valid")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted || response.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("POST status/content type = %d/%q", response.Code, response.Header().Get("Content-Type"))
	}
	var run RunResponse
	if err := json.Unmarshal(response.Body.Bytes(), &run); err != nil {
		t.Fatal(err)
	}
	if run.ID == uuid.Nil || run.Trigger != TriggerManual || run.Status != StatusQueued || run.FinishedAt != nil {
		t.Fatalf("POST response = %#v", run)
	}

	unauthenticated := httptest.NewRequest(http.MethodGet, "/v1/ingestion-runs/latest", nil)
	problem := httptest.NewRecorder()
	handler.ServeHTTP(problem, unauthenticated)
	if problem.Code != http.StatusUnauthorized || problem.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("problem status/content type = %d/%q", problem.Code, problem.Header().Get("Content-Type"))
	}
	if problem.Body.String() != "{\"detail\":{\"title\":\"Authentication required\",\"status\":401}}\n" {
		t.Fatalf("problem body = %s", problem.Body.String())
	}
}

type recordingLauncher struct {
	mu     sync.Mutex
	RunIDs []uuid.UUID
}

type ingestionVerifier struct{}

func (ingestionVerifier) Verify(context.Context, string) (identity.Identity, error) {
	return identity.Identity{UID: "user", Email: "allowed@example.com"}, nil
}

func (launcher *recordingLauncher) Launch(_ context.Context, runID uuid.UUID) error {
	launcher.mu.Lock()
	defer launcher.mu.Unlock()
	launcher.RunIDs = append(launcher.RunIDs, runID)
	return nil
}

func ingestionIntegrationDatabase(t *testing.T) *sql.DB {
	t.Helper()
	name := "daily-news-go-ingestion-test"
	ingestionDocker(t, "run", "--rm", "--detach", "--name", name, "-e", "POSTGRES_USER=daily_news", "-e", "POSTGRES_PASSWORD=daily_news", "-e", "POSTGRES_DB=daily_news", "-p", "127.0.0.1::5432", "postgres:16-alpine")
	t.Cleanup(func() { _ = exec.Command("docker", "rm", "-f", name).Run() })
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if exec.Command("docker", "exec", name, "pg_isready", "-U", "daily_news", "-d", "daily_news").Run() == nil {
			break
		}
		time.Sleep(time.Second)
	}
	port := strings.TrimSpace(ingestionDocker(t, "port", name, "5432/tcp"))
	port = port[strings.LastIndex(port, ":")+1:]
	url := "postgres://daily_news:daily_news@127.0.0.1:" + port + "/daily_news?sslmode=disable"
	waitForIngestionPostgreSQLTCP(t, url)
	path, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatal(err)
	}
	if err = platform.ApplyMigrations("file://"+path, url); err != nil {
		t.Fatal(err)
	}
	database, err := platform.OpenDB(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	return database
}

func waitForIngestionPostgreSQLTCP(t *testing.T, databaseURL string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		connectionContext, cancel := context.WithTimeout(context.Background(), time.Second)
		database, err := platform.OpenDB(connectionContext, databaseURL)
		cancel()
		if err == nil {
			_ = database.Close()
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("PostgreSQL did not accept TCP connections within 30 seconds")
}

func ingestionDocker(t *testing.T, arguments ...string) string {
	t.Helper()
	output, err := exec.Command("docker", arguments...).CombinedOutput()
	if err != nil {
		t.Fatalf("docker %s: %v\\n%s", strings.Join(arguments, " "), err, output)
	}
	return string(output)
}
