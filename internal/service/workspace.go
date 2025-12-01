package service

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	errors "github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceService struct {
	workspaceRepo  port.WorkspaceRepository
	patientRepo    port.PatientRepository
	patientService port.IPatientService
	uow            port.UnitOfWorkFactory
}

func NewWorkspaceService(
	workspaceRepo port.WorkspaceRepository,
	patientRepo port.PatientRepository,
	patientService port.IPatientService,
	uow port.UnitOfWorkFactory,
) *WorkspaceService {
	return &WorkspaceService{
		workspaceRepo:  workspaceRepo,
		patientRepo:    patientRepo,
		patientService: patientService,
		uow:            uow,
	}
}

func (ws *WorkspaceService) validateWorkspaceInput(ctx context.Context, input *port.CreateWorkspaceInput) error {

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

func (ws *WorkspaceService) CreateNewWorkspace(ctx context.Context, input *port.CreateWorkspaceInput) (*model.Workspace, error) {

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

func (ws *WorkspaceService) UpdateWorkspace(ctx context.Context, id string, input port.UpdateWorkspaceInput) error {

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
	errChan := make(chan error, len(workspaceIDs))
	semaphore := make(chan struct{}, 3)
	for _, workspaceID := range workspaceIDs {
		semaphore <- struct{}{}
		go func(wid string) {
			defer func() { <-semaphore }()
			if err := ws.CascadeDeleteWorkspace(ctx, wid); err != nil {
				errChan <- fmt.Errorf("failed to delete workspace %s: %w", wid, err)
			} else {
				errChan <- nil
			}
		}(workspaceID)
	}

	var firstErr error
	for range workspaceIDs {
		if err := <-errChan; err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (ws *WorkspaceService) CountWorkspaces(ctx context.Context, filters []sharedQuery.Filter) (int64, error) {
	return ws.workspaceRepo.Count(ctx, filters)
}

func (ws *WorkspaceService) CascadeDeleteWorkspace(ctx context.Context, workspaceID string) error {
	_, err := ws.workspaceRepo.Read(ctx, workspaceID)
	if err != nil {
		return err
	}

	const scanBatchSize = 100
	const processBatchSize = 10
	offset := 0
	totalPatients := 0

	for {

		patientsFilter := []sharedQuery.Filter{
			{
				Field:    constants.PatientWorkspaceIDField,
				Operator: sharedQuery.OpEqual,
				Value:    workspaceID,
			},
		}
		pagination := &sharedQuery.Pagination{
			Limit:  scanBatchSize,
			Offset: offset,
		}

		patientResult, err := ws.patientRepo.FindByFilters(ctx, patientsFilter, pagination)
		if err != nil {
			return errors.NewInternalError("failed to find patients", err)
		}

		if len(patientResult.Data) == 0 {
			break
		}

		patientIDs := make([]string, 0, len(patientResult.Data))
		for _, patient := range patientResult.Data {
			patientIDs = append(patientIDs, patient.ID)
		}

		totalPatients += len(patientIDs)

		for i := 0; i < len(patientIDs); i += processBatchSize {
			end := i + processBatchSize
			if end > len(patientIDs) {
				end = len(patientIDs)
			}
			batch := patientIDs[i:end]

			if err := ws.patientService.BatchDelete(ctx, batch); err != nil {
				return errors.NewInternalError("failed to cascade delete patients", err)
			}

		}

		if !patientResult.HasMore {
			break
		}
		offset += scanBatchSize
	}

	if err := ws.workspaceRepo.Delete(ctx, workspaceID); err != nil {
		return errors.NewInternalError("failed to delete workspace", err)
	}

	return nil
}
