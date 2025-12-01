package service

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	errors "github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type PatientService struct {
	patientRepo         port.PatientRepository
	workspaceRepo       port.WorkspaceRepository
	imageRepo           port.ImageRepository
	annotationRepo      port.AnnotationRepository
	imageEventPublisher port.ImageEventPublisher
	uow                 port.UnitOfWorkFactory
}

func NewPatientService(
	patientRepo port.PatientRepository,
	workspaceRepo port.WorkspaceRepository,
	imageRepo port.ImageRepository,
	annotationRepo port.AnnotationRepository,
	imageEventPublisher port.ImageEventPublisher,
	uow port.UnitOfWorkFactory,
) *PatientService {
	return &PatientService{
		patientRepo:         patientRepo,
		workspaceRepo:       workspaceRepo,
		imageRepo:           imageRepo,
		annotationRepo:      annotationRepo,
		imageEventPublisher: imageEventPublisher,
		uow:                 uow,
	}
}

func (ps *PatientService) CreateNewPatient(ctx context.Context, input port.CreatePatientInput) (*model.Patient, error) {

	patient, err := ps.patientRepo.FindByName(ctx, input.Name)
	if err != nil {
		return nil, errors.NewInternalError("failed to check existing patient name", err)
	}
	if patient != nil {
		return nil, errors.NewConflictError("patient with the same name already exists", map[string]interface{}{"name": "Patient name must be unique"})
	}

	ws, err := ps.workspaceRepo.Read(ctx, input.WorkspaceID)
	if err != nil {
		return nil, errors.NewValidationError("invalid workspace_id",
			map[string]interface{}{"workspace_id": "Workspace does not exist"})
	}

	if ws.AnnotationTypeID == nil || *ws.AnnotationTypeID == "" {
		return nil, errors.NewValidationError("workspace is not ready for patients",
			map[string]interface{}{"workspace_id": "Workspace must have an Annotation Type assigned before adding patients."})
	}
	createdPatient, err := ps.patientRepo.Create(ctx, &model.Patient{
		WorkspaceID: input.WorkspaceID,
		CreatorID:   input.CreatorID,
		Name:        input.Name,
		Age:         input.Age,
		Gender:      input.Gender,
		Race:        input.Race,
		Disease:     input.Disease,
		Subtype:     input.Subtype,
		Grade:       input.Grade,
		History:     input.History,
	})

	if err != nil {
		return nil, err
	}

	return createdPatient, nil
}

func (ps *PatientService) GetPatientByID(ctx context.Context, patientID string) (*model.Patient, error) {
	patient, err := ps.patientRepo.Read(ctx, patientID)
	if err != nil {
		return nil, err
	}
	return patient, nil
}

func (ps *PatientService) GetPatientsByWorkspaceID(ctx context.Context, workspaceID string, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[*model.Patient], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    constants.PatientWorkspaceIDField,
			Operator: sharedQuery.OpEqual,
			Value:    workspaceID,
		},
	}

	return ps.patientRepo.FindByFilters(ctx, filters, paginationOpts)
}

func (ps *PatientService) ListPatients(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[*model.Patient], error) {
	return ps.patientRepo.FindByFilters(ctx, []sharedQuery.Filter{}, paginationOpts)
}

func (ps *PatientService) DeletePatientByID(ctx context.Context, patientId string) error {
	uowerr := ps.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		filter := []sharedQuery.Filter{
			{
				Field:    constants.ImagePatientIDField,
				Operator: sharedQuery.OpEqual,
				Value:    patientId,
			},
		}
		pagination := &sharedQuery.Pagination{
			Limit:  1,
			Offset: 0,
		}

		existingImages, err := repos.ImageRepo.FindByFilters(
			txCtx,
			filter,
			pagination,
		)
		if err != nil {
			return err
		}

		if len(existingImages.Data) > 0 {
			return errors.NewConflictError("cannot delete patient with associated images", nil)
		}

		return repos.PatientRepo.Delete(txCtx, patientId)
	})

	if uowerr != nil {
		return uowerr
	}

	return nil
}

func (ps *PatientService) UpdatePatient(ctx context.Context, patientID string, input port.UpdatePatientInput) error {
	updates := make(map[string]interface{})

	if input.Name != nil {
		updates[constants.PatientNameField] = *input.Name
	}
	if input.Age != nil {
		updates[constants.PatientAgeField] = *input.Age
	}
	if input.Gender != nil {
		updates[constants.PatientGenderField] = *input.Gender
	}
	if input.Race != nil {
		updates[constants.PatientRaceField] = *input.Race
	}
	if input.Disease != nil {
		updates[constants.PatientDiseaseField] = *input.Disease
	}
	if input.Subtype != nil {
		updates[constants.PatientSubtypeField] = *input.Subtype
	}
	if input.Grade != nil {
		updates[constants.PatientGradeField] = *input.Grade
	}
	if input.History != nil {
		updates[constants.PatientHistoryField] = *input.History
	}

	if len(updates) == 0 {
		return nil
	}

	if err := ps.patientRepo.Update(ctx, patientID, updates); err != nil {
		return err
	}

	return nil
}

func (ps *PatientService) TransferPatientWorkspace(ctx context.Context, patientID string, newWorkspaceID string) error {

	uowerr := ps.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		_, err := repos.WorkspaceRepo.Read(txCtx, newWorkspaceID)
		if err != nil {
			return errors.NewConflictError("new workspace does not exist", nil)
		}

		return repos.PatientRepo.Transfer(txCtx, patientID, newWorkspaceID)
	})

	if uowerr != nil {
		return uowerr
	}

	return nil
}

