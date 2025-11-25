package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
)

type RetryConfig struct {
	MaxAttempts       int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:       5,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

func (rc RetryConfig) CalculateBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		return rc.InitialBackoff
	}

	backoff := float64(rc.InitialBackoff)
	for i := 0; i < attempt; i++ {
		backoff *= rc.BackoffMultiplier
		if backoff >= float64(rc.MaxBackoff) {
			return rc.MaxBackoff
		}
	}

	return time.Duration(backoff)
}

type EventError struct {
	Err        error
	Retryable  bool
	Category   events.ErrorCategory
	Severity   events.ErrorSeverity
	Context    map[string]interface{}
	StackTrace string
}

func (e *EventError) Error() string {
	return fmt.Sprintf("[%s][%s] %v", e.Category, e.Severity, e.Err)
}

func NewRetryableError(err error, category events.ErrorCategory, severity events.ErrorSeverity) *EventError {
	return &EventError{
		Err:       err,
		Retryable: true,
		Category:  category,
		Severity:  severity,
		Context:   make(map[string]interface{}),
	}
}

func NewNonRetryableError(err error, category events.ErrorCategory, severity events.ErrorSeverity) *EventError {
	return &EventError{
		Err:       err,
		Retryable: false,
		Category:  category,
		Severity:  severity,
		Context:   make(map[string]interface{}),
	}
}

// WithContext adds context to the error
func (e *EventError) WithContext(key string, value interface{}) *EventError {
	e.Context[key] = value
	return e
}

// WithStackTrace adds stack trace to the error
func (e *EventError) WithStackTrace(trace string) *EventError {
	e.StackTrace = trace
	return e
}

// BaseEventHandler provides common functionality for all event handlers
type BaseEventHandler struct {
	logger             *slog.Logger
	serializer         events.EventSerializer
	telemetryPublisher port.TelemetryEventPublisher
	retryConfig        RetryConfig
}

// NewBaseEventHandler creates a new base event handler
func NewBaseEventHandler(
	logger *slog.Logger,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	retryConfig RetryConfig,
) *BaseEventHandler {
	return &BaseEventHandler{
		logger:             logger,
		serializer:         serializer,
		telemetryPublisher: telemetryPublisher,
		retryConfig:        retryConfig,
	}
}

// HandleWithRetry wraps event handling with retry logic
func (b *BaseEventHandler) HandleWithRetry(
	ctx context.Context,
	data []byte,
	attributes map[string]string,
	handler func(ctx context.Context, data []byte, attributes map[string]string) error,
) error {
	attempt := b.getAttemptCount(attributes)

	// Check if we've exceeded max attempts
	if attempt >= b.retryConfig.MaxAttempts {
		b.logger.Error("Max retry attempts exceeded",
			slog.Int("attempt", attempt),
			slog.Int("max_attempts", b.retryConfig.MaxAttempts))

		// Send to DLQ via telemetry
		if err := b.sendToDLQ(ctx, data, attributes, attempt, "max attempts exceeded"); err != nil {
			b.logger.Error("Failed to send to DLQ", slog.String("error", err.Error()))
		}

		// Acknowledge the message to prevent further retries
		return nil
	}

	// Execute handler
	err := handler(ctx, data, attributes)
	if err == nil {
		return nil
	}

	// Check if error is retryable
	eventErr, ok := err.(*EventError)
	if !ok {
		// Unknown error type - treat as retryable with high severity
		eventErr = NewRetryableError(err, events.CategoryUnknown, events.SeverityHigh)
	}

	// Log error
	b.logger.Error("Event handling failed",
		slog.Int("attempt", attempt),
		slog.Bool("retryable", eventErr.Retryable),
		slog.String("category", string(eventErr.Category)),
		slog.String("severity", string(eventErr.Severity)),
		slog.String("error", eventErr.Error()))

	// Publish telemetry error
	telemetryErr := events.NewTelemetryErrorEvent(
		"event-handler",
		b.getOperationName(attributes),
		eventErr.Error(),
		eventErr.Category,
		eventErr.Severity,
	)

	if eventErr.StackTrace != "" {
		telemetryErr.WithStackTrace(eventErr.StackTrace)
	}

	for key, value := range eventErr.Context {
		telemetryErr.WithContext(key, value)
	}

	if publishErr := b.telemetryPublisher.PublishTelemetryError(ctx, &telemetryErr); publishErr != nil {
		b.logger.Error("Failed to publish telemetry error", slog.String("error", publishErr.Error()))
	}

	// Handle based on retryability
	if !eventErr.Retryable {
		b.logger.Error("Non-retryable error, sending to DLQ")

		if dlqErr := b.sendToDLQ(ctx, data, attributes, attempt, eventErr.Error()); dlqErr != nil {
			b.logger.Error("Failed to send to DLQ", slog.String("error", dlqErr.Error()))
		}

		// Acknowledge to prevent retry
		return nil
	}

	// Calculate backoff and return error to trigger retry
	backoff := b.retryConfig.CalculateBackoff(attempt)
	b.logger.Info("Will retry after backoff",
		slog.Duration("backoff", backoff),
		slog.Int("next_attempt", attempt+1))

	// Return error to trigger Pub/Sub retry with exponential backoff
	return eventErr
}

