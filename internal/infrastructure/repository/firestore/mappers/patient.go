package firestoreMappers

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PatientMapper struct{}

func (pm *PatientMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Patient, error) {
	fp := &model.Patient{}

	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	beMapper := &BaseEntityMapper{}
	baseEntity, _ := beMapper.FromFirestoreDoc(doc)

	if baseEntity == nil {
		return nil, fmt.Errorf("failed to map base entity from firestore document")
	}

	fp.BaseEntity = *baseEntity

	if v, ok := data["age"]; ok && v != nil {
		age := int(v.(int64))
		fp.Age = &age
	}
	if v, ok := data["gender"]; ok && v != nil {
		gender := v.(string)
		fp.Gender = &gender
	}
	if v, ok := data["race"]; ok && v != nil {
		race := v.(string)
		fp.Race = &race
	}
	if v, ok := data["disease"]; ok && v != nil {
		disease := v.(string)
		fp.Disease = &disease
	}
	if v, ok := data["subtype"]; ok && v != nil {
		subtype := v.(string)
		fp.Subtype = &subtype
	}
	if v, ok := data["grade"]; ok && v != nil {
		grade := int(v.(int64))
		fp.Grade = &grade
	}
	if v, ok := data["history"]; ok && v != nil {
		history := v.(string)
		fp.History = &history
	}

	return fp, nil
}

func (pm *PatientMapper) ToFirestoreMap(p *model.Patient) map[string]interface{} {
	beMapper := &BaseEntityMapper{}
	m := beMapper.ToFirestoreMap(&p.BaseEntity)

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
		m["grade"] = *p.Grade
	}
	if p.History != nil {
		m["history"] = *p.History
	}

	return m
}

func (pm *PatientMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}
	firestoreUpdates := make(map[string]interface{})

	for k, v := range updates {
		switch k {
		case constants.PatientAgeField:
			firestoreUpdates["age"] = v
		case constants.PatientGenderField:
			firestoreUpdates["gender"] = v
		case constants.PatientRaceField:
			firestoreUpdates["race"] = v
		case constants.PatientDiseaseField:
			firestoreUpdates["disease"] = v
		case constants.PatientSubtypeField:
			firestoreUpdates["subtype"] = v
		case constants.PatientGradeField:
			firestoreUpdates["grade"] = v
		case constants.PatientHistoryField:
			firestoreUpdates["history"] = v

		default:
			return nil, fmt.Errorf("unknown field in patient updates: %s", k)
		}
	}

	return firestoreUpdates, nil
}

func (pm *PatientMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	beMapper := &BaseEntityMapper{}
	mappedFilters, _ := beMapper.MapFilters(filters)

	for _, filter := range filters {
		switch filter.Field {
		case constants.PatientAgeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "age",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.PatientGenderField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "gender",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.PatientRaceField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "race",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.PatientDiseaseField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "disease",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.PatientSubtypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "subtype",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.PatientGradeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "grade",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.PatientHistoryField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "history",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		}
	}

	return mappedFilters, nil
}
