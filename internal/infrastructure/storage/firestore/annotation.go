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

type AnnotationRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Annotation]

	_ port.AnnotationTypeRepository // ensure interface compliance
}

func NewAnnotationRepositoryImpl(client *firestore.Client, hasUniqueName bool) *AnnotationRepositoryImpl {
	return &AnnotationRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl[*model.Annotation](
			client,
			constants.AnnotationsCollection,
			hasUniqueName,
			annotationFromFirestoreDoc,
			annotationToFirestoreMap,
			annotationMapUpdate,
			annotationMapFilters,
		),
	}
}

func annotationToFirestoreMap(a *model.Annotation) map[string]interface{} {
	m := EntityToFirestoreMap(&a.Entity)

	m["polygon"] = a.Polygon

	m["tag_value"] = a.TagValue.Value
	m["tag_type"] = string(a.TagValue.Type)
	m["is_global"] = a.TagValue.Global
	if a.TagValue.Color != nil {
		m["color"] = *a.TagValue.Color
	} else {
		m["color"] = nil
	}
	return m
}
func annotationFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Annotation, error) {
	var a model.Annotation
	data := doc.Data()

	entity, err := EntityFromFirestore(doc)
	if err != nil {
		return nil, err
	}

	a.Entity = *entity

	a.TagValue.Value = data["tag_value"]
	tag_type, err := vobj.NewTagTypeFromString(data["tag_type"].(string))
	if err != nil {
		return nil, err
	}
	a.TagValue.Type = tag_type

	if v, ok := data["color"].(string); ok {
		a.TagValue.Color = &v
	}

	if v, ok := data["is_global"].(bool); ok {
		a.TagValue.Global = v
	}

	if polygonData, ok := data["polygon"].([]interface{}); ok {
		jsonPoints := make([]map[string]float64, len(polygonData))
		for i, p := range polygonData {
			pointMap, ok := p.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid point format at index %d", i)
			}

			jsonPoints[i] = make(map[string]float64)
			if x, ok := pointMap["X"].(float64); ok {
				jsonPoints[i]["X"] = x
			} else if xInt, ok := pointMap["X"].(int64); ok {
				jsonPoints[i]["X"] = float64(xInt)
			} else {
				return nil, fmt.Errorf("invalid X value at index %d", i)
			}

			if y, ok := pointMap["Y"].(float64); ok {
				jsonPoints[i]["Y"] = y
			} else if yInt, ok := pointMap["Y"].(int64); ok {
				jsonPoints[i]["Y"] = float64(yInt)
			} else {
				return nil, fmt.Errorf("invalid Y value at index %d", i)
			}
		}

		polygon := vobj.FromJSONPoints(jsonPoints)
		a.Polygon = &polygon
	}

	return &a, nil
}

func annotationMapUpdate(updates map[string]interface{}) (map[string]interface{}, error) {

	entityUpdates, err := EntityMapUpdates(updates)
	if err != nil {
		return nil, err
	}

	firestoreUpdates := make(map[string]interface{})
	for key, value := range entityUpdates {
		firestoreUpdates[key] = value
	}

	for key, value := range updates {
		if EntityFields[key] {
			continue
		}
		switch key {
		case constants.PolygonField:
			firestoreUpdates["polygon"] = value
		case constants.TagValueField:
			firestoreUpdates["tag_value"] = value
		case constants.TagTypeField:
			firestoreUpdates["tag_type"] = value.(string)
		case constants.TagColorField:
			if value != nil {
				firestoreUpdates["color"] = value.(string)
			} else {
				firestoreUpdates["color"] = nil
			}
		case constants.TagGlobalField:
			firestoreUpdates["is_global"] = value.(bool)

		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}
	}
	return firestoreUpdates, nil
}

func annotationMapFilters(filters []query.Filter) ([]query.Filter, error) {
	entityFilter, err := EntityMapFilter(filters)
	if err != nil {
		return nil, err
	}

	mappedFilters := make([]query.Filter, len(entityFilter))
	copy(mappedFilters, entityFilter)

	for _, f := range filters {
		if EntityFields[f.Field] {
			continue
		}
		switch f.Field {
		case constants.PolygonField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "polygon",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagValueField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_value",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_type",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagColorField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "color",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagGlobalField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_global",
				Operator: f.Operator,
				Value:    f.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field: %s", f.Field)
		}
	}
	return mappedFilters, nil
}

func (ar *AnnotationRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// Annotations typically do not have an owner field to transfer.
	return nil
}
