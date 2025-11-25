package events

import "time"

const (
	EventTypeTelemetryDLQMessage EventType = "telemetry.dlq.message.v1"
	EventTypeTelemetryError      EventType = "telemetry.error.v1"
)

// ErrorSeverity indicates the severity level of an error
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "CRITICAL" // System-breaking errors
	SeverityHigh     ErrorSeverity = "HIGH"     // Major functionality impact
	SeverityMedium   ErrorSeverity = "MEDIUM"   // Degraded functionality
	SeverityLow      ErrorSeverity = "LOW"      // Minor issues
)

// ErrorCategory categorizes the type of error
type ErrorCategory string

const (
	CategoryProcessing    ErrorCategory = "PROCESSING"    // Image processing errors
	CategoryStorage       ErrorCategory = "STORAGE"       // GCS/Storage errors
	CategoryDatabase      ErrorCategory = "DATABASE"      // Firestore errors
	CategoryValidation    ErrorCategory = "VALIDATION"    // Input validation errors
	CategoryNetwork       ErrorCategory = "NETWORK"       // Network/connectivity errors
	CategorySerialization ErrorCategory = "SERIALIZATION" // JSON/data serialization errors
	CategoryUnknown       ErrorCategory = "UNKNOWN"       // Unclassified errors
)

// DLQMessageEvent represents a message that was sent to a DLQ
type DLQMessageEvent struct {
	BaseEvent
	OriginalTopic        string        `json:"original-topic"`
	OriginalSubscription string        `json:"original-subscription"`
	MessageID            string        `json:"message-id"`
	OriginalEventType    EventType     `json:"original-event-type"`
	OriginalPayload      string        `json:"original-payload"` // Base64 encoded or JSON string
	ErrorMessage         string        `json:"error-message"`
	ErrorCategory        ErrorCategory `json:"error-category"`
	ErrorSeverity        ErrorSeverity `json:"error-severity"`
	RetryCount           int           `json:"retry-count"`
	StackTrace           *string       `json:"stack-trace,omitempty"`
	ImageID              *string       `json:"image-id,omitempty"`
	PatientID            *string       `json:"patient-id,omitempty"`
	UserID               *string       `json:"user-id,omitempty"`
}

// TelemetryErrorEvent represents application errors for monitoring
type TelemetryErrorEvent struct {
	BaseEvent
	Service       string                 `json:"service"`
	Operation     string                 `json:"operation"`
	ErrorMessage  string                 `json:"error-message"`
	ErrorCategory ErrorCategory          `json:"error-category"`
	ErrorSeverity ErrorSeverity          `json:"error-severity"`
	StackTrace    *string                `json:"stack-trace,omitempty"`
	Context       map[string]interface{} `json:"context,omitempty"`
	ImageID       *string                `json:"image-id,omitempty"`
	PatientID     *string                `json:"patient-id,omitempty"`
	UserID        *string                `json:"user-id,omitempty"`
}

// NewDLQMessageEvent creates a new DLQ message event
func NewDLQMessageEvent(
	originalTopic, originalSubscription, messageID string,
	originalEventType EventType,
	originalPayload, errorMessage string,
	category ErrorCategory,
	severity ErrorSeverity,
	retryCount int,
) DLQMessageEvent {
	return DLQMessageEvent{
		BaseEvent:            NewBaseEvent(EventTypeTelemetryDLQMessage),
		OriginalTopic:        originalTopic,
		OriginalSubscription: originalSubscription,
		MessageID:            messageID,
		OriginalEventType:    originalEventType,
		OriginalPayload:      originalPayload,
		ErrorMessage:         errorMessage,
		ErrorCategory:        category,
		ErrorSeverity:        severity,
		RetryCount:           retryCount,
	}
}

// NewTelemetryErrorEvent creates a new telemetry error event
func NewTelemetryErrorEvent(
	service, operation, errorMessage string,
	category ErrorCategory,
	severity ErrorSeverity,
) TelemetryErrorEvent {
	return TelemetryErrorEvent{
		BaseEvent:     NewBaseEvent(EventTypeTelemetryError),
		Service:       service,
		Operation:     operation,
		ErrorMessage:  errorMessage,
		ErrorCategory: category,
		ErrorSeverity: severity,
		Context:       make(map[string]interface{}),
	}
}

// WithStackTrace adds a stack trace to the DLQ message event
func (e *DLQMessageEvent) WithStackTrace(trace string) *DLQMessageEvent {
	e.StackTrace = &trace
	return e
}

// WithImageID adds an image ID to the event
func (e *DLQMessageEvent) WithImageID(imageID string) *DLQMessageEvent {
	e.ImageID = &imageID
	return e
}

// WithPatientID adds a patient ID to the event
func (e *DLQMessageEvent) WithPatientID(patientID string) *DLQMessageEvent {
	e.PatientID = &patientID
	return e
}

// WithUserID adds a user ID to the event
func (e *DLQMessageEvent) WithUserID(userID string) *DLQMessageEvent {
	e.UserID = &userID
	return e
}

// WithStackTrace adds a stack trace to the telemetry error event
func (e *TelemetryErrorEvent) WithStackTrace(trace string) *TelemetryErrorEvent {
	e.StackTrace = &trace
	return e
}

// WithContext adds context information to the event
func (e *TelemetryErrorEvent) WithContext(key string, value interface{}) *TelemetryErrorEvent {
	e.Context[key] = value
	return e
}

// WithImageID adds an image ID to the event
func (e *TelemetryErrorEvent) WithImageID(imageID string) *TelemetryErrorEvent {
	e.ImageID = &imageID
	return e
}

// WithPatientID adds a patient ID to the event
func (e *TelemetryErrorEvent) WithPatientID(patientID string) *TelemetryErrorEvent {
	e.PatientID = &patientID
	return e
}

// WithUserID adds a user ID to the event
func (e *TelemetryErrorEvent) WithUserID(userID string) *TelemetryErrorEvent {
	e.UserID = &userID
	return e
}

// TelemetryMessage represents a telemetry message for UI display
type TelemetryMessage struct {
	ID                string                 `json:"id"`
	Timestamp         time.Time              `json:"timestamp"`
	Service           string                 `json:"service"`
	Operation         string                 `json:"operation"`
	ErrorMessage      string                 `json:"error_message"`
	ErrorCategory     ErrorCategory          `json:"error_category"`
	ErrorSeverity     ErrorSeverity          `json:"error_severity"`
	ImageID           *string                `json:"image_id,omitempty"`
	PatientID         *string                `json:"patient_id,omitempty"`
	UserID            *string                `json:"user_id,omitempty"`
	RetryCount        int                    `json:"retry_count"`
	OriginalEventType *EventType             `json:"original_event_type,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}
