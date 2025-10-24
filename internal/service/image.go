package service

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
)

type ImageService struct {
	imageRepo   repository.ImageRepository
	patientRepo repository.PatientRepository
	logger      *slog.Logger
	//

}

func NewImageService(
	imageRepo repository.ImageRepository,
	patientRepo repository.PatientRepository,
	logger *slog.Logger,
) *ImageService {
	return &ImageService{
		imageRepo:   imageRepo,
		patientRepo: patientRepo,
		logger:      logger,
	}
}

func (is *ImageService) UploadImage(ctx context.Context, image *model.Image) (*model.Image, error) {
	// validate image request
	// generate record metadata
	// generate signed URL
	// add image record to signed URL to response
	//
}

func (is *ImageService) ConfirmUpload() {
	// subscribe to pub/sub topic
	// validate upload
	// update image record status
	// trigger image processing
}
