// adapter/event/pubsub/serializer.go
package pubsub

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type EventSerializer struct{}

func NewEventSerializer() *EventSerializer {
	return &EventSerializer{}
}

func (s *EventSerializer) Serialize(event domainevent.Event) ([]byte, error) {
	var dto interface{}

	switch e := event.(type) {
	case *domainevent.UploadedEvent:
		dto = uploadedEventDTO{
			EventID:   e.EventID,
			EventType: string(e.EventType),
			Timestamp: e.Timestamp.Format(time.RFC3339),
			Content:   contentToDTO(e.Content),
		}

	case *domainevent.DeleteEvent:
		dto = deleteEventDTO{
			EventID:   e.EventID,
			EventType: string(e.EventType),
			Timestamp: e.Timestamp.Format(time.RFC3339),
			Content:   contentToDTO(e.Content),
		}

	case *domainevent.ImageProcessingRequestedEvent:
		dto = imageProcessingRequestedDTO{
			EventID:           e.EventID,
			EventType:         string(e.EventType),
			Timestamp:         e.Timestamp.Format(time.RFC3339),
			ID:                e.ID,
			Content:           contentToDTO(e.Content),
			ProcessingVersion: string(e.ProcessingVersion),
		}

	case *domainevent.ImageProcessingCompletedEvent:
		var result *processingResultDTO
		if e.Result != nil {
			result = &processingResultDTO{
				Width:  e.Result.Width,
				Height: e.Result.Height,
				Size:   e.Result.Size,
			}
		}

		dto = imageProcessingCompletedDTO{
			EventID:       e.EventID,
			EventType:     string(e.EventType),
			Timestamp:     e.Timestamp.Format(time.RFC3339),
			ID:            e.ID,
			Content:       contentToDTO(e.Content),
			Success:       e.Success,
			Result:        result,
			FailureReason: e.FailureReason,
			Retryable:     e.Retryable,
		}

	default:
		return nil, fmt.Errorf("unsupported event type: %T", event)
	}

	return json.Marshal(dto)
}

func (s *EventSerializer) Deserialize(data []byte, eventType domainevent.EventType) (domainevent.Event, error) {
	switch eventType {
	case domainevent.UploadedEventType:
		var dto uploadedEventDTO
		if err := json.Unmarshal(data, &dto); err == nil {
			return s.uploadedDTOToDomain(dto)
		}

		return s.parseGCSNotificationToUploadedEvent(data)

	case domainevent.DeleteEventType:
		var dto deleteEventDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.deleteDTOToDomain(dto)

	case domainevent.ImageProcessingRequestedEventType:
		var dto imageProcessingRequestedDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.processingRequestedDTOToDomain(dto)

	case domainevent.ImageProcessingCompletedEventType:
		var dto imageProcessingCompletedDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.processingCompletedDTOToDomain(dto)

	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}

type uploadedEventDTO struct {
	EventID   string     `json:"event_id"`
	EventType string     `json:"event_type"`
	Timestamp string     `json:"timestamp"`
	Content   contentDTO `json:"content"`
}

type deleteEventDTO struct {
	EventID   string     `json:"event_id"`
	EventType string     `json:"event_type"`
	Timestamp string     `json:"timestamp"`
	Content   contentDTO `json:"content"`
}

type imageProcessingRequestedDTO struct {
	EventID           string     `json:"event_id"`
	EventType         string     `json:"event_type"`
	Timestamp         string     `json:"timestamp"`
	ID                string     `json:"id"`
	Content           contentDTO `json:"content"`
	ProcessingVersion string     `json:"processing_version"`
}

type imageProcessingCompletedDTO struct {
	EventID       string               `json:"event_id"`
	EventType     string               `json:"event_type"`
	Timestamp     string               `json:"timestamp"`
	ID            string               `json:"id"`
	Content       contentDTO           `json:"content"`
	Success       bool                 `json:"success"`
	Result        *processingResultDTO `json:"result,omitempty"`
	FailureReason string               `json:"failure_reason,omitempty"`
	Retryable     bool                 `json:"retryable"`
}

type processingResultDTO struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Size   int64 `json:"size"`
}

