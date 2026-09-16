package main

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRunRejectsMissingRunID(t *testing.T) {
	t.Parallel()

	if got, want := run(context.Background(), "", io.Discard), 2; got != want {
		t.Fatalf("run() = %d, want %d", got, want)
	}
}

func TestJobCompositionPlansWorkBeforeRunningOrchestrator(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "NewWorkPlanner") || strings.Contains(string(source), "Run(ctx, id, nil)") {
		t.Fatalf("job composition does not pass planned work to the orchestrator")
	}
}

func TestJobCompositionUsesCompositeSourceAdapter(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	if !strings.Contains(text, "NewHostRouter") || !strings.Contains(text, "NewWebSourceAdapter") {
		t.Fatalf("job composition does not build the composite web source adapter")
	}
	if strings.Contains(text, "NewRSSAdapter(http.DefaultClient)") {
		t.Fatalf("job composition still treats every HTTP(S) URL as a direct RSS feed")
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
