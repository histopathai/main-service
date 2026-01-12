package firestore

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"

	"cloud.google.com/go/firestore"
)

type AnnotationTypeRepositoryImpl struct {
	*GenericRepositoryImpl[*model.AnnotationType]

	_ port.AnnotationTypeRepository // ensure interface compliance
}

func NewAnnotationTypeRepositoryImpl(client *firestore.Client, hasUniqueName bool) *AnnotationTypeRepositoryImpl {
	return &AnnotationTypeRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl[*model.AnnotationType](
			client,
			constants.AnnotationTypesCollection,
			hasUniqueName,
			annotationTypeToFirestoreDoc,
			annotatationTypeFirestoreToMap,
			annotationTypeMapUpdates,
			annotationTypeMapFilters,
		),
	}
}

func annotationTypeToFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationType, error) {
	atModel := &model.AnnotationType{}

	entity, err := EntityFromFirestore(doc)
	if err != nil {
		return nil, err
	}
	atModel.Entity = *entity

	data := doc.Data()

	if v, ok := data["type"].(string); ok {
		atModel.Type, err = vobj.NewTagTypeFromString(v)
		if err != nil {
			return nil, err
		}
	}
	if v, ok := data["is_global"].(bool); ok {
		atModel.Global = v
	}
	if v, ok := data["is_required"].(bool); ok {
		atModel.Required = v
	}
	if v, ok := data["options"].([]interface{}); ok {
		options := make([]string, len(v))
		for i, option := range v {
			options[i] = option.(string)
		}
		atModel.Options = options
	}
	if v, ok := data["min"].(float64); ok {
		atModel.Min = &v
	}
	if v, ok := data["max"].(float64); ok {
		atModel.Max = &v
	}
	if v, ok := data["color"].(string); ok {
		atModel.Color = &v
	}

	return atModel, nil
}

func annotatationTypeFirestoreToMap(at *model.AnnotationType) map[string]interface{} {
	m_entity := EntityToFirestoreMap(&at.Entity)

	m := make(map[string]interface{})
	for k, v := range m_entity {
		m[k] = v
	}
	m["type"] = string(at.Type)
	m["is_global"] = at.Global
	m["is_required"] = at.Required
	m["options"] = at.Options
	if at.Min != nil {
		m["min"] = *at.Min
	}
	if at.Max != nil {
		m["max"] = *at.Max
	}
	if at.Color != nil {
		m["color"] = *at.Color
	}

	return m
}

func annotationTypeMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {

	entityUpdates, err := EntityMapUpdates(updates)
	if err != nil {
		return nil, err
	}

	firestoreUpdates := make(map[string]interface{})
	for k, v := range entityUpdates {
		firestoreUpdates[k] = v
	}
	for key, value := range updates {
		if EntityFields[key] {
			continue
		}
		switch key {
		case constants.TagTypeField:
			firestoreUpdates["type"] = value.(string)
		case constants.TagGlobalField:
			firestoreUpdates["is_global"] = value.(bool)
		case constants.TagRequiredField:
			firestoreUpdates["is_required"] = value.(bool)
		case constants.TagOptionsField:
			firestoreUpdates["options"] = value.([]string)
		case constants.TagMaxField:
			firestoreUpdates["max"] = value.(*float64)
		case constants.TagMinField:
			firestoreUpdates["min"] = value.(*float64)
		case constants.TagColorField:
			firestoreUpdates["color"] = value.(*string)
		}
	}

	return firestoreUpdates, nil
}

func annotationTypeMapFilters(filters []query.Filter) ([]query.Filter, error) {
	entityMappedFilters, err := EntityMapFilter(filters)
	if err != nil {
		return nil, err
	}
	mappedFilters := make([]query.Filter, len(entityMappedFilters))
	copy(mappedFilters, entityMappedFilters)

	for _, f := range filters {
		if EntityFields[f.Field] {
			continue
		}
		switch f.Field {
		case constants.TagTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "type",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagGlobalField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_global",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagRequiredField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_required",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagOptionsField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "options",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagMinField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "min",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagMaxField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "max",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagColorField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "color",
				Operator: f.Operator,
				Value:    f.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field: %s", f.Field)
		}

	}

	return mappedFilters, nil
}

func (atr *AnnotationTypeRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// AnnotationType does not have an owner field; no action needed

	return nil
}
