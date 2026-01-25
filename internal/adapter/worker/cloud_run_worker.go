package worker

import (
	"context"
	"fmt"
	"log/slog"

	run "cloud.google.com/go/run/apiv2"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/pkg/config"
)

const (
	Size128MB = 128 * 1024 * 1024
	Size1GB   = 1 * 1024 * 1024 * 1024
)

type CloudRunWorker struct {
	client    *run.JobsClient
	config    config.WorkerConfig
	gcpConfig config.GCPConfig
	logger    *slog.Logger
}

func NewCloudRunWorker(ctx context.Context, cfg config.WorkerConfig, gcpCfg config.GCPConfig, logger *slog.Logger) (*CloudRunWorker, error) {
	client, err := run.NewJobsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cloud Run Jobs client: %w", err)
	}

	return &CloudRunWorker{
		client:    client,
		config:    cfg,
		gcpConfig: gcpCfg,
		logger:    logger,
	}, nil
}

func (w *CloudRunWorker) ProcessImage(ctx context.Context, content model.Content, processingVersion vobj.ProcessingVersion) error {
	jobName := w.determineJobName(content.Size)

	// Determine image ID from content parent
	// Assuming content parent is the image
	imageID := content.Parent.ID

	w.logger.Info("Selected Cloud Run Job based on size",
		slog.String("image_id", imageID),
		slog.Int64("size_bytes", content.Size),
		slog.String("job_name", jobName),
		slog.String("version", processingVersion.String()),
	)

	req := &runpb.RunJobRequest{
		Name: jobName,
		Overrides: &runpb.RunJobRequest_Overrides{
			ContainerOverrides: []*runpb.RunJobRequest_Overrides_ContainerOverride{
				{
					Env: []*runpb.EnvVar{
						{
							Name:   "INPUT_IMAGE_ID",
							Values: &runpb.EnvVar_Value{Value: imageID},
						},
						{
							Name:   "INPUT_ORIGIN_PATH",
							Values: &runpb.EnvVar_Value{Value: content.Path},
						},
						{
							Name:   "INPUT_BUCKET_NAME",
							Values: &runpb.EnvVar_Value{Value: w.gcpConfig.OriginalBucketName},
						},
						{
							Name:   "INPUT_PROCESSING_VERSION",
							Values: &runpb.EnvVar_Value{Value: processingVersion.String()},
						},
						{
							Name:   "WORKER_TYPE_OVERRIDE",
							Values: &runpb.EnvVar_Value{Value: w.getWorkerTypeLabel(content.Size)},
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
