package event

import "github.com/histopathai/main-service/internal/domain/model"

type NewFileExistEvent struct {
	BaseEvent
	model.Content
}

type DeleteFileEvent struct {
	BaseEvent
	model.Content
}
