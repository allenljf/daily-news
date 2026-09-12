// Command daily-news-job is the Cloud Run Job entrypoint for one Ingestion Run.
package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/allenljf/daily-news/backend/internal/ingestion"
	"github.com/allenljf/daily-news/backend/internal/platform"
	"github.com/google/uuid"
)

const (
	exitFailure      = 1
	exitInvalidInput = 2
)

func main() {
	context, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(run(context, os.Getenv("RUN_ID"), os.Stderr))
}

func run(ctx context.Context, runID string, stderr io.Writer) int {
	if strings.TrimSpace(runID) == "" {
		fmt.Fprintln(stderr, "RUN_ID is required")
		return exitInvalidInput
	}
	id, err := uuid.Parse(runID)
	if err != nil {
		fmt.Fprintln(stderr, "RUN_ID must be a UUID")
		return exitInvalidInput
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if strings.TrimSpace(databaseURL) == "" {
		fmt.Fprintln(stderr, "DATABASE_URL is required")
		return exitFailure
	}
	if err := platform.ApplyMigrations("file:///app/migrations", databaseURL); err != nil {
		fmt.Fprintln(stderr, "apply migrations:", err)
		return exitFailure
	}
	database, err := platform.OpenDB(ctx, databaseURL)
	if err != nil {
		fmt.Fprintln(stderr, "open database:", err)
		return exitFailure
	}
	defer database.Close()
	id, err = ingestion.NewRunStore(database).EnsureJobRun(ctx, id)
	if err != nil {
		fmt.Fprintln(stderr, "acquire ingestion run:", err)
		return exitFailure
	}
	work, err := ingestion.NewWorkPlanner(database, ingestion.NewRSSAdapter(http.DefaultClient)).Load(ctx)
	if err != nil {
		fmt.Fprintln(stderr, "plan ingestion work:", err)
		return exitFailure
	}
	if _, err = ingestion.NewOrchestrator(database).Run(ctx, id, work); err != nil {
		fmt.Fprintln(stderr, "run ingestion:", err)
		return exitFailure
	}

	return 0
}
