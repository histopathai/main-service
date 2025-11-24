package events

const (
	EventTypeImageUploaded EventType = "image.uploaded.v1"
)

type ImageUploadedEvent struct {
	BaseEvent
	ImageID    string `json:"image-id"`
	PatientID  string `json:"patient-id"`
	CreatorID  string `json:"creator-id"`
	Name       string `json:"name"`
	Format     string `json:"format"`
	Width      *int   `json:"width,omitempty"`
	Height     *int   `json:"height,omitempty"`
	Size       *int64 `json:"size,omitempty"`
	OriginPath string `json:"origin-path"`
	Status     string `json:"status"`
}

func NewImageUploadedEvent(
	imageID, patientID, creatorID, Name, format string,
	width *int, height *int, size *int64,
	originPath, status string,
) ImageUploadedEvent {
	return ImageUploadedEvent{
		BaseEvent:  NewBaseEvent(EventTypeImageUploaded),
		ImageID:    imageID,
		PatientID:  patientID,
		CreatorID:  creatorID,
		Name:       Name,
		Format:     format,
		Width:      width,
		Height:     height,
		Size:       size,
		OriginPath: originPath,
		Status:     status,
	}
}
