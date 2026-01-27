package pubsub

import (
	"encoding/json"
	"testing"
	"time"

	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/stretchr/testify/assert"
)

func TestEventSerializer_ParseGCSNotificationToNewFileExistEvent(t *testing.T) {
	serializer := NewEventSerializer()

	gcsNotification := map[string]interface{}{
		"kind":        "storage#object",
		"id":          "test-event-id",
		"name":        "path/to/image.svs",
		"bucket":      "test-bucket",
		"contentType": "image/x-aperio-svs",
		"size":        "123456",
		"timeCreated": "2023-10-26T12:00:00Z",
		"metadata": map[string]string{
			"id":           "img-123",
			"name":         "Test Image",
			"creator-id":   "user-1",
			"entity-type":  "image",
			"provider":     "gcs",
			"path":         "path/to/image.svs",
			"content-type": "image/x-aperio-svs",
			"size":         "123456",
		},
	}

	data, err := json.Marshal(gcsNotification)
	assert.NoError(t, err)

	event, err := serializer.Deserialize(data, domainevent.NewFileExistEventType)
	assert.NoError(t, err)
	assert.NotNil(t, event)

	newFileEvent, ok := event.(*domainevent.NewFileExistEvent)
	assert.True(t, ok)

	assert.Equal(t, "test-event-id", newFileEvent.EventID)
	assert.Equal(t, domainevent.NewFileExistEventType, newFileEvent.EventType)

	// Check Content model populated from metadata
	assert.Equal(t, "img-123", newFileEvent.Content.ID)
	assert.Equal(t, "Test Image", newFileEvent.Content.Name)
	assert.Equal(t, vobj.EntityTypeImage, newFileEvent.Content.EntityType)
	assert.Equal(t, vobj.ContentProviderGCS, newFileEvent.Content.Provider)
	assert.Equal(t, "path/to/image.svs", newFileEvent.Content.Path)
	assert.Equal(t, vobj.ContentTypeImageSVS, newFileEvent.Content.ContentType)
	assert.Equal(t, int64(123456), newFileEvent.Content.Size)
}

func TestEventSerializer_ImageProcessCompleteEvent_RoundTrip(t *testing.T) {
	serializer := NewEventSerializer()

	originalEvent := &domainevent.ImageProcessCompleteEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   "evt-123",
			EventType: domainevent.ImageProcessCompleteEventType,
			Timestamp: time.Now().Truncate(time.Second), // Truncate for JSON precision
		},
		ImageID: "proc-123",
		Contents: []model.Content{
			{
				Entity: vobj.Entity{
					ID:         "content-1",
					Name:       "image.svs",
					EntityType: vobj.EntityTypeImage,
				},
				Provider:    vobj.ContentProviderGCS,
				Path:        "output/image.svs",
				ContentType: vobj.ContentTypeImageSVS,
				Size:        1000,
			},
		},
		Success: true,
		Result: &domainevent.ProcessResult{
			Width:  100,
			Height: 100,
			Size:   1000,
		},
	}

	// Serialize
	data, err := serializer.Serialize(originalEvent)
	assert.NoError(t, err)

	// Deserialize
	deserializedEvent, err := serializer.Deserialize(data, domainevent.ImageProcessCompleteEventType)
	assert.NoError(t, err)

	resultEvent, ok := deserializedEvent.(*domainevent.ImageProcessCompleteEvent)
	assert.True(t, ok)

	assert.Equal(t, originalEvent.EventID, resultEvent.EventID)
	assert.Equal(t, originalEvent.ImageID, resultEvent.ImageID)
	assert.Equal(t, originalEvent.Success, resultEvent.Success)
	assert.Equal(t, len(originalEvent.Contents), len(resultEvent.Contents))

	// Check Content details
	assert.Equal(t, originalEvent.Contents[0].ID, resultEvent.Contents[0].ID)
	assert.Equal(t, originalEvent.Contents[0].Name, resultEvent.Contents[0].Name)
	// assert.Equal(t, originalEvent.Contents[0].EntityType, resultEvent.Contents[0].EntityType) // Check if EntityType was preserved
}
