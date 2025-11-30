package firestore

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
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
	m := map[string]interface{}{
		"image_id":     a.ImageID,
		"annotator_id": a.AnnotatorID,
		"polygon":      a.Polygon,
		"created_at":   a.CreatedAt,
		"updated_at":   a.UpdatedAt,
	}

	if a.Score != nil {
		m["score"] = *a.Score
	}
	if a.Class != nil {
		m["class"] = *a.Class
	}
	return m
}

func annotationFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Annotation, error) {
	var a model.Annotation
	data := doc.Data()

	a.ID = doc.Ref.ID
	a.ImageID = data["image_id"].(string)
	a.AnnotatorID = data["annotator_id"].(string)

	points := data["polygon"].([]interface{})
	a.Polygon = make([]model.Point, len(points))
	for i, p := range points {
		pointMap := p.(map[string]interface{})
		a.Polygon[i] = model.Point{
			X: pointMap["X"].(float64),
			Y: pointMap["Y"].(float64),
		}
	}
	if description, ok := data["description"]; ok {
		desc := description.(string)
		a.Description = &desc
	}
	if score, ok := data["score"]; ok {
		s := score.(float64)
		a.Score = &s
	}
	if class, ok := data["class"]; ok {
		c := class.(string)
		a.Class = &c
	}
	a.CreatedAt = data["created_at"].(time.Time)
	a.UpdatedAt = data["updated_at"].(time.Time)

	return &a, nil
}

func annotationMapUpdate(updates map[string]interface{}) (map[string]interface{}, error) {

	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.AnnotationAnnotatorIDField:
			firestoreUpdates["annotator_id"] = value
		case constants.AnnotationPolygonField:
			firestoreUpdates["polygon"] = value
		case constants.AnnotationScoreField:
			firestoreUpdates["score"] = value
		case constants.AnnotationClassField:
			firestoreUpdates["class"] = value
		case constants.AnnotationDescriptionField:
			firestoreUpdates["description"] = value
		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}
	}
	return firestoreUpdates, nil
}

func annotationMapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))
	for _, f := range filters {
		switch f.Field {
		case constants.AnnotationAnnotatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "annotator_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.AnnotationImageIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "image_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.AnnotationClassField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "class",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.AnnotationScoreField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "score",
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

func (ar *AnnotationRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// Annotations typically do not have an owner field to transfer.
	return nil
}
