package event

type EventType string

const (
	UploadEventType               EventType = "upload.v1"
	DeleteEventType               EventType = "delete.v1"
	ImageProcessEventType         EventType = "image.process.request.v1"
	ImageProcessCompleteEventType EventType = "image.process.complete.v1"
	ImageProcessDlqEventType      EventType = "image.process.dlq.v1"
	ImageUploadDlqEventType       EventType = "image.upload.dlq.v1"
)
