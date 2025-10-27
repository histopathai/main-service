package firestore

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	errors "github.com/histopathai/main-service-refactor/internal/shared/errors"

	"cloud.google.com/go/firestore"
)

type AnnotationRepositoryImpl struct {
	client     *firestore.Client
	collection string
}

func NewAnnotationRepositoryImpl(client *firestore.Client) *AnnotationRepositoryImpl {
	return &AnnotationRepositoryImpl{
		client:     client,
		collection: constants.AnnotationsCollection,
	}
}

func (ar *AnnotationRepositoryImpl) toFirestoreMap(a *model.Annotation) map[string]interface{} {
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

func (ar *AnnotationRepositoryImpl) fromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Annotation, error) {
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

func (ar *AnnotationRepositoryImpl) Create(ctx context.Context, entity *model.Annotation) (string, error) {
	if entity == nil {
		return "", errors.NewInternalError("annotation entity is nil", nil)
	}

	if entity.ID == "" {
		newDocRef := ar.client.Collection(ar.collection).NewDoc()
		entity.ID = newDocRef.ID
	}

	_, err := ar.client.Collection(ar.collection).Doc(entity.ID).Set(ctx, ar.toFirestoreMap(entity))
	if err != nil {
		return "", errors.FromExternalError(err, "firestore")
	}
	return entity.ID, nil
}

func (ar *AnnotationRepositoryImpl) Read(ctx context.Context, id string) (*model.Annotation, error) {
	docSnap, err := ar.client.Collection(ar.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, errors.FromExternalError(err, "firestore")
	}

	annotation, err := ar.fromFirestoreDoc(docSnap)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse annotation data", err)
	}
	return annotation, nil
}

func (ar *AnnotationRepositoryImpl) Update(ctx context.Context, id string, updates map[string]interface{}) error {

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
		default:
			return errors.NewValidationError(fmt.Sprintf("unknown field %s for annotation update", key), nil)
		}

	}
	firestoreUpdates["updated_at"] = time.Now()
	_, err := ar.client.Collection(ar.collection).Doc(id).Set(ctx, firestoreUpdates, firestore.MergeAll)
	if err != nil {
		return errors.FromExternalError(err, "firestore")
	}
	return nil
}

func (ar *AnnotationRepositoryImpl) Delete(ctx context.Context, id string) error {
	_, err := ar.client.Collection(ar.collection).Doc(id).Delete(ctx)
	if err != nil {
		return errors.FromExternalError(err, "firestore")
	}
	return nil
}

func (ar *AnnotationRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// Annotations typically do not have an owner field to transfer.
	return nil
}
