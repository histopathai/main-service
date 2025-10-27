package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
)

type PatientRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Patient]
	_ repository.PatientRepository // ensure interface compliance
}

func NewPatientRepositoryImpl(client *firestore.Client) *PatientRepositoryImpl {
	return &PatientRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl(
			client,
			constants.PatientsCollection,
			patientFromFirestoreDoc,
			patientToFirestoreMap,
			patientMapUpdates,
		),
	}
}

func patientToFirestoreMap(p *model.Patient) map[string]interface{} {
	m := map[string]interface{}{

		"workspace_id": p.WorkspaceID,
		"anonym_name":  p.AnonymName,
	}
	if p.Age != nil {
		m["age"] = *p.Age
	}
	if p.Gender != nil {
		m["gender"] = *p.Gender
	}
	if p.Race != nil {
		m["race"] = *p.Race
	}
	if p.Disease != nil {
		m["disease"] = *p.Disease
	}
	if p.Subtype != nil {
		m["subtype"] = *p.Subtype
	}
	if p.Grade != nil {
		grade := int(*p.Grade)
		m["grade"] = grade
	}
	if p.History != nil {
		m["history"] = *p.History
	}
	m["created_at"] = p.CreatedAt
	m["updated_at"] = p.UpdatedAt
	return m
}

func patientFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Patient, error) {
	p := &model.Patient{}
	data := doc.Data()

	p.ID = doc.Ref.ID
	p.WorkspaceID = data["workspace_id"].(string)
	p.AnonymName = data["anonym_name"].(string)

	if v, ok := data["age"].(int64); ok {
		age := int(v)
		p.Age = &age
	}

	if v, ok := data["gender"].(string); ok {
		p.Gender = &v
	}

	if v, ok := data["race"].(string); ok {
		p.Race = &v
	}

	if v, ok := data["disease"].(string); ok {
		p.Disease = &v
	}

	if v, ok := data["subtype"].(string); ok {
		p.Subtype = &v
	}

	if v, ok := data["grade"].(int64); ok {
		grade := int(v)
		p.Grade = &grade
	}

	if v, ok := data["history"].(string); ok {
		p.History = &v
	}

	p.CreatedAt, _ = data["created_at"].(time.Time)
	p.UpdatedAt, _ = data["updated_at"].(time.Time)

	return p, nil
}

func patientMapUpdates(updates map[string]interface{}) map[string]interface{} {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.PatientWorkspaceIDField:
			firestoreUpdates["workspace_id"] = value
		case constants.PatientAnonymNameField:
			firestoreUpdates["anonym_name"] = value
		case constants.PatientAgeField:
			firestoreUpdates["age"] = value
		case constants.PatientGenderField:
			firestoreUpdates["gender"] = value
		case constants.PatientRaceField:
			firestoreUpdates["race"] = value
		case constants.PatientDiseaseField:
			firestoreUpdates["disease"] = value
		case constants.PatientSubtypeField:
			firestoreUpdates["subtype"] = value
		case constants.PatientGradeField:
			firestoreUpdates["grade"] = value
		case constants.PatientHistoryField:
			firestoreUpdates["history"] = value
		}
	}
	return firestoreUpdates
}

func (pr *PatientRepositoryImpl) Transfer(ctx context.Context, patientID string, newWorkspaceID string) error {
	updates := map[string]interface{}{
		constants.PatientWorkspaceIDField: newWorkspaceID,
	}
	return pr.GenericRepositoryImpl.Update(ctx, patientID, updates)
}
