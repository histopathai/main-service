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

type UploadMetadata struct {
	ImageID    string `json:"image-id"`
	PatientID  string `json:"patient-id"`
	CreatorID  string `json:"creator-id"`
	Name       string `json:"name"`
	Format     string `json:"format"`
	Width      string `json:"width,omitempty"`
	Height     string `json:"height,omitempty"`
	Size       string `json:"size,omitempty"`
	OriginPath string `json:"origin-path"`
	Status     string `json:"status"`
}

func (h *UploadStatusHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	h.logger.Info("Processing upload status event",
		"EventType", attributes["event_type"],
		"EventID", attributes["event_id"],
	)

	var metadata UploadMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		h.logger.Error("Failed to unmarshal upload metadata", slog.String("error", err.Error()))
		return errors.NewInternalError("Failed to unmarshal upload metadata: %v", err)
	}

	if err := validatateUploadMetadata(&metadata); err != nil {
		h.logger.Error("Invalid upload metadata", slog.String("error", err.Error()))
		return errors.NewInternalError("Invalid upload metadata: %v", err)
	}

	input := h.buildConfirmInput(&metadata)

	if err := h.imageService.ConfirmUpload(ctx, input); err != nil {
		h.logger.Error("Failed to confirm image upload", slog.String("error", err.Error()))
		return errors.NewInternalError("Failed to confirm image upload: %v", err)
	}

	h.logger.Info("Successfully processed upload status event",
		"ImageID", metadata.ImageID,
		"Status", metadata.Status,
	)

	return nil
}

func validatateUploadMetadata(metadata *UploadMetadata) error {
	details := make([]string, 0)
	if metadata.ImageID == "" {
		details = append(details, "ImageID is required")
	}
	if metadata.PatientID == "" {
		details = append(details, "PatientID is required")
	}
	if metadata.CreatorID == "" {
		details = append(details, "CreatorID is required")
	}
	if metadata.Name == "" {
		details = append(details, "Name is required")
	}
	if metadata.Format == "" {
		details = append(details, "Format is required")
	}
	if metadata.OriginPath == "" {
		details = append(details, "OriginPath is required")
	}
	if metadata.Status == "" {
		details = append(details, "Status is required")
	}

	if len(details) > 0 {
		return fmt.Errorf("validation errors: %v", details)
	}
	return nil
}

func (h *UploadStatusHandler) buildConfirmInput(metadata *UploadMetadata) *service.ConfirmUploadInput {
	m := &service.ConfirmUploadInput{
		ImageID:    metadata.ImageID,
		PatientID:  metadata.PatientID,
		CreatorID:  metadata.CreatorID,
		Name:       metadata.Name,
		Format:     metadata.Format,
		OriginPath: metadata.OriginPath,
		Status:     model.ImageStatus(metadata.Status),
	}
	if metadata.Width != "" {
		if width, err := strconv.Atoi(metadata.Width); err == nil {
			m.Width = &width
		}
	}
	if metadata.Height != "" {
		if height, err := strconv.Atoi(metadata.Height); err == nil {
			m.Height = &height
		}
	}
	if metadata.Size != "" {
		if size, err := strconv.ParseInt(metadata.Size, 10, 64); err == nil {
			m.Size = &size
		}
	}
	return m
}
