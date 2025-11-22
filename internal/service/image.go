package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/domain/storage"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type ImageService struct {
	imgRepo     repository.ImageRepository
	patientRepo repository.PatientRepository
	storage     storage.ObjectStorage
	bucketName  string
	publisher   events.ImageEventPublisher
	uow         repository.UnitOfWorkFactory
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
		uow:         uow,
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

func (is *ImageService) UploadImage(ctx context.Context, input *UploadImageInput) (*storage.SignedURLPayload, error) {

	err := is.validateImageInput(ctx, input)
	if err != nil {
		return nil, err
	}
	uuid := uuid.New().String()
	originpath := fmt.Sprintf("%s-%s", uuid, input.Name)

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

	storagePayload, err := is.storage.GenerateSignedURL(ctx, is.bucketName, storage.MethodPut, image, input.ContentType, expr_time)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate signed URL", err)
	}

	return storagePayload, nil
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

func (is *ImageService) BatchDeleteImages(ctx context.Context, imageIDs []string) error {
	return is.imgRepo.BatchDelete(ctx, imageIDs)
}

func (is *ImageService) BatchTransferImages(ctx context.Context, imageIDs []string, newPatientID string) error {
	uowerr := is.uow.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {
		_, err := repos.PatientRepo.Read(txCtx, newPatientID)
		if err != nil {
			return err
		}

		return repos.ImageRepo.BatchTransfer(txCtx, imageIDs, newPatientID)
	})

	if uowerr != nil {
		return uowerr
	}

	return nil
}

func (is *ImageService) CountImages(ctx context.Context, filters []sharedQuery.Filter) (int64, error) {
	return is.imgRepo.Count(ctx, filters)
}
