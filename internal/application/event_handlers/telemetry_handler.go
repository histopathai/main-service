package eventhandlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

// TelemetryDLQHandler handles DLQ messages for telemetry
type TelemetryDLQHandler struct {
	*BaseEventHandler
	telemetryRepo port.TelemetryRepository
	logger        *slog.Logger
}

// NewTelemetryDLQHandler creates a new telemetry DLQ handler
func NewTelemetryDLQHandler(
	telemetryRepo port.TelemetryRepository,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	logger *slog.Logger,
) *TelemetryDLQHandler {
	return &TelemetryDLQHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		telemetryRepo: telemetryRepo,
		logger:        logger,
	}
}

// Handle processes DLQ telemetry events
func (h *TelemetryDLQHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *TelemetryDLQHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event
	var event events.DLQMessageEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	h.logger.Info("Processing DLQ telemetry event",
		slog.String("message_id", event.MessageID),
		slog.String("original_topic", event.OriginalTopic),
		slog.String("error_category", string(event.ErrorCategory)),
		slog.String("error_severity", string(event.ErrorSeverity)))

	// Convert to telemetry message
	message := &events.TelemetryMessage{
		ID:                event.EventID,
		Timestamp:         event.Timestamp,
		Service:           "DLQ",
		Operation:         event.OriginalTopic,
		ErrorMessage:      event.ErrorMessage,
		ErrorCategory:     event.ErrorCategory,
		ErrorSeverity:     event.ErrorSeverity,
		ImageID:           event.ImageID,
		PatientID:         event.PatientID,
		UserID:            event.UserID,
		RetryCount:        event.RetryCount,
		OriginalEventType: &event.OriginalEventType,
		Metadata: map[string]interface{}{
			"original_subscription": event.OriginalSubscription,
			"message_id":            event.MessageID,
			"original_payload":      event.OriginalPayload,
		},
	}

	if event.StackTrace != nil {
		message.Metadata["stack_trace"] = *event.StackTrace
	}

	// Store in telemetry repository
	if err := h.telemetryRepo.Create(ctx, message); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to store DLQ message: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("message_id", event.MessageID)
	}

	h.logger.Info("Successfully stored DLQ telemetry message",
		slog.String("message_id", event.MessageID))

	return nil
}

// TelemetryErrorHandler handles error telemetry events
type TelemetryErrorHandler struct {
	*BaseEventHandler
	telemetryRepo port.TelemetryRepository
	logger        *slog.Logger
}

// NewTelemetryErrorHandler creates a new telemetry error handler
func NewTelemetryErrorHandler(
	telemetryRepo port.TelemetryRepository,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	logger *slog.Logger,
) *TelemetryErrorHandler {
	return &TelemetryErrorHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		telemetryRepo: telemetryRepo,
		logger:        logger,
	}
}

// Handle processes error telemetry events
func (h *TelemetryErrorHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *TelemetryErrorHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event
	var event events.TelemetryErrorEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	h.logger.Info("Processing error telemetry event",
		slog.String("service", event.Service),
		slog.String("operation", event.Operation),
		slog.String("error_category", string(event.ErrorCategory)),
		slog.String("error_severity", string(event.ErrorSeverity)))

	// Convert to telemetry message
	message := &events.TelemetryMessage{
		ID:            event.EventID,
		Timestamp:     event.Timestamp,
		Service:       event.Service,
		Operation:     event.Operation,
		ErrorMessage:  event.ErrorMessage,
		ErrorCategory: event.ErrorCategory,
		ErrorSeverity: event.ErrorSeverity,
		ImageID:       event.ImageID,
		PatientID:     event.PatientID,
		UserID:        event.UserID,
		Metadata:      event.Context,
	}

	if event.StackTrace != nil {
		if message.Metadata == nil {
			message.Metadata = make(map[string]interface{})
		}
		message.Metadata["stack_trace"] = *event.StackTrace
	}

	// Store in telemetry repository
	if err := h.telemetryRepo.Create(ctx, message); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to store error telemetry: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("service", event.Service).
			WithContext("operation", event.Operation)
	}

	h.logger.Info("Successfully stored error telemetry message",
		slog.String("service", event.Service),
		slog.String("operation", event.Operation))

	return nil
}

// TelemetryAggregator aggregates telemetry data for dashboard
type TelemetryAggregator struct {
	telemetryRepo port.TelemetryRepository
	logger        *slog.Logger
}

// NewTelemetryAggregator creates a new telemetry aggregator
func NewTelemetryAggregator(
	telemetryRepo port.TelemetryRepository,
	logger *slog.Logger,
) *TelemetryAggregator {
	return &TelemetryAggregator{
		telemetryRepo: telemetryRepo,
		logger:        logger,
	}
}

// GetDashboardStats aggregates telemetry stats for dashboard
func (a *TelemetryAggregator) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	a.logger.Info("Aggregating telemetry stats for dashboard")

	// Get total error count
	totalCount, err := a.telemetryRepo.Count(ctx, []query.Filter{})
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Get counts by severity
	bySeverity := make(map[events.ErrorSeverity]int64)
	for _, severity := range []events.ErrorSeverity{
		events.SeverityCritical,
		events.SeverityHigh,
		events.SeverityMedium,
		events.SeverityLow,
	} {
		count, err := a.telemetryRepo.Count(ctx, []query.Filter{
			{Field: "error_severity", Operator: query.OpEqual, Value: severity},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to count by severity %s: %w", severity, err)
		}
		bySeverity[severity] = count
	}

	// Get counts by category
	byCategory := make(map[events.ErrorCategory]int64)
	for _, category := range []events.ErrorCategory{
		events.CategoryProcessing,
		events.CategoryStorage,
		events.CategoryDatabase,
		events.CategoryValidation,
		events.CategoryNetwork,
		events.CategorySerialization,
		events.CategoryUnknown,
	} {
		count, err := a.telemetryRepo.Count(ctx, []query.Filter{
			{Field: "error_category", Operator: query.OpEqual, Value: category},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to count by category %s: %w", category, err)
		}
		byCategory[category] = count
	}

	// Get recent errors
	recentResult, err := a.telemetryRepo.FindByFilters(ctx, []query.Filter{}, &query.Pagination{
		Limit: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get recent errors: %w", err)
	}

	stats := &DashboardStats{
		TotalErrors:  totalCount,
		BySeverity:   bySeverity,
		ByCategory:   byCategory,
		RecentErrors: recentResult.Data,
		Timestamp:    time.Now(),
	}

	a.logger.Info("Successfully aggregated telemetry stats",
		slog.Int64("total_errors", totalCount),
		slog.Int("recent_errors", len(recentResult.Data)))

	return stats, nil
}

// DashboardStats contains aggregated telemetry statistics
type DashboardStats struct {
	TotalErrors  int64                          `json:"total_errors"`
	BySeverity   map[events.ErrorSeverity]int64 `json:"by_severity"`
	ByCategory   map[events.ErrorCategory]int64 `json:"by_category"`
	RecentErrors []*events.TelemetryMessage     `json:"recent_errors"`
	Timestamp    time.Time                      `json:"timestamp"`
}
