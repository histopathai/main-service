package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/domain/storage"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
)

type ImageService struct {
	imageRepo   repository.ImageRepository
	patientRepo repository.PatientRepository
	storage     storage.ObjectStorage
	bucketName  string
	logger      *slog.Logger
}

func NewImageService(
	imageRepo repository.ImageRepository,
	patientRepo repository.PatientRepository,
	storage storage.ObjectStorage,
	bucketName string,
	logger *slog.Logger,
) *ImageService {
	return &ImageService{
		imageRepo:   imageRepo,
		patientRepo: patientRepo,
		storage:     storage,
		bucketName:  bucketName,
		logger:      logger,
	}
}

type UploadImageInput struct {
	PatientID   string
	CreatorID   string
	ContentType string
	FileName    string
	Format      string
	Width       *int
	Height      *int
	Size        *int64
}

func (is *ImageService) validateImageInput(ctx context.Context, input *UploadImageInput) error {
	patient, err := is.patientRepo.GetByID(ctx, input.PatientID)
	if err != nil {
		return errors.NewInternalError("Failed to fetch patient: %v", err)
	}
	if patient == nil {
		return errors.NewNotFoundError(fmt.Sprintf("Patient not found with ID: %s", input.PatientID))
	}
	return nil
}

func (is *ImageService) UploadImage(ctx context.Context, input *UploadImageInput) (*string, error) {

	err := is.validateImageInput(ctx, input)
	if err != nil {
		return nil, err
	}
	uuid := uuid.New().String()
	originpath := fmt.Sprintf("gcs://%s/%s-%s", is.bucketName, uuid, input.FileName)

	image := &model.Image{
		ID:         uuid,
		PatientID:  input.PatientID,
		CreatorID:  input.CreatorID,
		FileName:   input.FileName,
		Format:     input.Format,
		Width:      input.Width,
		Height:     input.Height,
		Size:       input.Size,
		Status:     model.StatusUploaded,
		OriginPath: originpath,
	}

	expr_time := time.Minute * 30
	url, err := is.storage.GenerateSignedURL(ctx, is.bucketName, storage.MethodPut, image, input.ContentType, expr_time)
	if err != nil {
		return nil, errors.NewInternalError("Failed to generate signed URL: %v", err)
	}
	if url == "" {
		return nil, errors.NewInternalError("Generated signed URL is empty", err)
	}
	return &url, nil
}

type ConfirmUploadInput struct {
	ImageID    string
	PatientID  string
	CreattorID string
	FileName   string
	Format     string
	Width      *int
	Height     *int
	Size       *int64
	Status     model.ImageStatus
	OriginPath string
}

func (is *ImageService) ConfirmUpload(ctx context.Context, input *ConfirmUploadInput) error {
	image := &model.Image{
		ID:         input.ImageID,
		PatientID:  input.PatientID,
		CreatorID:  input.CreattorID,
		FileName:   input.FileName,
		Format:     input.Format,
		Width:      input.Width,
		Height:     input.Height,
		Size:       input.Size,
		Status:     input.Status,
		OriginPath: input.OriginPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := is.imageRepo.Create(ctx, image)
	if err != nil {
		return errors.NewInternalError("Failed to create image record: %v", err)
	}

	return nil
}
