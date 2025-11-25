package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	errors "github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type PatientService struct {
	patientRepo   port.PatientRepository
	workspaceRepo port.WorkspaceRepository
	uow           port.UnitOfWorkFactory
}

func NewPatientService(
	patientRepo port.PatientRepository,
	workspaceRepo port.WorkspaceRepository,
	uow port.UnitOfWorkFactory,
) *PatientService {
	return &PatientService{
		patientRepo:   patientRepo,
		workspaceRepo: workspaceRepo,
		uow:           uow,
	}
}

func (ps *PatientService) CreateNewPatient(ctx context.Context, input CreatePatientInput) (*model.Patient, error) {

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

func (ps *PatientService) UpdatePatient(ctx context.Context, patientID string, input UpdatePatientInput) error {
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
	return ps.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		imageIDs := make([]string, 0)
		annotationIDs := make([]string, 0)

		offset := 0
		limit := 100

		for {
			imageFilters := []sharedQuery.Filter{
				{
					Field:    constants.ImagePatientIDField,
					Operator: sharedQuery.OpEqual,
					Value:    patientID,
				},
			}
			pagination := &sharedQuery.Pagination{Limit: limit, Offset: offset}

			imageResult, err := repos.ImageRepo.FindByFilters(txCtx, imageFilters, pagination)
			if err != nil {
				return errors.NewInternalError("failed to find images", err)
			}

			for _, image := range imageResult.Data {
				imageIDs = append(imageIDs, image.ID)

				annoOffset := 0
				for {
					annoFilters := []sharedQuery.Filter{
						{
							Field:    constants.AnnotationImageIDField,
							Operator: sharedQuery.OpEqual,
							Value:    image.ID,
						},
					}
					annoPagination := &sharedQuery.Pagination{Limit: limit, Offset: annoOffset}

					annoResult, err := repos.AnnotationRepo.FindByFilters(txCtx, annoFilters, annoPagination)
					if err != nil {
						return errors.NewInternalError("failed to find annotations", err)
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

			if !imageResult.HasMore {
				break
			}
			offset += limit
		}

		if len(annotationIDs) > 0 {
			if err := repos.AnnotationRepo.BatchDelete(txCtx, annotationIDs); err != nil {
				return errors.NewInternalError("failed to batch delete annotations", err)
			}
		}

		if len(imageIDs) > 0 {
			if err := repos.ImageRepo.BatchDelete(txCtx, imageIDs); err != nil {
				return errors.NewInternalError("failed to batch delete images", err)
			}
		}

		if err := repos.PatientRepo.Delete(txCtx, patientID); err != nil {
			return errors.NewInternalError("failed to delete patient", err)
		}

		return nil
	})
}

func (ps *PatientService) BatchDelete(ctx context.Context, patientIDs []string) error {
	for _, patientID := range patientIDs {
		if err := ps.CascadeDelete(ctx, patientID); err != nil {
			return errors.NewInternalError("failed to delete patient: "+patientID, err)
		}
	}
	return nil
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
