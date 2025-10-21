package service

import (
	"context"
	"fmt"
	"log/slog"

	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"

	"github.com/histopathai/models"
)

type CreateWorkspaceRequest struct {
	Name           string `json:"name"`
	OrganType      string `json:"organ_type"`
	Description    string `json:"description"`
	License        string `json:"license"`
	Organization   string `json:"organization"`
	ResourceURL    string `json:"resource_url,omitempty"`
	ReleaseYear    int    `json:"release_year,omitempty"`
	ReleaseVersion string `json:"release_version,omitempty"`
}

type CreateWorkspaceResponse struct {
	WorkspaceID string `json:"workspace_id"`
}

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

func (s *WorkspaceService) validateCreateWorkspaceRequest(req *CreateWorkspaceRequest) error {
	details := make(map[string]interface{})
	if req.Name == "" {
		details["name"] = "name is required"
	}
	if req.OrganType == "" {
		details["organ_type"] = "organ_type is required"
	}
	if req.Description == "" {
		details["description"] = "description is required"
	}
	if req.License == "" {
		details["license"] = "license is required"
	}
	if req.Organization == "" {
		details["organization"] = "organization is required"
	}

	if len(details) > 0 {
		return apperrors.NewBadRequestError("invalid create workspace request", details)
	}

	return nil
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, req *CreateWorkspaceRequest) (*CreateWorkspaceResponse, error) {

	CreatorID, ok := ctx.Value("user_id").(string)
	if !ok || CreatorID == "" {
		return nil, apperrors.NewUnauthorizedError("missing user ID in context")
	}

	if err := s.validateCreateWorkspaceRequest(req); err != nil {
		return nil, err
	}

	workspaces, err := s.ListWorkspaces(ctx, map[string]interface{}{"name": req.Name}, repository.Pagination{Limit: 1, Offset: 0})
	if err != nil {
		s.logger.Error("failed to list workspaces", "error", err)
		return nil, apperrors.NewInternalError("failed to list workspaces", err)
	}
	if len(workspaces) > 0 {
		return nil, apperrors.NewConflictError(fmt.Sprintf("workspace with name %s already exists", req.Name))
	}

	workspace := &models.Workspace{
		ID:               "",
		Name:             req.Name,
		OrganType:        req.OrganType,
		CreatorID:        CreatorID,
		Description:      req.Description,
		License:          req.License,
		Organization:     req.Organization,
		AnnotationTypeID: "", // Default empty
		ResourceURL:      req.ResourceURL,
		ReleaseYear:      req.ReleaseYear,
		ReleaseVersion:   req.ReleaseVersion,
	}

	id, err := s.repo.CreateWorkspace(ctx, workspace)
	if err != nil {
		s.logger.Error("failed to create workspace", "error", err)
		return nil, apperrors.NewInternalError("failed to create workspace", err)
	}

	return &CreateWorkspaceResponse{
		WorkspaceID: id,
	}, nil
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, workspaceID string) (*models.Workspace, error) {
	workspace, err := s.repo.ReadWorkspace(ctx, workspaceID)
	if err != nil {
		s.logger.Error("failed to get workspace", "error", err)
		return nil, apperrors.NewInternalError("failed to get workspace", err)
	}
	return workspace, nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, filters map[string]interface{}, pagination repository.Pagination) ([]*models.Workspace, error) {
	workspaces, err := s.repo.QueryWorkspaces(ctx, filters, pagination)
	if err != nil {
		s.logger.Error("failed to list workspaces", "error", err)
		return nil, apperrors.NewInternalError("failed to list workspaces", err)
	}
	return workspaces, nil
}

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	err := s.repo.DeleteWorkspace(ctx, workspaceID)
	if err != nil {
		s.logger.Error("failed to delete workspace", "error", err)
		return apperrors.NewInternalError("failed to delete workspace", err)
	}
	return nil
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, workspaceID string, updates map[string]interface{}) error {
	err := s.repo.UpdateWorkspace(ctx, workspaceID, updates)
	if err != nil {
		s.logger.Error("failed to update workspace", "error", err)
		return apperrors.NewInternalError("failed to update workspace", err)
	}
	return nil
}
