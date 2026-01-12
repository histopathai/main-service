package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type ImageService struct {
	imgRepo     port.ImageRepository
	patientRepo port.PatientRepository
	storage     port.ObjectStorage
	bucketName  string
	publisher   port.ImageEventPublisher
	uow         port.UnitOfWorkFactory
}

func NewImageService(
	imageRepo port.ImageRepository,
	uow port.UnitOfWorkFactory,
	storage port.ObjectStorage,
	bucketName string,
	publisher port.ImageEventPublisher,
	patientRepo port.PatientRepository,
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

func (is *ImageService) validateImageInput(ctx context.Context, input *port.UploadImageInput) error {
	_, err := is.patientRepo.Read(ctx, input.Parent.ID)
	if err != nil {
		return err
	}
	return nil
}

func (is *ImageService) UploadImage(ctx context.Context, input *port.UploadImageInput) (*port.SignedURLPayload, error) {

	err := is.validateImageInput(ctx, input)
	if err != nil {
		return nil, err
	}
	uuid := uuid.New().String()
	originpath := fmt.Sprintf("%s-%s", uuid, input.Name)

	entity, err := vobj.NewEntity(vobj.EntityTypeImage, &input.Name, input.CreatorID, input.Parent)
	if err != nil {
		return nil, err
	}

	entity.ID = uuid
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	image := &model.Image{
		Entity:     *entity,
		Format:     input.Format,
		Width:      input.Width,
		Height:     input.Height,
		Size:       input.Size,
		Status:     model.StatusUploaded,
		OriginPath: originpath,
	}

	expr_time := time.Minute * 30

	storagePayload, err := is.storage.GenerateSignedURL(ctx, is.bucketName, port.MethodPut, image, input.ContentType, expr_time)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate signed URL", err)
	}

	return storagePayload, nil
}

func (is *ImageService) ConfirmUpload(ctx context.Context, input *port.ConfirmUploadInput) error {

	entity := &vobj.Entity{
		ID:         *input.ID,
		EntityType: vobj.EntityTypeImage,
		Name:       &input.Name,
		CreatorID:  input.CreatorID,
		Parent:     input.Parent,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	image := &model.Image{
		Entity:     *entity,
		Format:     input.Format,
		Width:      input.Width,
		Height:     input.Height,
		Size:       input.Size,
		Status:     input.Status,
		OriginPath: input.OriginPath,
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
			Field:    constants.ParentIDField,
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
	event := events.NewImageDeletionRequestedEvent(imageID)

	return is.publisher.PublishImageDeletionRequested(ctx, &event)

}

func (is *ImageService) TransferImage(ctx context.Context, imageID string, newPatientID string) error {
	uowerr := is.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		patientRepo := repos.PatientRepo
		_, err := patientRepo.Read(txCtx, newPatientID)
		if err != nil {
			return err
		}

		return repos.ImageRepo.Transfer(txCtx, imageID, newPatientID)
	})

	if uowerr != nil {
		return uowerr
	}

	return nil
}

func (is *ImageService) BatchDeleteImages(ctx context.Context, imageIDs []string) error {

	for _, imageID := range imageIDs {
		event := events.NewImageDeletionRequestedEvent(imageID)

		if err := is.publisher.PublishImageDeletionRequested(ctx, &event); err != nil {
			return errors.NewInternalError("Failed to publish image deletion event for image ID ", err)
		}
	}

	return nil
}

func (is *ImageService) BatchTransferImages(ctx context.Context, imageIDs []string, newPatientID string) error {
	uowerr := is.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		patientRepo := repos.PatientRepo
		_, err := patientRepo.Read(txCtx, newPatientID)
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
