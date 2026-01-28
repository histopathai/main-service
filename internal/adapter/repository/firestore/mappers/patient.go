package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
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
		m[fields.PatientAge.FirestoreName()] = *entity.Age
	}
	if entity.Gender != nil {
		m[fields.PatientGender.FirestoreName()] = *entity.Gender
	}
	if entity.Race != nil {
		m[fields.PatientRace.FirestoreName()] = *entity.Race
	}
	if entity.Disease != nil {
		m[fields.PatientDisease.FirestoreName()] = *entity.Disease
	}
	if entity.Subtype != nil {
		m[fields.PatientSubtype.FirestoreName()] = *entity.Subtype
	}
	if entity.Grade != nil {
		m[fields.PatientGrade.FirestoreName()] = *entity.Grade
	}
	if entity.History != nil {
		m[fields.PatientHistory.FirestoreName()] = *entity.History
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

	if age, ok := data[fields.PatientAge.FirestoreName()].(int); ok {
		patient.Age = &age
	}
	if gender, ok := data[fields.PatientGender.FirestoreName()].(string); ok {
		patient.Gender = &gender
	}
	if race, ok := data[fields.PatientRace.FirestoreName()].(string); ok {
		patient.Race = &race
	}
	if disease, ok := data[fields.PatientDisease.FirestoreName()].(string); ok {
		patient.Disease = &disease
	}
	if subtype, ok := data[fields.PatientSubtype.FirestoreName()].(string); ok {
		patient.Subtype = &subtype
	}
	if grade, ok := data[fields.PatientGrade.FirestoreName()].(int); ok {
		patient.Grade = &grade
	}
	if history, ok := data[fields.PatientHistory.FirestoreName()].(string); ok {
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
		case fields.PatientAge.DomainName():
			if age, ok := v.(*int); ok {
				mappedUpdates[fields.PatientAge.FirestoreName()] = *age
			} else if ageInt, ok := v.(int); ok {
				mappedUpdates[fields.PatientAge.FirestoreName()] = ageInt
			} else {
				return nil, errors.NewValidationError("invalid type for age field", nil)
			}

		case fields.PatientGender.DomainName():
			if gender, ok := v.(*string); ok {
				mappedUpdates[fields.PatientGender.FirestoreName()] = *gender
			} else if genderStr, ok := v.(string); ok {
				mappedUpdates[fields.PatientGender.FirestoreName()] = genderStr
			} else {
				return nil, errors.NewValidationError("invalid type for gender field", nil)
			}

		case fields.PatientRace.DomainName():
			if race, ok := v.(*string); ok {
				mappedUpdates[fields.PatientRace.FirestoreName()] = *race
			} else if raceStr, ok := v.(string); ok {
				mappedUpdates[fields.PatientRace.FirestoreName()] = raceStr
			} else {
				return nil, errors.NewValidationError("invalid type for race field", nil)
			}

		case fields.PatientDisease.DomainName():
			if disease, ok := v.(*string); ok {
				mappedUpdates[fields.PatientDisease.FirestoreName()] = *disease
			} else if diseaseStr, ok := v.(string); ok {
				mappedUpdates[fields.PatientDisease.FirestoreName()] = diseaseStr
			} else {
				return nil, errors.NewValidationError("invalid type for disease field", nil)
			}

		case fields.PatientSubtype.DomainName():
			if subtype, ok := v.(*string); ok {
				mappedUpdates[fields.PatientSubtype.FirestoreName()] = *subtype
			} else if subtypeStr, ok := v.(string); ok {
				mappedUpdates[fields.PatientSubtype.FirestoreName()] = subtypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for subtype field", nil)
			}

		case fields.PatientGrade.DomainName():
			if grade, ok := v.(*int); ok {
				mappedUpdates[fields.PatientGrade.FirestoreName()] = *grade
			} else if gradeInt, ok := v.(int); ok {
				mappedUpdates[fields.PatientGrade.FirestoreName()] = gradeInt
			} else {
				return nil, errors.NewValidationError("invalid type for grade field", nil)
			}

		case fields.PatientHistory.DomainName():
			if history, ok := v.(*string); ok {
				mappedUpdates[fields.PatientHistory.FirestoreName()] = *history
			} else if historyStr, ok := v.(string); ok {
				mappedUpdates[fields.PatientHistory.FirestoreName()] = historyStr
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
		firestoreField := fields.MapToFirestore(f.Field)

		if fields.EntityField(f.Field).IsValid() {
			continue
		}

		if fields.PatientField(f.Field).IsValid() {
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    firestoreField,
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return mappedFilters, nil
}
