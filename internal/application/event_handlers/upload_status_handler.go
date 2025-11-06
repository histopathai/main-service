package eventhandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type UploadStatusHandler struct {
	imageService *service.ImageService
	logger       *slog.Logger
}

func NewUploadStatusHandler(
	imageService *service.ImageService,
	logger *slog.Logger,
) *UploadStatusHandler {
	return &UploadStatusHandler{
		imageService: imageService,
		logger:       logger,
	}
}

type GCSNotification struct {
	Name     string            `json:"name"`
	Bucket   string            `json:"bucket"`
	Metadata map[string]string `json:"metadata"`
}

func (h *UploadStatusHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	h.logger.Info("Processing GCS upload event",
		"EventID", attributes["event_id"],
	)

	var n GCSNotification
	if err := json.Unmarshal(data, &n); err != nil {
		h.logger.Error("Failed to unmarshal GCS notification, dropping message", slog.String("error", err.Error()), "data", string(data))
		return nil
	}

	metadata := n.Metadata
	if metadata == nil {
		h.logger.Error("GCS notification missing metadata block, dropping message", "data", string(data))
		return nil
	}

	imageID, ok := metadata["image-id"]
	if !ok {
		msg := "GCS metadata missing 'image-id', dropping message"
		h.logger.Error(msg, "data", string(data))
		return nil
	}

	input := &service.ConfirmUploadInput{
		ImageID:    imageID,
		PatientID:  metadata["patient-id"],
		CreatorID:  metadata["creator-id"],
		Name:       metadata["file-name"],
		Format:     metadata["format"],
		OriginPath: metadata["origin-path"],
		Status:     model.ImageStatus(metadata["status"]),
	}

	if widthStr, ok := metadata["width"]; ok {
		if width, err := strconv.Atoi(widthStr); err == nil {
			input.Width = &width
		}
	}
	if heightStr, ok := metadata["height"]; ok {
		if height, err := strconv.Atoi(heightStr); err == nil {
			input.Height = &height
		}
	}
	if sizeStr, ok := metadata["size"]; ok {
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			input.Size = &size
		}
	}

	if err := h.imageService.ConfirmUpload(ctx, input); err != nil {
		h.logger.Error("Failed to confirm image upload after GCS event, will retry", slog.String("error", err.Error()), "imageID", imageID)
		return errors.NewInternalError(fmt.Sprintf("Failed to confirm image upload: %v", err), err)
	}

	h.logger.Info("Successfully confirmed upload from GCS event",
		"ImageID", imageID,
		"Status", input.Status,
	)

	return nil
}
