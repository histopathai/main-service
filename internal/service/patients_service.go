package service

import (
	"context"
	"log/slog"

	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/utils"
	"github.com/histopathai/models"
)

type CreatePatientRequest struct {
	AnonymousName *string `json:"anonymous_name,omitempty"`
	Age           *int    `json:"age,omitempty"`
	Gender        *string `json:"gender,omitempty"`
	Race          *string `json:"race,omitempty"`
	Disease       *string `json:"disease,omitempty"`
	History       *string `json:"history,omitempty"`
	WorkspaceID   string  `json:"workspace_id"`
}

type CreatePatientResponse struct {
	ID string `json:"id"`
}

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

func (p *PatientService) validateCreateWorkspaceRequest(req *CreatePatientRequest) error {
	details := make(map[string]interface{})
	if req.WorkspaceID == "" {
		details["workspace_id"] = "workspace_id is required"
	}

	if len(details) > 0 {
		return apperrors.NewBadRequestError("invalid create patient request", details)
	}
	return nil
}

func (p *PatientService) CreatePatient(ctx context.Context, req *CreatePatientRequest) (*CreatePatientResponse, error) {
	// Validate request
	if err := p.validateCreateWorkspaceRequest(req); err != nil {
		return nil, err
	}
	// Create patient model
	patient := &models.Patient{
		Age:         req.Age,
		Gender:      req.Gender,
		Race:        req.Race,
		Disease:     req.Disease,
		History:     req.History,
		WorkspaceID: req.WorkspaceID,
	}
	if req.AnonymousName != nil {
		patient.AnonymousName = *req.AnonymousName
	}
	if patient.AnonymousName == "" {
		fakeName := utils.GenerateFakeName()
		patient.AnonymousName = fakeName
	}

	// Save to repository
	patientID, err := p.repo.CreatePatient(ctx, patient)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to create patient", err)
	}

	// Log creation
	p.logger.Info("Created new patient", "patientID", patientID)
	return &CreatePatientResponse{
		ID: patientID,
	}, nil
}

func (p *PatientService) ListPatientsByWorkspaceID(ctx context.Context, workspaceID string) ([]*models.Patient, error) {
	filter := &models.Patient{
		WorkspaceID: workspaceID,
	}
	patients, err := p.repo.QueryPatients(ctx, filter.ToMap(), repository.Pagination{})
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list patients by workspace ID", err)
	}
	return patients, nil
}

func (p *PatientService) GetPatientByID(ctx context.Context, patientID string) (*models.Patient, error) {
	patient, err := p.repo.ReadPatient(ctx, patientID)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to get patient by ID", err)
	}
	return patient, nil
}
func (p *PatientService) UpdatePatient(ctx context.Context, patientID string, updates map[string]interface{}) error {
	err := p.repo.UpdatePatient(ctx, patientID, updates)
	if err != nil {
		return apperrors.NewInternalError("failed to update patient", err)
	}
	return nil
}
