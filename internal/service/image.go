package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/main-service-refactor/internal/domain/events"
	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/domain/storage"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
)

type ImageService struct {
	imgRepo     repository.ImageRepository
	patientRepo repository.PatientRepository
	storage     storage.ObjectStorage
	bucketName  string
	publisher   events.ImageEventPublisher
}

func NewImageService(
	imageRepo repository.ImageRepository,
	uow repository.UnitOfWorkFactory,
	storage storage.ObjectStorage,
	bucketName string,
	publisher events.ImageEventPublisher,
	patientRepo repository.PatientRepository,
) *ImageService {
	return &ImageService{
		imgRepo:     imageRepo,
		patientRepo: patientRepo,
		storage:     storage,
		bucketName:  bucketName,
		publisher:   publisher,
	}
}

type UploadImageInput struct {
	PatientID   string
	CreatorID   string
	ContentType string
	Name        string
	Format      string
	Width       *int
	Height      *int
	Size        *int64
}

func (is *ImageService) validateImageInput(ctx context.Context, input *UploadImageInput) error {
	patientID := input.PatientID
	_, err := is.patientRepo.Read(ctx, patientID)
	if err != nil {
		return err
	}
	return nil
}

func (is *ImageService) UploadImage(ctx context.Context, input *UploadImageInput) (*string, error) {

	err := is.validateImageInput(ctx, input)
	if err != nil {
		return nil, err
	}
	uuid := uuid.New().String()
	originpath := fmt.Sprintf("gcs://%s/%s-%s", is.bucketName, uuid, input.Name)

	image := &model.Image{
		ID:         uuid,
		PatientID:  input.PatientID,
		CreatorID:  input.CreatorID,
		Name:       input.Name,
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
		return nil, err
	}

	return &url, nil
}

type ConfirmUploadInput struct {
	ImageID    string
	PatientID  string
	CreatorID  string
	Name       string
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
		CreatorID:  input.CreatorID,
		Name:       input.Name,
		Format:     input.Format,
		Width:      input.Width,
		Height:     input.Height,
		Size:       input.Size,
		Status:     input.Status,
		OriginPath: input.OriginPath,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	createdImage, err := is.imgRepo.Create(ctx, image)
	if err != nil {
		return err
	}
	processingEvent := events.NewImageProcessingRequestedEvent(
		createdImage.ID,
		createdImage.OriginPath,
	)

	if err := is.publisher.PublishImageProcessingRequested(ctx, &processingEvent); err != nil {
		return errors.NewInternalError("Failed to publish image processing event: %v", err)
	}

	return nil
}

func (is *ImageService) GetImageByID(ctx context.Context, imageID string) (*model.Image, error) {
	image, err := is.imgRepo.Read(ctx, imageID)
	if err != nil {
		return nil, err
	}
	return image, nil
}

func (is *ImageService) ListImageByPatientID(ctx context.Context, patientID string, pagination *sharedQuery.Pagination) (*sharedQuery.Result[*model.Image], error) {

	filters := []sharedQuery.Filter{
		{
			Field:    constants.ImagePatientIDField,
			Operator: sharedQuery.OpEqual,
			Value:    patientID,
		},
	}

	images, err := is.imgRepo.FindByFilters(ctx, filters, pagination)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func (is *ImageService) DeleteImageByID(ctx context.Context, imageID string) error {
	return is.imgRepo.Delete(ctx, imageID)
}
