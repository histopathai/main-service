package handler

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

type UploadImageHandler struct {
	img_service service.IImageService
	logger      *slog.Logger
}

func NewUploadImageHandler(
	imageService service.IImageService,
	logger *slog.Logger,
) *UploadImageHandler {
	return &UploadImageHandler{
		img_service: imageService,
		logger:      logger,
	}
}

type Notification struct {
	Name     string            `json:"name"`
	Bucket   string            `json:"bucket"`
	Metadata map[string]string `json:"metadata"`
}

func (h *UploadImageHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {

	var n Notification

	if err := json.Unmarshal(data, &n); err != nil {
		h.logger.Error("Failed to unmarshal Upload notification, dropping message", slog.String("error", err.Error()), "data", string(data))
		return nil
	}

	metadata := n.Metadata
	if metadata == nil {
		h.logger.Error("Upload notification missing metadata block, dropping message", "data", string(data))
		return nil
	}

	imageID, ok := metadata["image-id"]
	if !ok {
		msg := "Upload metadata missing 'image-id', dropping message"
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

	if err := h.img_service.ConfirmUpload(ctx, input); err != nil {
		h.logger.Error("Failed to confirm upload", slog.String("error", err.Error()), "imageID", imageID)
		return errors.NewInternalError(fmt.Sprintf("Failed to confirm upload for imageID %s", imageID), err)
	}

	return nil
}
