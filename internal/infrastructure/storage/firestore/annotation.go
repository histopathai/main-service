package firestore

import (
	"context"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"

	"cloud.google.com/go/firestore"
)

type AnnotationRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Annotation]

	_ repository.AnnotationTypeRepository // ensure interface compliance
}

func NewAnnotationRepositoryImpl(client *firestore.Client) *AnnotationRepositoryImpl {
	return &AnnotationRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl(
			client,
			constants.AnnotationsCollection,
			annotationFromFirestoreDoc,
			annotationToFirestoreMap,
			annotationMapUpdate,
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

func annotationMapUpdate(updates map[string]interface{}) map[string]interface{} {

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

		}

	}
	return firestoreUpdates
}

func (ar *AnnotationRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// Annotations typically do not have an owner field to transfer.
	return nil
}
