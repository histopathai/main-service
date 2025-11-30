package worker

import (
	"context"
	"fmt"
	"log/slog"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/pkg/config"
)

// CloudRunWorker implements ImageProcessingWorker using Cloud Run Jobs
type CloudRunWorker struct {
	client  *run.JobsClient
	jobName string
	logger  *slog.Logger
}

func NewCloudRunWorker(ctx context.Context, cfg config.WorkerConfig, logger *slog.Logger) (*CloudRunWorker, error) {
	client, err := run.NewJobsClient(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run Jobs client: %w", err)
	}

	return &CloudRunWorker{
		client:  client,
		jobName: cfg.RunJobName,
		logger:  logger,
	}, nil

}

func (a *CloudRunWorker) ProcessImage(ctx context.Context, input *port.ProcessingInput) error {

	req := &runpb.RunJobRequest{
		Name: a.jobName,
		Overrides: &runpb.RunJobRequest_Overrides{
			ContainerOverrides: []*runpb.RunJobRequest_Overrides_ContainerOverride{
				{
					Env: []*runpb.EnvVar{
						{
							Name:   "INPUT_IMAGE_ID",
							Values: &runpb.EnvVar_Value{Value: input.ImageID},
						},
						{
							Name:   "INPUT_ORIGIN_PATH",
							Values: &runpb.EnvVar_Value{Value: input.OriginPath},
						},
						{
							Name:   "INPUT_BUCKET_NAME",
							Values: &runpb.EnvVar_Value{Value: input.BucketName},
						},
					},
				},
			},
		},
	}

	op, err := a.client.RunJob(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to run job: %w", err)
	}
	logger.Info(ctx, a.logger, "Cloud Run job started", slog.String("operation", op.Name()))

	return nil
}
