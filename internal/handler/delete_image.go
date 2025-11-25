package handler

import (
	"log/slog"

	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/pkg/config"
)

type ImageDeleteHandler struct {
	storage *port.ObjectStorage
	imgRepo *port.ImageRepository
	logger  *slog.Logger
	cfg     *config.GCPConfig
}

func NewImageDeleteHandler(
	storage *port.ObjectStorage,
	imageRepo *port.ImageRepository,
	logger *slog.Logger,
	cfg *config.GCPConfig,
) *ImageDeleteHandler {
	return &ImageDeleteHandler{
		storage: storage,
		imgRepo: imageRepo,
		logger:  logger,
		cfg:     cfg,
	}
}