type contentDTO struct {
	Provider    string            `json:"provider"`
	Path        string            `json:"path"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

func contentToDTO(c vobj.Content) contentDTO {
	return contentDTO{
		Provider:    string(c.Provider),
		Path:        c.Path,
		ContentType: string(c.ContentType),
		Size:        c.Size,
	}
}

func dtoToContent(dto contentDTO) vobj.Content {
	return vobj.Content{
		Provider:    vobj.ContentProvider(dto.Provider),
		Path:        dto.Path,
		ContentType: vobj.ContentType(dto.ContentType),
		Size:        dto.Size,
	}
}

func (s *EventSerializer) uploadedDTOToDomain(dto uploadedEventDTO) (*domainevent.UploadedEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.UploadedEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		Content: dtoToContent(dto.Content),
	}, nil
}

func (s *EventSerializer) deleteDTOToDomain(dto deleteEventDTO) (*domainevent.DeleteEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.DeleteEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		Content: dtoToContent(dto.Content),
	}, nil
}

func (s *EventSerializer) processingRequestedDTOToDomain(dto imageProcessingRequestedDTO) (*domainevent.ImageProcessingRequestedEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.ImageProcessingRequestedEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		ID:                dto.ID,
		Content:           dtoToContent(dto.Content),
		ProcessingVersion: vobj.ProcessingVersion(dto.ProcessingVersion),
	}, nil
}

func (s *EventSerializer) processingCompletedDTOToDomain(dto imageProcessingCompletedDTO) (*domainevent.ImageProcessingCompletedEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	var result *domainevent.ProcessingResult
	if dto.Result != nil {
		result = &domainevent.ProcessingResult{
			Width:  dto.Result.Width,
			Height: dto.Result.Height,
			Size:   dto.Result.Size,
		}
	}

	return &domainevent.ImageProcessingCompletedEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		ID:            dto.ID,
		Content:       dtoToContent(dto.Content),
		Success:       dto.Success,
		Result:        result,
		FailureReason: dto.FailureReason,
		Retryable:     dto.Retryable,
	}, nil
}

func (s *EventSerializer) parseGCSNotificationToUploadedEvent(data []byte) (*domainevent.UploadedEvent, error) {
	// GCS notification yapısı
	var gcsNotif struct {
		Kind        string            `json:"kind"`
		ID          string            `json:"id"`
		Name        string            `json:"name"`
		Bucket      string            `json:"bucket"`
		ContentType string            `json:"contentType"`
		Size        string            `json:"size"`
		Metadata    map[string]string `json:"metadata"` // Intereted section
		TimeCreated string            `json:"timeCreated"`
	}

	if err := json.Unmarshal(data, &gcsNotif); err != nil {
		return nil, fmt.Errorf("failed to parse as GCS notification: %w", err)
	}

	if gcsNotif.Metadata["image-id"] == "" {
		return nil, fmt.Errorf("image-id not found in GCS notification metadata")
	}

	size, err := strconv.ParseInt(gcsNotif.Size, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse size: %w", err)
	}

	timestamp, err := time.Parse(time.RFC3339, gcsNotif.TimeCreated)
	if err != nil {
		timestamp = time.Now()
	}

	return &domainevent.UploadedEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   gcsNotif.ID,
			EventType: domainevent.UploadedEventType,
			Timestamp: timestamp,
		},
		ID: gcsNotif.Metadata["image-id"],
		Content: vobj.Content{
			Provider:    vobj.ContentProvider(gcsNotif.Metadata["origin-provider"]),
			Path:        gcsNotif.Metadata["origin-path"],
			ContentType: vobj.ContentType(gcsNotif.Metadata["content-type"]),
			Size:        size,
		},
	}, nil
}
