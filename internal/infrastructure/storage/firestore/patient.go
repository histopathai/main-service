package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PatientRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Patient]
	_ repository.PatientRepository // ensure interface compliance
}

func NewPatientRepositoryImpl(client *firestore.Client, hasUniqueName bool) *PatientRepositoryImpl {
	return &PatientRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl[*model.Patient](
			client,
			constants.PatientsCollection,
			hasUniqueName,
			patientFromFirestoreDoc,
			patientToFirestoreMap,
			patientMapUpdates,
			patientMapFilters,
		),
	}
}

func patientToFirestoreMap(p *model.Patient) map[string]interface{} {
	m := map[string]interface{}{

		"workspace_id": p.WorkspaceID,
		"name":         p.Name,
		"creator_id":   p.CreatorID,
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
	p.Name = data["name"].(string)
	p.CreatorID = data["creator_id"].(string)

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

func patientMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.PatientWorkspaceIDField:
			firestoreUpdates["workspace_id"] = value
		case constants.PatientCreatorIDField:
			firestoreUpdates["creator_id"] = value
		case constants.PatientNameField:
			firestoreUpdates["name"] = value
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
		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}
	}
	return firestoreUpdates, nil
}

func patientMapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))
	for _, f := range filters {
		switch f.Field {
		case constants.PatientWorkspaceIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "workspace_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientCreatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "creator_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientNameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "name",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientAgeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "age",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientGenderField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "gender",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientRaceField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "race",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientDiseaseField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "disease",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientSubtypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "subtype",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.PatientGradeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "grade",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.CreatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "created_at",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.UpdatedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "updated_at",
				Operator: f.Operator,
				Value:    f.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field: %s", f.Field)
		}
	}
	return mappedFilters, nil
}

func (pr *PatientRepositoryImpl) Transfer(ctx context.Context, patientID string, newWorkspaceID string) error {
	updates := map[string]interface{}{
		constants.PatientWorkspaceIDField: newWorkspaceID,
	}
	return pr.GenericRepositoryImpl.Update(ctx, patientID, updates)
}
