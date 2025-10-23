package service

import (
	"context"
	"log/slog"

	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/utils"
	"github.com/histopathai/models"
)

type PatientService struct {
	repo   *repository.PatientRepository
	logger *slog.Logger
}

func NewPatientService(repo *repository.PatientRepository, logger *slog.Logger) *PatientService {
	return &PatientService{
		repo:   repo,
		logger: logger,
	}
}

func (p *PatientService) CreatePatient(ctx context.Context,
	patient *models.Patient) (string, error) {

	ok, err := p.repo.GetMainRepository().Exists(ctx,
		repository.WorkspacesCollection, patient.WorkspaceID)
	if err != nil {
		return "", apperrors.NewInternalError("failed to verify workspace existence", err)
	}
	if !ok {
		return "", apperrors.NewValidationError("workspace does not exist", nil)
	}

	if patient.AnonymousName == "" {
		fakeName := utils.GenerateFakeName()
		patient.AnonymousName = fakeName
	}
	patientID, err := p.repo.Create(ctx, patient)
	if err != nil {
		return "", apperrors.NewInternalError("failed to create patient", err)
	}
	p.logger.Info("Created new patient", "patientID", patientID)
	return patientID, nil
}

func (p *PatientService) GetPatient(ctx context.Context,
	patientID string) (*models.Patient, error) {

	patient, err := p.repo.Read(ctx, patientID)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to get patient", err)
	}
	if patient == nil {
		return nil, apperrors.NewNotFoundError("patient not found")
	}
	return patient, nil
}

func (p *PatientService) UpdatePatient(ctx context.Context,
	patientID string, updates map[string]interface{}) error {
	_, ok := updates["workspace_id"]
	if ok {
		exists, err := p.repo.GetMainRepository().Exists(ctx,
			repository.WorkspacesCollection, updates["workspace_id"].(string))
		if err != nil {
			return apperrors.NewInternalError("failed to verify workspace existence", err)
		}
		if !exists {
			return apperrors.NewValidationError("workspace does not exist", nil)
		}
	}

	_, ok = updates["creator_id"]
	if ok {

		exists, err := p.repo.GetMainRepository().Exists(ctx,
			"users", updates["creator_id"].(string))
		if err != nil {
			return apperrors.NewInternalError("failed to verify creator existence", err)
		}
		if !exists {
			return apperrors.NewValidationError("creator does not exist for patient", nil)
		}
	}

	err := p.repo.Update(ctx, patientID, updates)
	if err != nil {
		return apperrors.NewInternalError("failed to update patient", err)
	}
	p.logger.Info("Updated patient", "patientID", patientID)
	return nil
}

func (p *PatientService) GetPatientByCreatorID(ctx context.Context,
	creatorID string, pagination repository.Pagination) (*repository.PatientQueryResult, error) {
	filters := []repository.Filter{
		{
			Field: "creator_id",
			Op:    repository.OpEqual,
			Value: creatorID,
		},
	}
	result, err := p.repo.List(ctx, filters, pagination)
	if err != nil {
		p.logger.Error("failed to list patients by creator ID", "error", err)
		return nil, apperrors.NewInternalError("failed to list patients by creator ID", err)
	}
	return result, nil
}

func (p *PatientService) GetPatientsByWorkspaceID(ctx context.Context,
	workspaceID string, pagination repository.Pagination) (*repository.PatientQueryResult, error) {
	filters := []repository.Filter{
		{
			Field: "workspace_id",
			Op:    repository.OpEqual,
			Value: workspaceID,
		},
	}
	result, err := p.repo.List(ctx, filters, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list patients by workspace ID", err)
	}
	return result, nil
}

func (p *PatientService) GetAllPatients(ctx context.Context,
	pagination repository.Pagination) (*repository.PatientQueryResult, error) {
	result, err := p.repo.List(ctx, []repository.Filter{}, pagination)
	if err != nil {
		p.logger.Error("failed to list all patients", "error", err)
		return nil, apperrors.NewInternalError("failed to list all patients", err)
	}
	return result, nil
}
