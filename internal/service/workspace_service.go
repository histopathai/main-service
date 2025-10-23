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

func (s *WorkspaceService) CreateWorkspace(ctx context.Context,
	workspace *models.Workspace) (string, error) {

	filter := repository.Filter{
		Field: "name",
		Op:    "=",
		Value: workspace.Name,
	}

	result, err := s.repo.List(ctx, []repository.Filter{filter}, repository.Pagination{Limit: 1, Offset: 0})
	if err != nil {
		s.logger.Error("failed to list workspaces", "error", err)
		return "", apperrors.NewInternalError("failed to list workspaces", err)
	}
	if result.Total > 0 {
		return "", apperrors.NewValidationError(fmt.Sprintf("workspace with name %s already exists", result.Workspaces[0].Name), nil)
	}

	workspaceID, err := s.repo.Create(ctx, workspace)
	if err != nil {
		s.logger.Error("failed to create workspace", "error", err)
		return "", apperrors.NewInternalError("failed to create workspace", err)
	}

	s.logger.Info("Created new workspace", "workspaceID", workspaceID)
	return workspaceID, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context,
	workspaceID string) (*models.Workspace, error) {
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

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context,
	workspaceID string, updates map[string]interface{}) error {
	_, ok := updates["name"]

	if ok {
		filter := repository.Filter{
			Field: "name",
			Op:    "=",
			Value: updates["name"],
		}

		result, err := s.repo.List(ctx, []repository.Filter{filter}, repository.Pagination{Limit: 1, Offset: 0})
		if err != nil {
			s.logger.Error("failed to list workspaces", "error", err)
			return apperrors.NewInternalError("failed to list workspaces", err)
		}
		if result.Total > 0 {
			return apperrors.NewValidationError(fmt.Sprintf("workspace with name %s already exists", updates["name"]), nil)
		}
	}

	_, ok = updates["creator_id"]
	if ok {
		exists, err := s.repo.GetMainRepository().Exists(ctx, "users", updates["creator_id"].(string))
		if err != nil {
			s.logger.Error("failed to verify creator existence", "error", err)
			return apperrors.NewInternalError("failed to verify creator existence", err)
		}
		if !exists {
			return apperrors.NewValidationError("creator does not exist for workspace", nil)
		}
	}

	_, ok = updates["annotation_type_id"]
	if ok {
		exists, err := s.repo.GetMainRepository().Exists(ctx, repository.AnnotationTypesCollection, updates["annotation_type_id"].(string))
		if err != nil {
			s.logger.Error("failed to verify annotation type existence", "error", err)
			return apperrors.NewInternalError("failed to verify annotation type existence", err)
		}
		if !exists {
			return apperrors.NewValidationError("annotation type does not exist for workspace", nil)
		}
	}

	err := s.repo.Update(ctx, workspaceID, updates)
	if err != nil {
		s.logger.Error("failed to update workspace", "error", err)
		return apperrors.NewInternalError("failed to update workspace", err)
	}
	return nil
}

func (s *WorkspaceService) GetWorkspacesByCreatorID(ctx context.Context,
	creatorID string, pagination repository.Pagination) (*repository.WorkspaceQueryResult, error) {
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
