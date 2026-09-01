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

func TestRunRejectsMalformedRunID(t *testing.T) {
	t.Parallel()

	if got, want := run(context.Background(), "not-a-uuid", io.Discard), 2; got != want {
		t.Fatalf("run() = %d, want %d", got, want)
	}
}

func TestRunDoesNotReportSuccessBeforeOrchestratorExists(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if got, want := run(context.Background(), "d6aa37d4-10e3-4f77-94c7-aedfe8ae29fc", io.Discard), 1; got != want {
		t.Fatalf("run() = %d, want %d", got, want)
	}
}
