package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/models"
)

type WorkspaceService struct {
	repo   *repository.WorkspaceRepository
	logger *slog.Logger
}

func NewWorkspaceService(repo *repository.WorkspaceRepository, logger *slog.Logger) *WorkspaceService {
	return &WorkspaceService{
		repo:   repo,
		logger: logger,
	}
}

func (s *WorkspaceService) validateWorkspaceCreator(ctx context.Context,
	creator_id string) error {

	exists, err := s.repo.GetMainRepository().Exists(ctx, "users", creator_id)

	if err != nil {
		return apperrors.NewInternalError("failed to verify creator existence", err)
	}

	if !exists {
		return apperrors.NewValidationError("creator does not exist for workspace", nil)
	}
	return nil
}

func (s *WorkspaceService) validateWorkspaceNameUnique(ctx context.Context,
	name string) error {

	filter := repository.Filter{
		Field: "name",
		Op:    "=",
		Value: name,
	}

	result, err := s.repo.List(ctx, []repository.Filter{filter}, repository.Pagination{Limit: 1, Offset: 0})

	if err != nil {
		return apperrors.NewInternalError("failed to list workspaces", err)
	}

	if result.Total > 0 {
		return apperrors.NewValidationError(fmt.Sprintf("workspace with name %s already exists", name), nil)
	}

	return nil
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context,
	workspace *models.Workspace) (string, error) {

	if err := s.validateWorkspaceCreator(ctx, workspace.CreatorID); err != nil {
		return "", err
	}

	if err := s.validateWorkspaceNameUnique(ctx, workspace.Name); err != nil {
		return "", err
	}

	workspaceID, err := s.repo.Create(ctx, workspace)

	if err != nil {
		s.logger.Error("failed to create workspace", "error", err)
		return "", apperrors.NewInternalError("failed to create workspace", err)
	}

	s.logger.Info("Created new workspace", "workspaceID", workspaceID)

	return workspaceID, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, workspaceID string) (*models.Workspace, error) {

	workspace, err := s.repo.Read(ctx, workspaceID)

	if err != nil {
		s.logger.Error("failed to get workspace", "error", err)
		return nil, apperrors.NewInternalError("failed to get workspace", err)
	}

	if workspace == nil {
		return nil, apperrors.NewNotFoundError("workspace not found")
	}

	return workspace, nil
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, workspaceID string, updates map[string]interface{}) error {

	var err error
	if name, exists := updates["name"]; exists {
		if err = s.validateWorkspaceNameUnique(ctx, name.(string)); err != nil {
			return err
		}
	}

	if creatorID, exists := updates["creator_id"]; exists {
		if err = s.validateWorkspaceCreator(ctx, creatorID.(string)); err != nil {
			return err
		}
	}

	err = s.repo.Update(ctx, workspaceID, updates)

	if err != nil {
		s.logger.Error("failed to update workspace", "error", err)
		return apperrors.NewInternalError("failed to update workspace", err)
	}
	return nil
}

func (s *WorkspaceService) GetWorkspacesByCreatorID(ctx context.Context, creatorID string,
	pagination repository.Pagination) (*repository.WorkspaceQueryResult, error) {

	filters := []repository.Filter{
		{
			Field: "creator_id",
			Op:    repository.OpEqual,
			Value: creatorID,
		},
	}

	result, err := s.repo.List(ctx, filters, pagination)

	if err != nil {
		s.logger.Error("failed to list workspaces by creator ID", "error", err)
		return nil, apperrors.NewInternalError("failed to list workspaces by creator ID", err)
	}
	return result, nil
}

func (s *WorkspaceService) GetAllWorkspaces(ctx context.Context,
	pagination repository.Pagination) (*repository.WorkspaceQueryResult, error) {

	result, err := s.repo.List(ctx, []repository.Filter{}, pagination)
	if err != nil {
		s.logger.Error("failed to list all workspaces", "error", err)
		return nil, apperrors.NewInternalError("failed to list all workspaces", err)
	}
	return result, nil
}
