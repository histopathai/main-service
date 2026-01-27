package event

type EventType string

const (
	ImageProcessReqEventType      EventType = "image.process.request.v1"
	ImageProcessCompleteEventType EventType = "image.process.complete.v1"
	ImageProcessDlqEventType      EventType = "image.process.dlq.v1"
)

const (
	NewFileExistEventType EventType = "new.file.exist.v1"
	DeleteFileEventType   EventType = "delete.file.v1"
)
