// adapter/event/pubsub/serializer.go
package pubsub

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type EventSerializer struct{}

func NewEventSerializer() *EventSerializer {
	return &EventSerializer{}
}

func (s *EventSerializer) Serialize(event domainevent.Event) ([]byte, error) {
	var dto interface{}

	switch e := event.(type) {
	case *domainevent.UploadEvent:
		dto = uploadEventDTO{
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

	case *domainevent.ImageProcessEvent:
		dto = imageProcessDTO{
			EventID:           e.EventID,
			EventType:         string(e.EventType),
			Timestamp:         e.Timestamp.Format(time.RFC3339),
			Content:           contentToDTO(e.Content),
			ProcessingVersion: string(e.ProcessingVersion),
		}

	case *domainevent.ImageProcessCompleteEvent:
		var result *processingResultDTO
		if e.Result != nil {
			result = &processingResultDTO{
				Width:  e.Result.Width,
				Height: e.Result.Height,
				Size:   e.Result.Size,
			}
		}

		var dtoContents []contentDTO
		for _, c := range e.Contents {
			dtoContents = append(dtoContents, contentToDTO(c))
		}

		dto = imageProcessCompleteDTO{
			EventID:       e.EventID,
			EventType:     string(e.EventType),
			Timestamp:     e.Timestamp.Format(time.RFC3339),
			Contents:      dtoContents,
			Success:       e.Success,
			Result:        result,
			FailureReason: e.FailureReason,
			Retryable:     e.Retryable,
			ImageID:       e.ImageID,
		}
	case *domainevent.ImageProcessDlqEvent:
		dto = imageProcessDlqDTO{
			EventID:       e.EventID,
			EventType:     string(e.EventType),
			Timestamp:     e.Timestamp.Format(time.RFC3339),
			ID:            e.ID,
			Content:       contentToDTO(e.Content),
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
	case domainevent.UploadEventType:
		var dto uploadEventDTO
		// Unmarshal tries to match JSON to struct. If fields are missing, it succeeds with zero values.
		// We need to check if it really looks like our DTO.
		if err := json.Unmarshal(data, &dto); err == nil && dto.EventType != "" {
			return s.uploadDTOToDomain(dto)
		}

		return s.parseGCSNotificationToUploadEvent(data)

	case domainevent.DeleteEventType:
		var dto deleteEventDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.deleteDTOToDomain(dto)

	case domainevent.ImageProcessEventType:
		var dto imageProcessDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.processingRequestedDTOToDomain(dto)

	case domainevent.ImageProcessCompleteEventType:
		var dto imageProcessCompleteDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.processingCompletedDTOToDomain(dto)

	case domainevent.ImageProcessDlqEventType:
		var dto imageProcessDlqDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.imageProcessDlqDTOToDomain(dto)

	default:
		return nil, fmt.Errorf("unsupported event type: %s", eventType)
	}
}

type uploadEventDTO struct {
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

type imageProcessDTO struct {
	EventID           string     `json:"event_id"`
	EventType         string     `json:"event_type"`
	Timestamp         string     `json:"timestamp"`
	ID                string     `json:"id"`
	Content           contentDTO `json:"content"`
	ProcessingVersion string     `json:"processing_version"`
}

type imageProcessCompleteDTO struct {
	EventID       string               `json:"event_id"`
	EventType     string               `json:"event_type"`
	Timestamp     string               `json:"timestamp"`
	ImageID       string               `json:"image_id"`
	Contents      []contentDTO         `json:"contents"`
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
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	CreatorID   string            `json:"creator_id"`
	EntityType  string            `json:"entity_type"`
	Provider    string            `json:"provider"`
	Path        string            `json:"path"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type imageProcessDlqDTO struct {
	EventID       string     `json:"event_id"`
	EventType     string     `json:"event_type"`
	Timestamp     string     `json:"timestamp"`
	ID            string     `json:"id"`
	Content       contentDTO `json:"content"`
	FailureReason string     `json:"failure_reason,omitempty"`
	Retryable     bool       `json:"retryable"`
}

func contentToDTO(c model.Content) contentDTO {
	return contentDTO{
		ID:          c.ID,
		Name:        c.Name,
		CreatorID:   c.CreatorID,
		EntityType:  string(c.EntityType),
		Provider:    string(c.Provider),
		Path:        c.Path,
		ContentType: string(c.ContentType),
		Size:        c.Size,
	}
}

func dtoToContent(dto contentDTO) model.Content {
	return model.Content{
		Entity: vobj.Entity{
			ID:         dto.ID,
			Name:       dto.Name,
			CreatorID:  dto.CreatorID,
			EntityType: vobj.EntityType(dto.EntityType),
		},
		Provider:    vobj.ContentProvider(dto.Provider),
		Path:        dto.Path,
		ContentType: vobj.ContentType(dto.ContentType),
		Size:        dto.Size,
	}
}

func (s *EventSerializer) uploadDTOToDomain(dto uploadEventDTO) (*domainevent.UploadEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.UploadEvent{
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

func (s *EventSerializer) processingRequestedDTOToDomain(dto imageProcessDTO) (*domainevent.ImageProcessEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.ImageProcessEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		Content:           dtoToContent(dto.Content),
		ProcessingVersion: vobj.ProcessingVersion(dto.ProcessingVersion),
	}, nil
}

func (s *EventSerializer) processingCompletedDTOToDomain(dto imageProcessCompleteDTO) (*domainevent.ImageProcessCompleteEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	var result *domainevent.ProcessResult
	if dto.Result != nil {
		result = &domainevent.ProcessResult{
			Width:  dto.Result.Width,
			Height: dto.Result.Height,
			Size:   dto.Result.Size,
		}
	}

	return &domainevent.ImageProcessCompleteEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		ImageID: dto.ImageID,
		Contents: func() []model.Content {
			var contents []model.Content
			for _, c := range dto.Contents {
				contents = append(contents, dtoToContent(c))
			}
			return contents
		}(),
		Success:       dto.Success,
		Result:        result,
		FailureReason: dto.FailureReason,
		Retryable:     dto.Retryable,
	}, nil
}

func (s *EventSerializer) parseGCSNotificationToUploadEvent(data []byte) (*domainevent.UploadEvent, error) {
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

	// GCS'den gelen metadata'dan content bilgisini oluştur
	size, _ := strconv.ParseInt(gcsNotif.Size, 10, 64)
	if s, err := strconv.ParseInt(gcsNotif.Metadata["size"], 10, 64); err == nil && s > 0 {
		size = s
	}

	timestamp, err := time.Parse(time.RFC3339, gcsNotif.TimeCreated)
	if err != nil {
		timestamp = time.Now()
	}

	// Metadata'dan fieldları al
	content := model.Content{
		Entity: vobj.Entity{
			ID:         gcsNotif.Metadata["id"],
			Name:       gcsNotif.Metadata["name"],
			CreatorID:  gcsNotif.Metadata["creator-id"],
			EntityType: vobj.EntityType(gcsNotif.Metadata["entity-type"]),
		},
		Provider:    vobj.ContentProvider(gcsNotif.Metadata["provider"]),
		Path:        gcsNotif.Metadata["path"],
		ContentType: vobj.ContentType(gcsNotif.Metadata["content-type"]),
		Size:        size,
	}

	// Fallbacks if metadata is missing (e.g. direct upload without signed URL metadata)
	// This might happen if someone uploads directly to bucket without using our signed URL flow
	if content.Provider == "" {
		content.Provider = vobj.ContentProviderGCS
	}
	if content.Path == "" {
		content.Path = gcsNotif.Name
	}
	if content.ContentType == "" {
		content.ContentType = vobj.ContentType(gcsNotif.ContentType)
	}
	// Entity ID is crucial, if missing, we might not be able to link it.
	// But let's assume if it came from our signed URL, it has it.
	// If not, we can use the GCS ID as a fallback or leave it empty?
	if content.ID == "" {
		// Log warning or error? For now, let's use GCS ID generation or error out?
		// As per requirement "in image-id .... is not competible with content structs"
		// The user implies we should just parse content from metadata.
		// If metadata is empty, this event might be irrelevant for us or we handle partials.
		// Let's rely on metadata primarily.
	}

	return &domainevent.UploadEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   gcsNotif.ID,
			EventType: domainevent.UploadEventType,
			Timestamp: timestamp,
		},
		Content: content,
	}, nil
}

func (s *EventSerializer) imageProcessDlqDTOToDomain(dto imageProcessDlqDTO) (*domainevent.ImageProcessDlqEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.ImageProcessDlqEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		ID:            dto.ID,
		Content:       dtoToContent(dto.Content),
		FailureReason: dto.FailureReason,
		Retryable:     dto.Retryable,
	}, nil
}
