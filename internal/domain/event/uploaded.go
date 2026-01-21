package event

import "github.com/histopathai/main-service/internal/domain/vobj"

type UploadedEvent struct {
	BaseEvent
	ID      string
	Content vobj.Content
}
