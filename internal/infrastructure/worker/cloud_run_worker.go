package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	event_handler "github.com/histopathai/main-service/internal/application/event_handlers"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
)

// CloudRunWorker implements ImageProcessingWorker using Cloud Run Jobs
type CloudRunWorker struct {
	jobURL     string
	httpClient *http.Client
	logger     *slog.Logger
	maxRetries int
	retryDelay time.Duration
}

// NewCloudRunWorker creates a new Cloud Run worker
func NewCloudRunWorker(
	jobURL string,
	logger *slog.Logger,
	timeout time.Duration,
) *CloudRunWorker {
	return &CloudRunWorker{
		jobURL: jobURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:     logger,
		maxRetries: 3,
		retryDelay: 2 * time.Second,
	}
}

// ProcessImage triggers a Cloud Run job for image processing
func (w *CloudRunWorker) ProcessImage(ctx context.Context, input *port.ProcessingInput) error {
	payload := map[string]interface{}{
		"image_id":    input.ImageID,
		"origin_path": input.OriginPath,
		"bucket_name": input.BucketName,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return &event_handler.EventError{
			Err:       fmt.Errorf("failed to marshal payload: %w", err),
			Retryable: false,
			Category:  events.CategorySerialization,
			Severity:  events.SeverityHigh,
		}
	}

	// Retry logic for triggering the job
	var lastErr error
	for attempt := 0; attempt < w.maxRetries; attempt++ {
		if attempt > 0 {
			w.logger.Info("Retrying Cloud Run job trigger",
				slog.Int("attempt", attempt+1),
				slog.String("image_id", input.ImageID))

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(w.retryDelay * time.Duration(attempt)):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", w.jobURL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = err
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := w.httpClient.Do(req)
		if err != nil {
			w.logger.Error("Failed to trigger Cloud Run job",
				slog.String("error", err.Error()),
				slog.String("image_id", input.ImageID),
				slog.Int("attempt", attempt+1))
			lastErr = err
			continue
		}

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Check response status
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			w.logger.Info("Successfully triggered Cloud Run job",
				slog.String("image_id", input.ImageID),
				slog.Int("status_code", resp.StatusCode))
			return nil
		}

		// Log error response
		w.logger.Error("Cloud Run job returned error",
			slog.Int("status_code", resp.StatusCode),
			slog.String("response", string(body)),
			slog.String("image_id", input.ImageID))

		lastErr = fmt.Errorf("cloud run job failed with status %d: %s", resp.StatusCode, string(body))

		// Don't retry on 4xx errors (except 429)
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return &event_handler.EventError{
				Err:       lastErr,
				Retryable: false,
				Category:  events.CategoryProcessing,
				Severity:  events.SeverityHigh,
			}
		}
	}

	// All retries exhausted
	return &event_handler.EventError{
		Err:       fmt.Errorf("failed to trigger Cloud Run job after %d attempts: %w", w.maxRetries, lastErr),
		Retryable: true,
		Category:  events.CategoryNetwork,
		Severity:  events.SeverityHigh,
	}
}

// GetStatus checks the status of a processing job
func (w *CloudRunWorker) GetStatus(ctx context.Context, jobID string) (string, error) {
	// This would require Cloud Run Jobs API integration
	// For now, return not implemented
	return "", fmt.Errorf("status check not implemented")
}
