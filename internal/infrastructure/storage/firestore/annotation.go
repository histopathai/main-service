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
	m_tag_value := TagValueToFirestoreMap(&a.TagValue)
	for k, v := range m_tag_value {
		m[k] = v
	}
	m["polygon"] = a.Polygon
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

	tagValue, err := TagValueFromFirestoreDoc(doc)
	if err != nil {
		return nil, err
	}
	a.TagValue = *tagValue

	points := data["polygon"].([]interface{})
	polygon := make([]vobj.Point, len(points))
	for i, p := range points {
		pointMap := p.(map[string]interface{})
		polygon[i] = vobj.Point{
			X: pointMap["X"].(float64),
			Y: pointMap["Y"].(float64),
		}
	}
	a.Polygon = &polygon

	return &a, nil
}

func annotationMapUpdate(updates map[string]interface{}) (map[string]interface{}, error) {

	entityUpdates, err := EntityMapUpdates(updates)
	if err != nil {
		return nil, err
	}
	tagvalueUpdates, err := TagValueMapUpdates(updates)
	if err != nil {
		return nil, err
	}

	firestoreUpdates := make(map[string]interface{})
	for key, value := range entityUpdates {
		firestoreUpdates[key] = value
	}
	for key, value := range tagvalueUpdates {
		firestoreUpdates[key] = value
	}

	for key, value := range updates {
		switch key {
		case constants.PolygonField:
			firestoreUpdates["polygon"] = value
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
	tagValueFilters, err := TagValueMapFilters(filters)
	if err != nil {
		return nil, err
	}

	mappedFilters := make([]query.Filter, 0, len(entityFilter)+len(tagValueFilters))
	mappedFilters = append(mappedFilters, entityFilter...)
	mappedFilters = append(mappedFilters, tagValueFilters...)

	return mappedFilters, nil
}

func (ar *AnnotationRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// Annotations typically do not have an owner field to transfer.
	return nil
}
