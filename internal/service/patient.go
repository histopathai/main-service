package service

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	errors "github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
)

type PatientService struct {
	patientRepo repository.PatientRepository
	uow         repository.UnitOfWorkFactory
}

func NewPatientService(
	patientRepo repository.PatientRepository,
	uow repository.UnitOfWorkFactory,
	logger *slog.Logger,
) *PatientService {
	return &PatientService{
		patientRepo: patientRepo,
		uow:         uow,
	}
}

type CreatePatientInput struct {
	WorkspaceID string
	AnonymName  string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *int
	History     *string
}

func (ps *PatientService) CreateNewPatient(ctx context.Context, input CreatePatientInput) (*model.Patient, error) {

	createdPatient, err := ps.patientRepo.Create(ctx, &model.Patient{
		WorkspaceID: input.WorkspaceID,
		AnonymName:  input.AnonymName,
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

func (ps *PatientService) GetPatientsByWorkspaceID(ctx context.Context, workspaceID string, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Patient], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    constants.PatientWorkspaceIDField,
			Operator: sharedQuery.OpEqual,
			Value:    workspaceID,
		},
	}

	return ps.patientRepo.FindByFilters(ctx, filters, paginationOpts)
}

func (ps *PatientService) GetAllPatients(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Patient], error) {
	return ps.patientRepo.FindByFilters(ctx, []sharedQuery.Filter{}, paginationOpts)
}

func (ps *PatientService) DeletePatientByID(ctx context.Context, patientId string) error {
	uowerr := ps.uow.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {
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

type UpdatePatientInput struct {
	WorkspaceID *string
	AnonymName  *string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *int
	History     *string
}

func (ps *PatientService) UpdatePatient(ctx context.Context, patientID string, input UpdatePatientInput) error {
	updates := make(map[string]interface{})

	if input.WorkspaceID != nil {
		updates[constants.PatientWorkspaceIDField] = *input.WorkspaceID
	}

	if input.AnonymName != nil {
		updates[constants.PatientAnonymNameField] = *input.AnonymName
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

	uowerr := ps.uow.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {
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
