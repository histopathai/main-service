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

type ListPatientsRequest struct {
	WorkspaceID   string  `form:"workspace_id" filter:"workspace_id"`
	CreatorID     string  `form:"creator_id" filter:"creator_id"`
	Age           *int    `form:"age" filter:"age"`
	Gender        *string `form:"gender" filter:"gender"`
	Disease       *string `form:"disease" filter:"disease"`
	AnonymousName *string `form:"anonymous_name" filter:"anonymous_name"`
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

func (p *PatientService) GetPatient(ctx context.Context, patientID string) (*models.Patient, error) {
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

func (p *PatientService) GetPatientsByWorkspaceID(ctx context.Context, workspaceID string, pagination repository.Pagination) (*repository.PatientQueryResult, error) {

	patients, err := p.repo.GetPatientsByWorkspaceID(ctx, workspaceID, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list patients by workspace ID", err)
	}
	if patients == nil {
		patients = &repository.PatientQueryResult{
			Patients: []*models.Patient{},
			Total:    0,
			Limit:    pagination.Limit,
			Offset:   pagination.Offset,
			HasMore:  false,
		}
	}
	return patients, nil
}

func (p *PatientService) GetPatientsByCreatorID(ctx context.Context, creatorID string, pagination repository.Pagination) (*repository.PatientQueryResult, error) {

	patients, err := p.repo.GetPatientByCreator(ctx, creatorID, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list patients by creator ID", err)
	}
	if patients == nil {
		patients = &repository.PatientQueryResult{
			Patients: []*models.Patient{},
			Total:    0,
			Limit:    pagination.Limit,
			Offset:   pagination.Offset,
			HasMore:  false,
		}
	}
	return patients, nil
}

func (p *PatientService) GetAllPatients(ctx context.Context, pagination repository.Pagination) (*repository.PatientQueryResult, error) {

	patients, err := p.repo.GetAllPatients(ctx, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list all patients", err)
	}

	if patients == nil {
		patients = &repository.PatientQueryResult{
			Patients: []*models.Patient{},
			Total:    0,
			Limit:    pagination.Limit,
			Offset:   pagination.Offset,
			HasMore:  false,
		}
	}
	return patients, nil
}

func (p *PatientService) SearchPatients(ctx context.Context, request *ListPatientsRequest, pagination repository.Pagination) (*repository.PatientQueryResult, error) {
	filters := make([]repository.Filter, 0)

	if request.WorkspaceID != "" {
		filters = append(filters, repository.Filter{
			Field: "workspace_id",
			Op:    repository.OpEqual,
			Value: request.WorkspaceID,
		})
	}
	if request.CreatorID != "" {
		filters = append(filters, repository.Filter{
			Field: "creator_id",
			Op:    repository.OpEqual,
			Value: request.CreatorID,
		})
	}
	if request.Age != nil {
		filters = append(filters, repository.Filter{
			Field: "age",
			Op:    repository.OpEqual,
			Value: *request.Age,
		})
	}
	if request.Gender != nil {
		filters = append(filters, repository.Filter{
			Field: "gender",
			Op:    repository.OpEqual,
			Value: *request.Gender,
		})
	}
	if request.Disease != nil {
		filters = append(filters, repository.Filter{
			Field: "disease",
			Op:    repository.OpEqual,
			Value: *request.Disease,
		})
	}
	if request.AnonymousName != nil {
		filters = append(filters, repository.Filter{
			Field: "anonymous_name",
			Op:    repository.OpEqual,
			Value: *request.AnonymousName,
		})
	}
	patients, err := p.repo.ListPatients(ctx, filters, pagination)
	if err != nil {
		return nil, apperrors.NewInternalError("failed to list patients", err)
	}
	if patients == nil {
		patients = &repository.PatientQueryResult{
			Patients: []*models.Patient{},
			Total:    0,
			Limit:    pagination.Limit,
			Offset:   pagination.Offset,
			HasMore:  false,
		}
	}
	return patients, nil
}
