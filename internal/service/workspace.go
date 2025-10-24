package service

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	errors "github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
)

type WorkspaceService struct {
	workspaceRepo      repository.WorkspaceRepository
	annotationTypeRepo repository.AnnotationTypeRepository
	logger             *slog.Logger
}

func NewWorkspaceService(
	workspaceRepo repository.WorkspaceRepository,
	annotationTypeRepo repository.AnnotationTypeRepository,
	logger *slog.Logger,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo:      workspaceRepo,
		annotationTypeRepo: annotationTypeRepo,
		logger:             logger,
	}
}

func (ws *WorkspaceService) validateWorkspaceCreation(ctx context.Context, workspaceName string, annotationTypeID *string) error {

	pagination := &sharedQuery.Pagination{
		Limit:  1,
		Offset: 0,
	}

	filters := []sharedQuery.Filter{
		{
			Field:    "Name",
			Operator: sharedQuery.OpEqual,
			Value:    workspaceName,
		},
	}

	existingWorkspaces, err := ws.workspaceRepo.GetByCriteria(ctx, filters, pagination)
	if err != nil {
		ws.logger.Error("Failed to query existing workspaces during validation", "error", err, "workspaceName", workspaceName)
		details := map[string]interface{}{"name": "Failed to validate workspace name."}
		return errors.NewValidationError("failed to validate workspace name", details)
	}

	if existingWorkspaces != nil && len(existingWorkspaces.Data) > 0 {
		details := map[string]interface{}{"name": "Workspace name already exists."}
		return errors.NewValidationError("workspace name already exists", details)
	}
	if annotationTypeID != nil {
		_, err := ws.annotationTypeRepo.GetByID(ctx, *annotationTypeID)
		if err != nil {
			ws.logger.Error("Failed to read annotation type during workspace validation", "error", err, "annotationTypeID", *annotationTypeID)
			details := map[string]interface{}{"annotation_type_id": "Failed to read annotation type."}
			return errors.NewValidationError("failed to read annotation type", details)
		}
	}

	return nil
}

type CreateWorkspaceInput struct {
	CreatorID        string
	Name             string
	OrganType        string
	AnnotationTypeID *string
	Organization     string
	Description      string
	License          string
	ResourceURL      *string
	ReleaseYear      *int
}

func (ws *WorkspaceService) CreateWorkspace(ctx context.Context, input CreateWorkspaceInput) (*model.Workspace, error) {

	if err := ws.validateWorkspaceCreation(ctx, input.Name, input.AnnotationTypeID); err != nil {
		return nil, err
	}

	newWorkspace := &model.Workspace{
		CreatorID:        input.CreatorID,
		Name:             input.Name,
		OrganType:        input.OrganType,
		AnnotationTypeID: input.AnnotationTypeID,
		Organization:     input.Organization,
		Description:      input.Description,
		License:          input.License,
		ResourceURL:      input.ResourceURL,
		ReleaseYear:      input.ReleaseYear,
	}

	created, err := ws.workspaceRepo.Create(ctx, newWorkspace)
	if err != nil {
		ws.logger.Error("Failed to create workspace", "error", err, "workspaceName", input.Name)
		return nil, errors.NewInternalError("failed to create workspace", err)
	}

	return created, nil
}

type UpdateWorkspaceInput struct {
	Name             *string
	OrganType        *string
	Organization     *string
	Description      *string
	License          *string
	ResourceURL      *string
	ReleaseYear      *int
	AnnotationTypeID *string
}

func (ws *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, input UpdateWorkspaceInput) error {
	updates := make(map[string]interface{})

	if input.Name != nil {
		updates["Name"] = *input.Name
	}
	if input.OrganType != nil {
		updates["OrganType"] = *input.OrganType
	}
	if input.Organization != nil {
		updates["Organization"] = *input.Organization
	}
	if input.Description != nil {
		updates["Description"] = *input.Description
	}
	if input.License != nil {
		updates["License"] = *input.License
	}
	if input.ResourceURL != nil {
		updates["ResourceURL"] = *input.ResourceURL
	}
	if input.ReleaseYear != nil {
		year := *input.ReleaseYear
		if year < 1900 || year > 2100 {
			details := map[string]interface{}{"release_year": "Release year must be between 1900 and 2100."}
			return errors.NewValidationError("invalid release year", details)
		}
		updates["ReleaseYear"] = *input.ReleaseYear
	}
	if input.AnnotationTypeID != nil {
		updates["AnnotationTypeID"] = *input.AnnotationTypeID
	}

	if len(updates) == 0 {
		return nil // No updates to apply
	}

	err := ws.workspaceRepo.Update(ctx, id, updates)
	if err != nil {
		ws.logger.Error("Failed to update workspace", "error", err, "workspaceID", id)
		return errors.NewInternalError("failed to update workspace", err)
	}

	return nil
}

func (ws *WorkspaceService) GetWorkspaceByID(ctx context.Context, id string) (*model.Workspace, error) {
	workspace, err := ws.workspaceRepo.GetByID(ctx, id)
	if err != nil {
		ws.logger.Error("Failed to retrieve workspace", "error", err, "workspaceID", id)
		return nil, errors.NewInternalError("failed to retrieve workspace", err)
	}
	if workspace == nil {
		details := map[string]interface{}{"workspace_id": "Workspace not found."}
		return nil, errors.NewValidationError("workspace not found", details)
	}
	return workspace, nil
}

func (ws *WorkspaceService) GetWorkspacesByCreatorID(ctx context.Context, creatorID string, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Workspace], error) {
	workspaces, err := ws.workspaceRepo.GetByCreatorID(ctx, creatorID)
	if err != nil {
		ws.logger.Error("Failed to retrieve workspaces by creator ID", "error", err, "creatorID", creatorID)
		return nil, errors.NewInternalError("failed to retrieve workspaces", err)
	}
	return workspaces, nil
}

func (ws *WorkspaceService) GetWorkspacesByOrganType(ctx context.Context, organType string, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Workspace], error) {
	workspaces, err := ws.workspaceRepo.GetByeOrganType(ctx, organType)
	if err != nil {
		ws.logger.Error("Failed to retrieve workspaces by organ type", "error", err, "organType", organType)
		return nil, errors.NewInternalError("failed to retrieve workspaces", err)
	}
	return workspaces, nil
}

func (ws *WorkspaceService) GetAllWorkspaces(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Workspace], error) {
	workspaces, err := ws.workspaceRepo.GetByCriteria(ctx, []sharedQuery.Filter{}, paginationOpts)
	if err != nil {
		ws.logger.Error("Failed to retrieve all workspaces", "error", err)
		return nil, errors.NewInternalError("failed to retrieve workspaces", err)
	}
	return workspaces, nil
}
