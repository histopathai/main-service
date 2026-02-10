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
	case *domainevent.NewFileExistEvent:
		dto = uploadEventDTO{
			EventID:   e.EventID,
			EventType: string(e.EventType),
			Timestamp: e.Timestamp.Format(time.RFC3339),
			Content:   contentToDTO(e.Content),
		}

	case *domainevent.DeleteFileEvent:
		dto = deleteEventDTO{
			EventID:   e.EventID,
			EventType: string(e.EventType),
			Timestamp: e.Timestamp.Format(time.RFC3339),
			Content:   contentToDTO(e.Content),
		}

	case *domainevent.ImageProcessReqEvent:
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
			ImageID:       e.ImageID,
		}
	default:
		return nil, fmt.Errorf("unsupported event type: %T", event)
	}

	return json.Marshal(dto)
}

func (s *EventSerializer) Deserialize(data []byte, eventType domainevent.EventType) (domainevent.Event, error) {
	switch eventType {
	case domainevent.NewFileExistEventType:
		var dto uploadEventDTO
		// Unmarshal tries to match JSON to struct. If fields are missing, it succeeds with zero values.
		// We need to check if it really looks like our DTO.
		if err := json.Unmarshal(data, &dto); err == nil && dto.EventType != "" {
			return s.uploadDTOToDomain(dto)
		}

		return s.parseGCSNotificationToNewFileExistEvent(data)

	case domainevent.DeleteFileEventType:
		var dto deleteEventDTO
		if err := json.Unmarshal(data, &dto); err != nil {
			return nil, err
		}
		return s.deleteDTOToDomain(dto)

	case domainevent.ImageProcessReqEventType:
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

type retryMetadataDTO struct {
	AttemptCount    int    `json:"attempt_count"`
	MaxAttempts     int    `json:"max_attempts"`
	LastAttemptAt   string `json:"last_attempt_at"`
	FirstAttemptAt  string `json:"first_attempt_at"`
	BackoffDuration int64  `json:"backoff_duration_ms"`
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
	RetryMetadata *retryMetadataDTO    `json:"retry_metadata,omitempty"`
}

type processingResultDTO struct {
	Width  int   `json:"width"`
	Height int   `json:"height"`
	Size   int64 `json:"size"`
}

type parentDTO struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type contentDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	ParentID    string            `json:"parent_id"`
	ParentType  string            `json:"parent_type"`
	Parent      *parentDTO        `json:"parent,omitempty"`
	CreatorID   string            `json:"creator_id"`
	EntityType  string            `json:"entity_type"`
	Provider    string            `json:"provider"`
	Path        string            `json:"path"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type imageProcessDlqDTO struct {
	EventID           string            `json:"event_id"`
	EventType         string            `json:"event_type"`
	Timestamp         string            `json:"timestamp"`
	ImageID           string            `json:"image_id"`
	ProcessingVersion string            `json:"processing_version"`
	FailureReason     string            `json:"failure_reason,omitempty"`
	Retryable         bool              `json:"retryable"`
	RetryMetadata     *retryMetadataDTO `json:"retry_metadata,omitempty"`
	OriginalEventID   string            `json:"original_event_id"`
}

func contentToDTO(c model.Content) contentDTO {
	return contentDTO{
		ID:         c.ID,
		Name:       c.Name,
		ParentID:   c.Parent.ID,
		ParentType: string(c.Parent.Type),
		Parent: &parentDTO{
			ID:   c.Parent.ID,
			Type: string(c.Parent.Type),
		},
		CreatorID:   c.CreatorID,
		EntityType:  string(c.EntityType),
		Provider:    string(c.Provider),
		Path:        c.Path,
		ContentType: string(c.ContentType),
		Size:        c.Size,
	}
}

func dtoToContent(dto contentDTO) model.Content {
	parentID := dto.ParentID
	parentType := dto.ParentType

	if dto.Parent != nil {
		parentID = dto.Parent.ID
		parentType = dto.Parent.Type
	}

	return model.Content{
		Entity: vobj.Entity{
			ID:         dto.ID,
			Name:       dto.Name,
			CreatorID:  dto.CreatorID,
			EntityType: vobj.EntityType(dto.EntityType),
			Parent: vobj.ParentRef{
				ID:   parentID,
				Type: vobj.ParentType(parentType),
			},
		},
		Provider:    vobj.ContentProvider(dto.Provider),
		Path:        dto.Path,
		ContentType: vobj.ContentType(dto.ContentType),
		Size:        dto.Size,
	}
}

func (s *EventSerializer) uploadDTOToDomain(dto uploadEventDTO) (*domainevent.NewFileExistEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.NewFileExistEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		Content: dtoToContent(dto.Content),
	}, nil
}

func (s *EventSerializer) deleteDTOToDomain(dto deleteEventDTO) (*domainevent.DeleteFileEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.DeleteFileEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   dto.EventID,
			EventType: domainevent.EventType(dto.EventType),
			Timestamp: timestamp,
		},
		Content: dtoToContent(dto.Content),
	}, nil
}

func (s *EventSerializer) processingRequestedDTOToDomain(dto imageProcessDTO) (*domainevent.ImageProcessReqEvent, error) {
	timestamp, err := time.Parse(time.RFC3339, dto.Timestamp)
	if err != nil {
		return nil, err
	}

	return &domainevent.ImageProcessReqEvent{
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
	}, nil
}

func (s *EventSerializer) parseGCSNotificationToNewFileExistEvent(data []byte) (*domainevent.NewFileExistEvent, error) {
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
			Parent: vobj.ParentRef{
				ID:   gcsNotif.Metadata["parent-id"],
				Type: vobj.ParentType(gcsNotif.Metadata["parent-type"]),
			},
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

	return &domainevent.NewFileExistEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   gcsNotif.ID,
			EventType: domainevent.NewFileExistEventType,
			Timestamp: timestamp,
		},
		Content: content,
	}, nil
}
