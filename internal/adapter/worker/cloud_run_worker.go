package worker

import (
	"context"
	"fmt"
	"log/slog"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/pkg/config"
)

const (
	Size128MB = 128 * 1024 * 1024
	Size1GB   = 1 * 1024 * 1024 * 1024
)

type CloudRunWorker struct {
	client *run.JobsClient
	config config.WorkerConfig
	logger *slog.Logger
}

func NewCloudRunWorker(ctx context.Context, cfg config.WorkerConfig, logger *slog.Logger) (*CloudRunWorker, error) {
	client, err := run.NewJobsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run Jobs client: %w", err)
	}

	return &CloudRunWorker{
		client: client,
		config: cfg,
		logger: logger,
	}, nil
}

func (w *CloudRunWorker) ProcessImage(ctx context.Context, input *port.ProcessingInput) error {
	jobName := w.determineJobName(input.Size)

	w.logger.Info("Selected Cloud Run Job based on size",
		slog.String("image_id", input.ImageID),
		slog.Int64("size_bytes", input.Size),
		slog.String("job_name", jobName),
	)

	req := &runpb.RunJobRequest{
		Name: jobName,
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
						{
							Name:   "WORKER_TYPE_OVERRIDE",
							Values: &runpb.EnvVar_Value{Value: w.getWorkerTypeLabel(input.Size)},
						},
					},
				},
			},
		},
	}

	op, err := w.client.RunJob(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to run job (%s): %w", jobName, err)
	}

	w.logger.Info("Cloud Run job triggered successfully",
		slog.String("operation", op.Name()),
		slog.String("target_job", jobName))

	return nil
}

func (w *CloudRunWorker) determineJobName(size int64) string {
	if size <= 0 {
		w.logger.Warn("File size is 0 or unknown, defaulting to SMALL worker")
		return w.config.JobSmall
	}

	switch {
	case size < Size128MB:
		return w.config.JobSmall
	case size < Size1GB:
		return w.config.JobMedium
	default:

		return w.config.JobLarge
	}
}

func (w *CloudRunWorker) getWorkerTypeLabel(size int64) string {
	if size < Size128MB {
		return "small"
	}
	if size < Size1GB {
		return "medium"
	}
	return "large"
}
