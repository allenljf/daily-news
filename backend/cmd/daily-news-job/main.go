// Command daily-news-job is the Cloud Run Job entrypoint for one Ingestion Run.
package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

func run(_ context.Context, runID string, stderr io.Writer) int {
	if strings.TrimSpace(runID) == "" {
		fmt.Fprintln(stderr, "RUN_ID is required")
		return exitInvalidInput
	}

	fmt.Fprintln(stderr, "ingestion orchestrator is not installed")
	return exitFailure
}
