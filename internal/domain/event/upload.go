package event

import (
	"github.com/histopathai/main-service/internal/domain/model"
)

type UploadEvent struct {
	BaseEvent
	model.Content
}
