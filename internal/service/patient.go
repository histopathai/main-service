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
	patientRepo   repository.PatientRepository
	workspaceRepo repository.WorkspaceRepository
	logger        *slog.Logger
}

func NewPatientService(
	patientRepo repository.PatientRepository,
	workspaceRepo repository.WorkspaceRepository,
	logger *slog.Logger,
) *PatientService {
	return &PatientService{
		patientRepo:   patientRepo,
		workspaceRepo: workspaceRepo,
		logger:        logger,
	}
}

func (ps *PatientService) validatePatientCreation(ctx context.Context, workspaceID string) error {

	workspace, err := ps.workspaceRepo.GetByID(ctx, workspaceID)
	if err != nil {
		ps.logger.Error("Failed to read workspace during patient validation", "error", err, "workspaceID", workspaceID)
		details := map[string]interface{}{"workspace_id": "Failed to read workspace."}
		return errors.NewValidationError("failed to read workspace", details)
	}
	if workspace == nil {
		ps.logger.Error("Workspace not found during patient validation", "workspaceID", workspaceID)
		details := map[string]interface{}{"workspace_id": "Workspace not found."}
		return errors.NewValidationError("workspace not found", details)
	}

	if workspace.AnnotationTypeID == nil || *workspace.AnnotationTypeID == "" {
		details := map[string]interface{}{"workspace_id": "Workspace must have an annotation type assigned."}
		return errors.NewValidationError("workspace must have an annotation type", details)
	}

	return nil
}

type CreatePatientInput struct {
	WorkspaceID string
	AnonymName  string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *string
	History     *string
}

func (ps *PatientService) CreatePatient(ctx context.Context, input CreatePatientInput) (*model.Patient, error) {

	if err := ps.validatePatientCreation(ctx, input.WorkspaceID); err != nil {
		return nil, err
	}

	newPatient := &model.Patient{
		WorkspaceID: input.WorkspaceID,
		AnonymName:  input.AnonymName,
		Age:         input.Age,
		Gender:      input.Gender,
		Race:        input.Race,
		Disease:     input.Disease,
		Subtype:     input.Subtype,
		Grade:       input.Grade,
		History:     input.History,
	}

	created, err := ps.patientRepo.Create(ctx, newPatient)
	if err != nil {
		ps.logger.Error("Failed to create patient", "error", err, "workspaceID", input.WorkspaceID)
		return nil, errors.NewInternalError("failed to create patient", err)
	}
	ps.logger.Info("Patient created successfully", "patientID", created.ID, "workspaceID", input.WorkspaceID)

	return created, nil
}

func (ps *PatientService) GetPatientByID(ctx context.Context, patientID string) (*model.Patient, error) {
	patient, err := ps.patientRepo.GetByID(ctx, patientID)
	if err != nil {
		ps.logger.Error("Failed to retrieve patient", "error", err, "patientID", patientID)
		return nil, errors.NewInternalError("failed to retrieve patient", err)
	}

	ps.logger.Info("Patient retrieved successfully", "patientID", patientID)
	return patient, nil
}

type UpdatePatientInput struct {
	WorkspaceID *string
	AnonymName  *string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *string
	History     *string
}

func (ps *PatientService) UpdatePatient(ctx context.Context, patientID string, input UpdatePatientInput) error {
	updates := make(map[string]interface{})

	if input.WorkspaceID != nil {
		if err := ps.validatePatientCreation(ctx, *input.WorkspaceID); err != nil {
			return err
		}
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
		ps.logger.Info("No updates provided for patient", "patientID", patientID)
		return nil
	}

	if err := ps.patientRepo.Update(ctx, patientID, updates); err != nil {
		ps.logger.Error("Failed to update patient", "error", err, "patientID", patientID)
		return errors.NewInternalError("failed to update patient", err)
	}

	ps.logger.Info("Patient updated successfully", "patientID", patientID)
	return nil
}

func (ps *PatientService) GetPatientsByWorkspaceID(ctx context.Context, workspaceID string, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Patient], error) {
	patients, err := ps.patientRepo.GetByWorkSpaceID(ctx, workspaceID, paginationOpts)
	if err != nil {
		ps.logger.Error("Failed to retrieve patients by workspace ID", "error", err, "workspaceID", workspaceID)
		return nil, errors.NewInternalError("failed to retrieve patients", err)
	}

	ps.logger.Info("Patients retrieved successfully", "workspaceID", workspaceID, "count", patients.Total)
	return patients, nil
}

func (ps *PatientService) GetAllPatients(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Patient], error) {
	patients, err := ps.patientRepo.GetByCriteria(ctx, []sharedQuery.Filter{}, paginationOpts)
	if err != nil {
		ps.logger.Error("Failed to retrieve all patients", "error", err)
		return nil, errors.NewInternalError("failed to retrieve patients", err)
	}

	ps.logger.Info("All patients retrieved successfully", "count", patients.Total)
	return patients, nil
}
