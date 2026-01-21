package event

import "github.com/histopathai/main-service/internal/domain/vobj"

type DeleteEvent struct {
	BaseEvent
	ID      string
	Content vobj.Content
}
