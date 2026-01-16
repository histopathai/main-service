package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PatientMapper struct {
	*EntityMapper[*model.Patient]
}

func NewPatientMapper() *PatientMapper {
	return &PatientMapper{
		EntityMapper: NewEntityMapper[*model.Patient](),
	}
}

func (pm *PatientMapper) ToFirestoreMap(entity *model.Patient) map[string]interface{} {

	m := pm.EntityMapper.ToFirestoreMap(entity)

	// Patient specific fields
	if entity.Age != nil {
		m["age"] = *entity.Age
	}
	if entity.Gender != nil {
		m["gender"] = *entity.Gender
	}
	if entity.Race != nil {
		m["race"] = *entity.Race
	}
	if entity.Disease != nil {
		m["disease"] = *entity.Disease
	}
	if entity.Subtype != nil {
		m["subtype"] = *entity.Subtype
	}
	if entity.Grade != nil {
		m["grade"] = *entity.Grade
	}
	if entity.History != nil {
		m["history"] = *entity.History
	}

	return m
}

func (pm *PatientMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Patient, error) {

	entity, err := pm.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}

	patient := &model.Patient{
		Entity: *entity,
	}

	data := doc.Data()

	if age, ok := data["age"].(int); ok {
		patient.Age = &age
	}
	if gender, ok := data["gender"].(string); ok {
		patient.Gender = &gender
	}
	if race, ok := data["race"].(string); ok {
		patient.Race = &race
	}
	if disease, ok := data["disease"].(string); ok {
		patient.Disease = &disease
	}
	if subtype, ok := data["subtype"].(string); ok {
		patient.Subtype = &subtype
	}
	if grade, ok := data["grade"].(int); ok {
		patient.Grade = &grade
	}
	if history, ok := data["history"].(string); ok {
		patient.History = &history
	}

	return patient, nil
}

func (pm *PatientMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {

	mappedUpdates, err := pm.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	// Patient specific updates
	for k, v := range updates {
		switch k {
		case constants.PatientAgeField:
			if age, ok := v.(*int); ok {
				mappedUpdates["age"] = *age
			} else if ageInt, ok := v.(int); ok {
				mappedUpdates["age"] = ageInt
			} else {
				return nil, errors.NewValidationError("invalid type for age field", nil)
			}

		case constants.PatientGenderField:
			if gender, ok := v.(*string); ok {
				mappedUpdates["gender"] = *gender
			} else if genderStr, ok := v.(string); ok {
				mappedUpdates["gender"] = genderStr
			} else {
				return nil, errors.NewValidationError("invalid type for gender field", nil)
			}

		case constants.PatientRaceField:
			if race, ok := v.(*string); ok {
				mappedUpdates["race"] = *race
			} else if raceStr, ok := v.(string); ok {
				mappedUpdates["race"] = raceStr
			} else {
				return nil, errors.NewValidationError("invalid type for race field", nil)
			}

		case constants.PatientDiseaseField:
			if disease, ok := v.(*string); ok {
				mappedUpdates["disease"] = *disease
			} else if diseaseStr, ok := v.(string); ok {
				mappedUpdates["disease"] = diseaseStr
			} else {
				return nil, errors.NewValidationError("invalid type for disease field", nil)
			}

		case constants.PatientSubtypeField:
			if subtype, ok := v.(*string); ok {
				mappedUpdates["subtype"] = *subtype
			} else if subtypeStr, ok := v.(string); ok {
				mappedUpdates["subtype"] = subtypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for subtype field", nil)
			}

		case constants.PatientGradeField:
			if grade, ok := v.(*int); ok {
				mappedUpdates["grade"] = *grade
			} else if gradeInt, ok := v.(int); ok {
				mappedUpdates["grade"] = gradeInt
			} else {
				return nil, errors.NewValidationError("invalid type for grade field", nil)
			}

		case constants.PatientHistoryField:
			if history, ok := v.(*string); ok {
				mappedUpdates["history"] = *history
			} else if historyStr, ok := v.(string); ok {
				mappedUpdates["history"] = historyStr
			} else {
				return nil, errors.NewValidationError("invalid type for history field", nil)
			}

		}
	}

	return mappedUpdates, nil
}

func (pm *PatientMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters, err := pm.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	// Patient specific filters
	for _, f := range filters {
		switch f.Field {
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
		case constants.PatientHistoryField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "history",
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return mappedFilters, nil
}
