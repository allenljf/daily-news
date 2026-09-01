package platform

import (
	"context"
	"testing"

	"github.com/google/uuid"
	cloudrun "google.golang.org/api/run/v2"
)

func TestCloudRunLauncherPassesRunIDAsJobEnvironmentOverride(t *testing.T) {
	var gotName string
	var gotRequest *cloudrun.GoogleCloudRunV2RunJobRequest
	launcher := newCloudRunLauncher("project", "asia-east1", "daily-news-job", func(_ context.Context, name string, request *cloudrun.GoogleCloudRunV2RunJobRequest) error {
		gotName, gotRequest = name, request
		return nil
	})
	runID := uuid.MustParse("d6aa37d4-10e3-4f77-94c7-aedfe8ae29fc")

	if err := launcher.Launch(context.Background(), runID); err != nil {
		t.Fatal(err)
	}
	if gotName != "projects/project/locations/asia-east1/jobs/daily-news-job" {
		t.Fatalf("job name = %q", gotName)
	}
	override := gotRequest.Overrides.ContainerOverrides[0]
	if len(override.Env) != 1 || override.Env[0].Name != "RUN_ID" || override.Env[0].Value != runID.String() {
		t.Fatalf("override = %#v", override)
	}
}
