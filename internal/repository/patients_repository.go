package repository

import (
	"context"
	"time"

	"github.com/histopathai/models"
)

const PatientsCollection = "patients"

type PatientQueryResult struct {
	Patients []*models.Patient
	Total    int
	Limit    int
	Offset   int
	HasMore  bool
}

type PatientRepository struct {
	repo MainRepository
}

func NewPatientRepository(repo MainRepository) *PatientRepository {
	return &PatientRepository{
		repo: repo,
	}
}
func (pr *PatientRepository) GetMainRepository() *MainRepository {
	return &pr.repo
}

func (pr *PatientRepository) Create(ctx context.Context, patient *models.Patient) (string, error) {
	patient.CreatedAt = time.Now()
	patient.UpdatedAt = time.Now()
	return pr.repo.Create(ctx, PatientsCollection, patient.ToMap())
}

func (pr *PatientRepository) Read(ctx context.Context, patientID string) (*models.Patient, error) {
	data, err := pr.repo.Read(ctx, PatientsCollection, patientID)
	if err != nil {
		return nil, err
	}
	patient := &models.Patient{}
	patient.FromMap(data)
	return patient, nil
}

func (pr *PatientRepository) Update(ctx context.Context, patientID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return pr.repo.Update(ctx, PatientsCollection, patientID, updates)
}

func (pr *PatientRepository) Delete(ctx context.Context, patientID string) error {
	return pr.repo.Delete(ctx, PatientsCollection, patientID)
}

func (pr *PatientRepository) List(ctx context.Context, filters []Filter, pagination Pagination) (*PatientQueryResult, error) {
	result, err := pr.repo.List(ctx, PatientsCollection, filters, pagination)
	if err != nil {
		return nil, err
	}

	patients := make([]*models.Patient, 0, len(result.Data))
	for _, data := range result.Data {
		patient := &models.Patient{}
		patient.FromMap(data)
		patients = append(patients, patient)
	}

	return &PatientQueryResult{
		Patients: patients,
		Total:    result.Total,
		Limit:    result.Limit,
		Offset:   result.Offset,
		HasMore:  result.HasMore,
	}, nil
}

func (pr *PatientRepository) Exists(ctx context.Context, patientID string) (bool, error) {
	return pr.repo.Exists(ctx, PatientsCollection, patientID)
}
