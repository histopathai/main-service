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
		firestoreField := fields.MapToFirestore(k)

		switch k {
		case fields.PatientAge.DomainName():
			if age, ok := v.(*int); ok {
				mappedUpdates[firestoreField] = *age
			} else if ageInt, ok := v.(int); ok {
				mappedUpdates[firestoreField] = ageInt
			} else {
				return nil, errors.NewValidationError("invalid type for age field", nil)
			}

		case fields.PatientGender.DomainName():
			if gender, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *gender
			} else if genderStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = genderStr
			} else {
				return nil, errors.NewValidationError("invalid type for gender field", nil)
			}

		case fields.PatientRace.DomainName():
			if race, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *race
			} else if raceStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = raceStr
			} else {
				return nil, errors.NewValidationError("invalid type for race field", nil)
			}

		case fields.PatientDisease.DomainName():
			if disease, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *disease
			} else if diseaseStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = diseaseStr
			} else {
				return nil, errors.NewValidationError("invalid type for disease field", nil)
			}

		case fields.PatientSubtype.DomainName():
			if subtype, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *subtype
			} else if subtypeStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = subtypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for subtype field", nil)
			}

		case fields.PatientGrade.DomainName():
			if grade, ok := v.(*int); ok {
				mappedUpdates[firestoreField] = *grade
			} else if gradeInt, ok := v.(int); ok {
				mappedUpdates[firestoreField] = gradeInt
			} else {
				return nil, errors.NewValidationError("invalid type for grade field", nil)
			}

		case fields.PatientHistory.DomainName():
			if history, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *history
			} else if historyStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = historyStr
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
