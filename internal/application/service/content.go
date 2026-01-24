package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

const READEXPIRY_DURATION = time.Minute * 30
const DELETEEXPIRY_DURATION = time.Minute * 15
const UPLOADEXPIRY_DURATION = time.Hour * 1

type ContentService struct {
	*Service[*model.Content]
	objectStorage  port.Storage
	contentUseCase *usecase.ContentUseCase
	logger         slog.Logger
}

func NewContentService(
	repo port.Repository[*model.Content],
	uowFactory port.UnitOfWorkFactory,
	originStorage port.Storage,
	logger slog.Logger,
) *ContentService {
	return &ContentService{
		Service: &Service[*model.Content]{
			repo:       repo,
			uowFactory: uowFactory,
		},
		objectStorage:  originStorage,
		contentUseCase: usecase.NewContentUseCase(repo, uowFactory),
		logger:         logger,
	}
}

// This function handles the upload of content based on the provided command.
// It Performs on just derivated proccesed content upload
func (s *ContentService) Upload(ctx context.Context, cmd any) (*port.PresignedURLPayload, error) {

	// Type assertion
	uploadCmd, ok := cmd.(command.CreateContentCommand)
	if !ok {
		return nil, errors.NewInternalError("invalid command type for uploading content", nil)
	}

	entity, err := uploadCmd.ToEntity()
	if err != nil {
		return nil, err
	}

	imageRepo := s.uowFactory.GetImageRepo()
	currentImage, err := imageRepo.Read(ctx, entity.Parent.ID)
	if err != nil {
		return nil, errors.NewInternalError("failed to read parent image for content upload", err)
	}

	entity.Provider = s.objectStorage.Provider()

	entity.Path = fmt.Sprintf("%s/%s/%s", currentImage.Parent.ID, currentImage.ID, entity.Name)

	presignedURLPayload, err := s.objectStorage.GenerateSignedURL(ctx, port.MethodPut, *entity, UPLOADEXPIRY_DURATION)

	if err != nil {
		return nil, errors.NewInternalError("failed to generate signed URL for content upload", err)
	}

	return presignedURLPayload, nil
}

func (s *ContentService) Get(ctx context.Context, cmd command.ReadCommand) (*port.PresignedURLPayload, error) {
	content, err := s.Service.Get(ctx, cmd)
	if err != nil {
		return nil, err
	}

	// Check if content exists in storage
	exists, err := s.objectStorage.Exists(ctx, *content)
	if err != nil {
		return nil, errors.NewInternalError("failed to check content existence in storage", err)
	}
	if !exists {
		return nil, errors.NewNotFoundError("content not found in storage")
	}

	presignedURLPayload, err := s.objectStorage.GenerateSignedURL(ctx, port.MethodGet, *content, READEXPIRY_DURATION)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate signed URL for content retrieval", err)
	}

	return presignedURLPayload, nil
}

func (s *ContentService) Delete(ctx context.Context, cmd command.DeleteCommand) (*port.PresignedURLPayload, error) {
	return nil, errors.NewForbiddenError("content deletion via signed URL is not allowed")
}

func (s *ContentService) DeleteMany(ctx context.Context, cmd command.DeleteCommands) error {
	return errors.NewForbiddenError("content deletion via signed URL is not allowed")
}
