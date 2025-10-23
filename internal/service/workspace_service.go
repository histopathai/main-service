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

type ListWorkSpaceRequest struct {
	OrganType    string `form:"organ_type" filter:"organ_type"`
	License      string `form:"license" filter:"license"`
	Organization string `form:"organization" filter:"organization"`
	CreatorID    string `form:"creator_id" filter:"creator_id"`
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

	workspaces, err := s.repo.GetWorkspaceByName(ctx, req.Name, repository.Pagination{Limit: 1, Offset: 0})

	if err != nil {
		s.logger.Error("failed to list workspaces", "error", err)
		return nil, apperrors.NewInternalError("failed to list workspaces", err)
	}
	if workspaces != nil && len(workspaces.Workspaces) > 0 {
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

func (s *WorkspaceService) DeleteWorkspace(ctx context.Context, workspaceID string) error {
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

func (s *WorkspaceService) WorkspaceExists(ctx context.Context, workspaceID string) (bool, error) {
	exists, err := s.repo.Exists(ctx, workspaceID)
	if err != nil {
		s.logger.Error("failed to check if workspace exists", "error", err)
		return false, apperrors.NewInternalError("failed to check if workspace exists", err)
	}
	return exists, nil
}

func (s *WorkspaceService) GetWorkspacesByCreatorID(ctx context.Context, creatorID string, pagination repository.Pagination) (*repository.WorkspaceQueryResult, error) {

	workspaces, err := s.repo.GetWorkspaceByCreator(ctx, creatorID, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list workspaces by creator ID", err)
	}
	if workspaces == nil {
		workspaces = &repository.WorkspaceQueryResult{
			Workspaces: []*models.Workspace{},
			Total:      0,
			Limit:      pagination.Limit,
			Offset:     pagination.Offset,
			HasMore:    false,
		}
	}
	return workspaces, nil
}

func (s *WorkspaceService) GetWorkspaceByOrganType(ctx context.Context, organType string, pagination repository.Pagination) (*repository.WorkspaceQueryResult, error) {

	workspaces, err := s.repo.GetWorkspaceByOrganType(ctx, organType, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list workspaces by organ type", err)
	}
	if workspaces == nil {
		workspaces = &repository.WorkspaceQueryResult{
			Workspaces: []*models.Workspace{},
			Total:      0,
			Limit:      pagination.Limit,
			Offset:     pagination.Offset,
			HasMore:    false,
		}
	}
	return workspaces, nil
}

func (s *WorkspaceService) GetAllWorkspaces(ctx context.Context, pagination repository.Pagination) (*repository.WorkspaceQueryResult, error) {

	workspaces, err := s.repo.GetAllWorkspaces(ctx, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list all workspaces", err)
	}
	if workspaces == nil {
		workspaces = &repository.WorkspaceQueryResult{
			Workspaces: []*models.Workspace{},
			Total:      0,
			Limit:      pagination.Limit,
			Offset:     pagination.Offset,
			HasMore:    false,
		}
	}
	return workspaces, nil
}

func (s *WorkspaceService) SearchWorkspaces(ctx context.Context, request *ListWorkSpaceRequest, pagination repository.Pagination) (*repository.WorkspaceQueryResult, error) {

	filters := make([]repository.Filter, 0)

	if request.OrganType != "" {
		filters = append(filters, repository.Filter{
			Field: "organ_type",
			Op:    repository.OpEqual,
			Value: request.OrganType,
		})
	}

	if request.License != "" {
		filters = append(filters, repository.Filter{
			Field: "license",
			Op:    repository.OpEqual,
			Value: request.License,
		})
	}

	if request.Organization != "" {
		filters = append(filters, repository.Filter{
			Field: "organization",
			Op:    repository.OpEqual,
			Value: request.Organization,
		})
	}

	if request.CreatorID != "" {
		filters = append(filters, repository.Filter{
			Field: "creator_id",
			Op:    repository.OpEqual,
			Value: request.CreatorID,
		})
	}

	workspaces, err := s.repo.ListWorkspaces(ctx, filters, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to search workspaces", err)
	}
	if workspaces == nil {
		workspaces = &repository.WorkspaceQueryResult{
			Workspaces: []*models.Workspace{},
			Total:      0,
			Limit:      pagination.Limit,
			Offset:     pagination.Offset,
			HasMore:    false,
		}
	}
	return workspaces, nil
}