func (ps *PatientService) CascadeDelete(ctx context.Context, patientID string) error {
	// Step 1: Delete annotations in batches
	if err := ps.deleteAnnotationsForPatient(ctx, patientID); err != nil {
		return errors.NewInternalError("failed to delete annotations", err)
	}

	// Step 2: Publish image deletion events asynchronously
	if err := ps.publishImageDeletionEvents(ctx, patientID); err != nil {
		return errors.NewInternalError("failed to publish image deletion events", err)
	}

	// Step 3: Delete patient record
	if err := ps.patientRepo.Delete(ctx, patientID); err != nil {
		return errors.NewInternalError("failed to delete patient", err)
	}

	return nil
}

func (ps *PatientService) deleteAnnotationsForPatient(ctx context.Context, patientID string) error {
	const batchSize = 500
	offset := 0

	for {
		// Get images for this patient
		imageFilters := []sharedQuery.Filter{
			{
				Field:    constants.ImagePatientIDField,
				Operator: sharedQuery.OpEqual,
				Value:    patientID,
			},
		}

		imageResult, err := ps.imageRepo.FindByFilters(ctx, imageFilters, &sharedQuery.Pagination{
			Limit:  100,
			Offset: offset,
		})
		if err != nil {
			return err
		}

		if len(imageResult.Data) == 0 {
			break
		}

		// Collect annotation IDs for these images
		annotationIDs := make([]string, 0)
		for _, image := range imageResult.Data {
			annoOffset := 0
			for {
				annoFilters := []sharedQuery.Filter{
					{
						Field:    constants.AnnotationImageIDField,
						Operator: sharedQuery.OpEqual,
						Value:    image.ID,
					},
				}

				annoResult, err := ps.annotationRepo.FindByFilters(ctx, annoFilters, &sharedQuery.Pagination{
					Limit:  100,
					Offset: annoOffset,
				})
				if err != nil {
					return err
				}

				for _, anno := range annoResult.Data {
					annotationIDs = append(annotationIDs, anno.ID)

					// Delete in batches to avoid memory issues
					if len(annotationIDs) >= batchSize {
						if err := ps.annotationRepo.BatchDelete(ctx, annotationIDs); err != nil {
							return err
						}
						annotationIDs = annotationIDs[:0] // Clear slice
					}
				}

				if !annoResult.HasMore {
					break
				}
				annoOffset += 100
			}
		}

		// Delete remaining annotations
		if len(annotationIDs) > 0 {
			if err := ps.annotationRepo.BatchDelete(ctx, annotationIDs); err != nil {
				return err
			}
		}

		if !imageResult.HasMore {
			break
		}
		offset += 100
	}

	return nil
}

func (ps *PatientService) publishImageDeletionEvents(ctx context.Context, patientID string) error {
	offset := 0
	const batchSize = 50 // Publish in smaller batches to avoid overwhelming the system

	for {
		imageFilters := []sharedQuery.Filter{
			{
				Field:    constants.ImagePatientIDField,
				Operator: sharedQuery.OpEqual,
				Value:    patientID,
			},
		}

		imageResult, err := ps.imageRepo.FindByFilters(ctx, imageFilters, &sharedQuery.Pagination{
			Limit:  batchSize,
			Offset: offset,
		})
		if err != nil {
			return err
		}

		if len(imageResult.Data) == 0 {
			break
		}

		// Publish deletion events for this batch
		for _, image := range imageResult.Data {
			event := events.NewImageDeletionRequestedEvent(image.ID)
			if err := ps.imageEventPublisher.PublishImageDeletionRequested(ctx, &event); err != nil {
				return fmt.Errorf("failed to publish deletion event for image %s: %w", image.ID, err)
			}
		}

		if !imageResult.HasMore {
			break
		}
		offset += batchSize
	}

	return nil
}

func (ps *PatientService) BatchDelete(ctx context.Context, patientIDs []string) error {
	// Process deletions in parallel but controlled manner
	errChan := make(chan error, len(patientIDs))
	semaphore := make(chan struct{}, 5) // Max 5 concurrent deletions

	for _, patientID := range patientIDs {
		semaphore <- struct{}{} // Acquire
		go func(pid string) {
			defer func() { <-semaphore }() // Release
			if err := ps.CascadeDelete(ctx, pid); err != nil {
				errChan <- fmt.Errorf("failed to delete patient %s: %w", pid, err)
			} else {
				errChan <- nil
			}
		}(patientID)
	}

	// Wait for all deletions and collect errors
	var firstErr error
	for range patientIDs {
		if err := <-errChan; err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

func (ps *PatientService) BatchTransfer(ctx context.Context, patientIDs []string, newWorkspaceID string) error {
	return ps.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		_, err := repos.WorkspaceRepo.Read(txCtx, newWorkspaceID)
		if err != nil {
			return errors.NewValidationError("new workspace does not exist",
				map[string]interface{}{"workspace_id": "Workspace not found"})
		}

		for _, patientID := range patientIDs {
			if err := repos.PatientRepo.Transfer(txCtx, patientID, newWorkspaceID); err != nil {
				return errors.NewInternalError("failed to transfer patient: "+patientID, err)
			}
		}

		return nil
	})
}

func (ps *PatientService) CountPatients(ctx context.Context, filters []sharedQuery.Filter) (int64, error) {
	return ps.patientRepo.Count(ctx, filters)
}
