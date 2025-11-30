package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type TelemetryService struct {
	telemetryRepo port.TelemetryRepository
}

func NewTelemetryService(
	telemetryRepo port.TelemetryRepository,
) *TelemetryService {
	return &TelemetryService{
		telemetryRepo: telemetryRepo,
	}
}

func (ts *TelemetryService) RecordDLQMessage(ctx context.Context, event *events.DLQMessageEvent) error {
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

	return ts.telemetryRepo.Create(ctx, message)
}

func (ts *TelemetryService) RecordError(ctx context.Context, event *events.TelemetryErrorEvent) error {
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

	return ts.telemetryRepo.Create(ctx, message)
}

func (ts *TelemetryService) ListMessages(ctx context.Context, filters []sharedQuery.Filter, pagination *sharedQuery.Pagination) (*sharedQuery.Result[*events.TelemetryMessage], error) {
	return ts.telemetryRepo.FindByFilters(ctx, filters, pagination)
}

func (ts *TelemetryService) GetMessageByID(ctx context.Context, id string) (*events.TelemetryMessage, error) {
	return ts.telemetryRepo.Read(ctx, id)
}

func (ts *TelemetryService) GetErrorStats(ctx context.Context) (*port.TelemetryStats, error) {
	// Get total count
	totalCount, err := ts.telemetryRepo.Count(ctx, []sharedQuery.Filter{})
	if err != nil {
		return nil, err
	}

	// Get counts by severity
	bySeverity := make(map[events.ErrorSeverity]int64)
	for _, severity := range []events.ErrorSeverity{
		events.SeverityCritical,
		events.SeverityHigh,
		events.SeverityMedium,
		events.SeverityLow,
	} {
		count, err := ts.telemetryRepo.Count(ctx, []sharedQuery.Filter{
			{Field: "error_severity", Operator: sharedQuery.OpEqual, Value: severity},
		})
		if err != nil {
			return nil, err
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
		count, err := ts.telemetryRepo.Count(ctx, []sharedQuery.Filter{
			{Field: "error_category", Operator: sharedQuery.OpEqual, Value: category},
		})
		if err != nil {
			return nil, err
		}
		byCategory[category] = count
	}

	// Get recent errors (last 10)
	recentResult, err := ts.telemetryRepo.FindByFilters(ctx, []sharedQuery.Filter{}, &sharedQuery.Pagination{
		Limit: 10,
	})
	if err != nil {
		return nil, err
	}

	return &port.TelemetryStats{
		TotalErrors:  totalCount,
		BySeverity:   bySeverity,
		ByCategory:   byCategory,
		RecentErrors: recentResult.Data,
	}, nil
}
