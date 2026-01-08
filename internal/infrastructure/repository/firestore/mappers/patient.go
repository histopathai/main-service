package firestoreMappers

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PatientMapper struct {
	entityMapper *EntityMapper
}

func NewPatientMapper() *PatientMapper {
	return &PatientMapper{
		entityMapper: &EntityMapper{},
	}
}

func (pm *PatientMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Patient, error) {
	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	entity, err := pm.entityMapper.FromFirestoreDoc(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity from firestore document: %w", err)
	}

	patient := &model.Patient{
		Entity: entity,
	}

	if v, ok := data["age"]; ok && v != nil {
		age := int(v.(int64))
		patient.Age = &age
	}

	if v, ok := data["gender"]; ok && v != nil {
		gender := v.(string)
		patient.Gender = &gender
	}

	if v, ok := data["race"]; ok && v != nil {
		race := v.(string)
		patient.Race = &race
	}

	if v, ok := data["disease"]; ok && v != nil {
		disease := v.(string)
		patient.Disease = &disease
	}

	if v, ok := data["subtype"]; ok && v != nil {
		subtype := v.(string)
		patient.Subtype = &subtype
	}

	if v, ok := data["grade"]; ok && v != nil {
		grade := int(v.(int64))
		patient.Grade = &grade
	}

	if v, ok := data["history"]; ok && v != nil {
		history := v.(string)
		patient.History = &history
	}

	return patient, nil
}

func (pm *PatientMapper) ToFirestoreMap(patient *model.Patient) map[string]interface{} {

	m := pm.entityMapper.ToFirestoreMap(patient.Entity)

	if patient.Age != nil {
		m["age"] = *patient.Age
	}

	if patient.Gender != nil {
		m["gender"] = *patient.Gender
	}

	if patient.Race != nil {
		m["race"] = *patient.Race
	}

	if patient.Disease != nil {
		m["disease"] = *patient.Disease
	}

	if patient.Subtype != nil {
		m["subtype"] = *patient.Subtype
	}

	if patient.Grade != nil {
		m["grade"] = *patient.Grade
	}

	if patient.History != nil {
		m["history"] = *patient.History
	}

	return m
}

func (pm *PatientMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	firestoreUpdates, err := pm.entityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		switch k {
		case constants.AgeField:
			firestoreUpdates["age"] = v
			delete(updates, constants.AgeField)

		case constants.GenderField:
			firestoreUpdates["gender"] = v
			delete(updates, constants.GenderField)

		case constants.RaceField:
			firestoreUpdates["race"] = v
			delete(updates, constants.RaceField)

		case constants.DiseaseField:
			firestoreUpdates["disease"] = v
			delete(updates, constants.DiseaseField)

		case constants.SubtypeField:
			firestoreUpdates["subtype"] = v
			delete(updates, constants.SubtypeField)

		case constants.GradeField:
			firestoreUpdates["grade"] = v
			delete(updates, constants.GradeField)

		case constants.HistoryField:
			firestoreUpdates["history"] = v
			delete(updates, constants.HistoryField)
		}
	}

	return firestoreUpdates, nil
}

func (pm *PatientMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	mappedFilters, err := pm.entityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	unprocessedIdx := 0
	for i, filter := range filters {
		processed := false

		switch filter.Field {
		case constants.AgeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "age",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.GenderField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "gender",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.RaceField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "race",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.DiseaseField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "disease",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.SubtypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "subtype",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.GradeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "grade",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.HistoryField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "history",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		}

		if !processed {
			filters[unprocessedIdx] = filters[i]
			unprocessedIdx++
		}
	}

	for i := unprocessedIdx; i < len(filters); i++ {
		filters[i] = query.Filter{}
	}

	return mappedFilters, nil
}