// sendToDLQ sends failed message to DLQ via telemetry
func (b *BaseEventHandler) sendToDLQ(
	ctx context.Context,
	data []byte,
	attributes map[string]string,
	retryCount int,
	errorMessage string,
) error {
	originalTopic := attributes["original-topic"]
	if originalTopic == "" {
		originalTopic = "unknown"
	}

	originalSubscription := attributes["subscription"]
	if originalSubscription == "" {
		originalSubscription = "unknown"
	}

	messageID := attributes["message-id"]
	if messageID == "" {
		messageID = "unknown"
	}

	eventType := events.EventType(attributes["event-type"])

	dlqEvent := events.NewDLQMessageEvent(
		originalTopic,
		originalSubscription,
		messageID,
		eventType,
		string(data),
		errorMessage,
		events.CategoryUnknown,
		events.SeverityHigh,
		retryCount,
	)

	// Add optional IDs if present
	if imageID := attributes["image-id"]; imageID != "" {
		dlqEvent.WithImageID(imageID)
	}
	if patientID := attributes["patient-id"]; patientID != "" {
		dlqEvent.WithPatientID(patientID)
	}
	if userID := attributes["user-id"]; userID != "" {
		dlqEvent.WithUserID(userID)
	}

	return b.telemetryPublisher.PublishTelemetryDLQMessage(ctx, &dlqEvent)
}

// getAttemptCount extracts retry attempt count from attributes
func (b *BaseEventHandler) getAttemptCount(attributes map[string]string) int {
	// Pub/Sub automatically adds delivery attempt count
	if attemptStr, ok := attributes["googclient_deliveryattempt"]; ok {
		var attempt int
		if _, err := fmt.Sscanf(attemptStr, "%d", &attempt); err == nil {
			return attempt
		}
	}
	return 1
}

// getOperationName extracts operation name from attributes
func (b *BaseEventHandler) getOperationName(attributes map[string]string) string {
	if op := attributes["operation"]; op != "" {
		return op
	}
	if eventType := attributes["event-type"]; eventType != "" {
		return string(eventType)
	}
	return "unknown"
}

// DeserializeEvent deserializes event data
func (b *BaseEventHandler) DeserializeEvent(data []byte, event interface{}) error {
	if err := json.Unmarshal(data, event); err != nil {
		return NewNonRetryableError(
			fmt.Errorf("failed to deserialize event: %w", err),
			events.CategorySerialization,
			events.SeverityHigh,
		)
	}
	return nil
}
