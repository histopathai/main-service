package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	errors "github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceService struct {
	workspaceRepo port.WorkspaceRepository
	uow           port.UnitOfWorkFactory
}

func NewWorkspaceService(
	workspaceRepo port.WorkspaceRepository,
	uow port.UnitOfWorkFactory,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo: workspaceRepo,
		uow:           uow,
	}
}

func (ws *WorkspaceService) validateWorkspaceInput(ctx context.Context, input *CreateWorkspaceInput) error {

	existing, err := ws.workspaceRepo.FindByName(ctx, input.Name)
	if err != nil {
		return errors.NewInternalError("failed to check existing workspace name", err)
	}
	if existing != nil {
		return errors.NewConflictError("workspace with the same name already exists", map[string]interface{}{"name": "Workspace name must be unique"})
	}

	if input.ReleaseYear != nil {
		year := *input.ReleaseYear
		if year < 1900 || year > 2100 {
			details := map[string]interface{}{"release_year": "Release year must be between 1900 and 2100."}
			return errors.NewValidationError("invalid release year", details)
		}
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

func (ws *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, input UpdateWorkspaceInput) error {

	//check ID existence
	_, err := ws.workspaceRepo.Read(ctx, id)
	if err != nil {
		return err
	}

	updates := make(map[string]interface{})

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

func (ws *WorkspaceService) ListWorkspaces(ctx context.Context, pagination *sharedQuery.Pagination) (*sharedQuery.Result[*model.Workspace], error) {
	workspaces, err := ws.workspaceRepo.FindByFilters(ctx, []sharedQuery.Filter{}, pagination)
	if err != nil {
		return nil, errors.NewInternalError("failed to list workspaces", err)
	}
	return workspaces, nil
}

func (ws *WorkspaceService) DeleteWorkspace(ctx context.Context, id string) error {

	uowerr := ws.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
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

func (ws *WorkspaceService) BatchDeleteWorkspaces(ctx context.Context, workspaceIDs []string) error {
	for _, workspaceID := range workspaceIDs {
		if err := ws.CascadeDeleteWorkspace(ctx, workspaceID); err != nil {
			return errors.NewInternalError("failed to delete workspace: "+workspaceID, err)
		}
	}
	return nil
}

func (ws *WorkspaceService) CountWorkspaces(ctx context.Context, filters []sharedQuery.Filter) (int64, error) {
	return ws.workspaceRepo.Count(ctx, filters)
}

func (ws *WorkspaceService) CascadeDeleteWorkspace(ctx context.Context, workspaceID string) error {

	_, err := ws.workspaceRepo.Read(ctx, workspaceID)
	if err != nil {
		return errors.NewInternalError("failed to read workspace", err)
	}

	return ws.uow.WithTx(ctx, func(txctx context.Context, repos *port.Repositories) error {

		patientIDs, err := collectionPatientIds(txctx, repos.PatientRepo, workspaceID)
		if err != nil {
			return err
		}

		allImageIds := make([]string, 0)
		allAnnotationIds := make([]string, 0)

		for _, patientID := range patientIDs {
			imageIds, annotationIds, err := collectImageAndAnnotationIDs(txctx, repos, patientID)
			if err != nil {
				return err
			}
			allImageIds = append(allImageIds, imageIds...)
			allAnnotationIds = append(allAnnotationIds, annotationIds...)
		}

		// Delete All Annotations
		max_batch_size := 500
		if len(allAnnotationIds) > 0 {
			for i := 0; i < len(allAnnotationIds); i += max_batch_size {
				end := i + max_batch_size
				if end > len(allAnnotationIds) {
					end = len(allAnnotationIds)
				}
				batch := allAnnotationIds[i:end]
				if err := repos.AnnotationRepo.BatchDelete(txctx, batch); err != nil {
					return err
				}
			}
		}

		// Delete All Images
		if len(allImageIds) > 0 {
			for i := 0; i < len(allImageIds); i += max_batch_size {
				end := i + max_batch_size
				if end > len(allImageIds) {
					end = len(allImageIds)
				}
				batch := allImageIds[i:end]
				if err := repos.ImageRepo.BatchDelete(txctx, batch); err != nil {
					return err
				}
			}
		}

		// Delete All Patients
		if len(patientIDs) > 0 {
			for i := 0; i < len(patientIDs); i += max_batch_size {
				end := i + max_batch_size
				if end > len(patientIDs) {
					end = len(patientIDs)
				}
				batch := patientIDs[i:end]
				if err := repos.PatientRepo.BatchDelete(txctx, batch); err != nil {
					return err
				}
			}
		}

		// Finally Delete Workspace
		if err := repos.WorkspaceRepo.Delete(txctx, workspaceID); err != nil {
			return err
		}

		return nil
	})
}

func collectionPatientIds(ctx context.Context, patientRepo port.PatientRepository, workspaceID string) ([]string, error) {
	patientIDs := make([]string, 0)
	offset := 0
	limit := 100

	for {
		filters := []sharedQuery.Filter{
			{
				Field:    constants.PatientWorkspaceIDField,
				Operator: sharedQuery.OpEqual,
				Value:    workspaceID,
			},
		}
		pagination := &sharedQuery.Pagination{
			Limit:  limit,
			Offset: offset,
		}
		result, err := patientRepo.FindByFilters(ctx, filters, pagination)
		if err != nil {
			return nil, err
		}
		for _, patient := range result.Data {
			patientIDs = append(patientIDs, patient.ID)
		}
		if !result.HasMore {
			break
		}
		offset += limit
	}
	return patientIDs, nil
}

func collectImageAndAnnotationIDs(ctx context.Context, repos *port.Repositories, patientID string) ([]string, []string, error) {
	imageIDs := make([]string, 0)
	annotationIDs := make([]string, 0)
	offset := 0
	limit := 100

	for {
		filters := []sharedQuery.Filter{
			{
				Field:    constants.ImagePatientIDField,
				Operator: sharedQuery.OpEqual,
				Value:    patientID,
			},
		}
		pagination := &sharedQuery.Pagination{
			Limit:  limit,
			Offset: offset,
		}
		result, err := repos.ImageRepo.FindByFilters(ctx, filters, pagination)
		if err != nil {
			return nil, nil, err
		}
		for _, image := range result.Data {
			imageIDs = append(imageIDs, image.ID)

			// Collect Annotations for this image
			annoOffset := 0
			for {
				annoFilters := []sharedQuery.Filter{
					{
						Field:    constants.AnnotationImageIDField,
						Operator: sharedQuery.OpEqual,
						Value:    image.ID,
					},
				}
				annoPagination := &sharedQuery.Pagination{
					Limit:  limit,
					Offset: annoOffset,
				}
				annoResult, err := repos.AnnotationRepo.FindByFilters(ctx, annoFilters, annoPagination)
				if err != nil {
					return nil, nil, err
				}
				for _, anno := range annoResult.Data {
					annotationIDs = append(annotationIDs, anno.ID)
				}
				if !annoResult.HasMore {
					break
				}
				annoOffset += limit
			}
		}
		if !result.HasMore {
			break
		}
		offset += limit
	}
	return imageIDs, annotationIDs, nil
}
