package platform

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	cloudrun "google.golang.org/api/run/v2"
)

// CloudRunLauncher triggers an existing regional Cloud Run Job using ADC.
type CloudRunLauncher struct {
	jobName string
	runJob  func(context.Context, string, *cloudrun.GoogleCloudRunV2RunJobRequest) error
}

func NewCloudRunLauncher(ctx context.Context, projectID, region, jobName string) (*CloudRunLauncher, error) {
	service, err := cloudrun.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("initialize Cloud Run client: %w", err)
	}
	return newCloudRunLauncher(projectID, region, jobName, func(ctx context.Context, name string, request *cloudrun.GoogleCloudRunV2RunJobRequest) error {
		_, err := service.Projects.Locations.Jobs.Run(name, request).Context(ctx).Do()
		return err
	}), nil
}

func newCloudRunLauncher(projectID, region, jobName string, runJob func(context.Context, string, *cloudrun.GoogleCloudRunV2RunJobRequest) error) *CloudRunLauncher {
	return &CloudRunLauncher{jobName: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", projectID, region, jobName), runJob: runJob}
}

func (launcher *CloudRunLauncher) Launch(ctx context.Context, runID uuid.UUID) error {
	request := &cloudrun.GoogleCloudRunV2RunJobRequest{Overrides: &cloudrun.GoogleCloudRunV2Overrides{ContainerOverrides: []*cloudrun.GoogleCloudRunV2ContainerOverride{{Env: []*cloudrun.GoogleCloudRunV2EnvVar{{Name: "RUN_ID", Value: runID.String()}}}}}}
	if err := launcher.runJob(ctx, launcher.jobName, request); err != nil {
		return fmt.Errorf("run Cloud Run Job: %w", err)
	}
	return nil
}
