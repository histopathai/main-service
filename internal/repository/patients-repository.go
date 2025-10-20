package repository

import (
	"context"
	"time"

	"github.com/histopathai/models"
)

const PatientsCollection = "patients"

type PatientRepository struct {
	repo Repository
}

func NewPatientRepository(repo Repository) *PatientRepository {
	return &PatientRepository{
		repo: repo,
	}
}

func (pr *PatientRepository) CreatePatient(ctx context.Context, patient *models.Patient) (string, error) {
	patient.CreatedAt = time.Now()
	patient.UpdatedAt = time.Now()
	return pr.repo.Create(ctx, PatientsCollection, patient.ToMap())
}

func (pr *PatientRepository) ReadPatient(ctx context.Context, patientID string) (*models.Patient, error) {
	data, err := pr.repo.Read(ctx, PatientsCollection, patientID)
	if err != nil {
		return nil, err
	}
	patient := &models.Patient{}
	patient.FromMap(data)
	return patient, nil
}

func (pr *PatientRepository) UpdatePatient(ctx context.Context, patientID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return pr.repo.Update(ctx, PatientsCollection, patientID, updates)
}

func (pr *PatientRepository) DeletePatient(ctx context.Context, patientID string) error {
	return pr.repo.Delete(ctx, PatientsCollection, patientID)
}

func (pr *PatientRepository) QueryPatients(ctx context.Context, filters map[string]interface{}) ([]*models.Patient, error) {
	results, err := pr.repo.Query(ctx, PatientsCollection, filters)
	if err != nil {
		return nil, err
	}
	patients := make([]*models.Patient, 0, len(results))
	for _, data := range results {
		patient := &models.Patient{}
		patient.FromMap(data)
		patients = append(patients, patient)
	}
	return patients, nil
}

func (pr *PatientRepository) Exists(ctx context.Context, patientID string) (bool, error) {
	return pr.repo.Exists(ctx, PatientsCollection, patientID)
}
