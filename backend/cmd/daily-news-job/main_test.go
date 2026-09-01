package main

import (
	"context"
	"io"
	"testing"
)

func TestRunRejectsMissingRunID(t *testing.T) {
	t.Parallel()

	if got, want := run(context.Background(), "", io.Discard), 2; got != want {
		t.Fatalf("run() = %d, want %d", got, want)
	}
}

func TestRunDoesNotReportSuccessBeforeOrchestratorExists(t *testing.T) {
	t.Parallel()

	if got, want := run(context.Background(), "run-123", io.Discard), 1; got != want {
		t.Fatalf("run() = %d, want %d", got, want)
	}
}
