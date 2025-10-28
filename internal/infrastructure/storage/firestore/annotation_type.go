package firestore

import (
	"context"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"

	"cloud.google.com/go/firestore"
)

type AnnotationTypeRepositoryImpl struct {
	*GenericRepositoryImpl[*model.AnnotationType]

	_ repository.AnnotationTypeRepository // ensure interface compliance
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
		),
	}
}

func annotationTypeToFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationType, error) {
	atModel := &model.AnnotationType{}
	data := doc.Data()

	atModel.ID = doc.Ref.ID
	atModel.Name = data["name"].(string)
	atModel.ScoreEnabled = data["score_enabled"].(bool)
	atModel.ClassificationEnabled = data["classification_enabled"].(bool)

	if val, ok := data["score_range"]; ok {
		scoreRange := val.([]interface{})
		if len(scoreRange) == 2 {
			var arr [2]float64
			for i, v := range scoreRange {
				arr[i] = v.(float64)
			}
			atModel.ScoreRange = &arr
		}
	}

	if val, ok := data["score_name"]; ok {
		scoreName := val.(string)
		atModel.ScoreName = &scoreName
	}
	if val, ok := data["class_list"]; ok {
		classListInterface := val.([]interface{})
		classList := make([]string, len(classListInterface))
		for i, v := range classListInterface {
			classList[i] = v.(string)
		}
		atModel.ClassList = classList
	}

	atModel.CreatedAt = data["created_at"].(time.Time)
	atModel.UpdatedAt = data["updated_at"].(time.Time)

	return atModel, nil
}

func annotatationTypeFirestoreToMap(at *model.AnnotationType) map[string]interface{} {
	m := map[string]interface{}{
		"name":                   at.Name,
		"score_enabled":          at.ScoreEnabled,
		"classification_enabled": at.ClassificationEnabled,
		"created_at":             at.CreatedAt,
		"updated_at":             at.UpdatedAt,
	}
	if at.ScoreRange != nil {
		m["score_range"] = *at.ScoreRange
	}
	if at.ScoreName != nil {
		m["score_name"] = *at.ScoreName
	}
	if len(at.ClassList) > 0 {
		m["class_list"] = at.ClassList
	}
	return m
}

func annotationTypeMapUpdates(updates map[string]interface{}) map[string]interface{} {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.AnnotationTypeNameField:
			firestoreUpdates["name"] = value
		case constants.AnnotationTypeScoreEnabledField:
			firestoreUpdates["score_enabled"] = value
		case constants.AnnotationTypeClassificationEnabledField:
			firestoreUpdates["classification_enabled"] = value
		case constants.AnnotationTypeScoreRangeField:
			firestoreUpdates["score_range"] = value
		case constants.AnnotationTypeScoreNameField:
			firestoreUpdates["score_name"] = value
		case constants.AnnotationTypeClassListField:
			firestoreUpdates["class_list"] = value
		}
	}
	return firestoreUpdates
}

func (atr *AnnotationTypeRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// AnnotationType does not have an owner field; no action needed

	return nil
}
