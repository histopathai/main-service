package service

import (
	"context"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	errors "github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
)

type WorkspaceService struct {
	workspaceRepo repository.WorkspaceRepository
	uow           repository.UnitOfWorkFactory
}

func NewWorkspaceService(
	workspaceRepo repository.WorkspaceRepository,
	uow repository.UnitOfWorkFactory,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		uow:           uow,
	}
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

func (ws *WorkspaceService) validateWorkspaceInput(ctx context.Context, input *CreateWorkspaceInput) error {

	filter := []sharedQuery.Filter{
		{
			Field:    "name",
			Operator: sharedQuery.OpEqual,
			Value:    input.Name,
		},
	}

	pagination := &sharedQuery.Pagination{
		Limit:  1,
		Offset: 0,
	}

	existingWorkspaces, err := ws.workspaceRepo.FindByFilters(ctx, filter, pagination)
	if err != nil {
		return err
	}

	if len(existingWorkspaces.Data) > 0 {
		details := map[string]interface{}{"name": "Workspace with this name already exists."}
		return errors.NewConflictError("workspace name already exists", details)
	}

	return nil

}

func (ws *WorkspaceService) CreateNewWorkspace(ctx context.Context, input *CreateWorkspaceInput) (*model.Workspace, error) {

	if err := ws.validateWorkspaceInput(ctx, input); err != nil {
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

	//check ID existence
	_, err := ws.workspaceRepo.Read(ctx, id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})
	if input.Name != nil {
		updates[constants.WorkspaceNameField] = *input.Name
		validateerr := ws.validateWorkspaceInput(ctx, &CreateWorkspaceInput{Name: *input.Name})
		if validateerr != nil {
			return validateerr
		}
	}
	if input.OrganType != nil {
		updates[constants.WorkspaceOrganTypeField] = *input.OrganType
	}
	if input.Organization != nil {
		updates[constants.WorkspaceOrganizationField] = *input.Organization
	}
	if input.Description != nil {
		updates[constants.WorkspaceDescField] = *input.Description
	}
	if input.License != nil {
		updates[constants.WorkspaceLicenseField] = *input.License
	}
	if input.ResourceURL != nil {
		updates[constants.WorkspaceResourceURLField] = *input.ResourceURL
	}
	if input.ReleaseYear != nil {
		year := *input.ReleaseYear
		if year < 1900 || year > 2100 {
			details := map[string]interface{}{"release_year": "Release year must be between 1900 and 2100."}
			return errors.NewValidationError("invalid release year", details)
		}
		updates[constants.WorkspaceReleaseYearField] = *input.ReleaseYear
	}
	if input.AnnotationTypeID != nil {
		updates[constants.WorkspaceAnnotationTypeIDField] = *input.AnnotationTypeID
	}

	if len(updates) == 0 {
		return nil // No updates to apply
	}

	err = ws.workspaceRepo.Update(ctx, id, updates)
	if err != nil {
		return errors.NewInternalError("failed to update workspace", err)
	}

	return nil
}

func (ws *WorkspaceService) GetWorkspaceByID(ctx context.Context, id string) (*model.Workspace, error) {
	workspace, err := ws.workspaceRepo.Read(ctx, id)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve workspace", err)
	}
	if workspace == nil {
		details := map[string]interface{}{"workspace_id": "Workspace not found."}
		return nil, errors.NewValidationError("workspace not found", details)
	}
	return workspace, nil
}

func (ws *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {

	uowerr := ws.uow.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {
		patientRepo := repos.PatientRepo

		pagination := &sharedQuery.Pagination{Limit: 1, Offset: 0}

		patientFilter := []sharedQuery.Filter{
			{
				Field:    constants.PatientWorkspaceIDField,
				Operator: sharedQuery.OpEqual,
				Value:    id,
			},
		}
		patientResult, err := patientRepo.FindByFilters(txCtx, patientFilter, pagination)
		if err != nil {
			return err
		}
		if len(patientResult.Data) > 0 {
			details := map[string]interface{}{"workspace_id": "Workspace is in use by one or more patients."}
			return errors.NewConflictError("workspace in use", details)
		}

		if err := repos.WorkspaceRepo.Delete(txCtx, id); err != nil {
			return err
		}
		return nil
	})
	if uowerr != nil {
		return uowerr
	}
	return nil
}

func (ws *WorkspaceService) ListWorkspaces(ctx context.Context, pagination *sharedQuery.Pagination) (*sharedQuery.Result[*model.Workspace], error) {
	workspaces, err := ws.workspaceRepo.FindByFilters(ctx, []sharedQuery.Filter{}, pagination)
	if err != nil {
		return nil, errors.NewInternalError("failed to list workspaces", err)
	}
	return workspaces, nil
}
