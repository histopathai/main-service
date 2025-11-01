package firestore

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/query"

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
			annotationTypeMapFilters,
		),
	}
}

func annotationTypeToFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationType, error) {
	atModel := &model.AnnotationType{}
	data := doc.Data()

	atModel.ID = doc.Ref.ID
	atModel.CreatorID = data["creator_id"].(string)
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
		"creator_id":             at.CreatorID,
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

func annotationTypeMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.AnnotationTypeNameField:
			firestoreUpdates["name"] = value
		case constants.AnnotationTypeCreatorIDField:
			firestoreUpdates["creator_id"] = value
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
		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}

	}
	return firestoreUpdates, nil
}

func annotationTypeMapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))
	for _, f := range filters {
		switch f.Field {
		case constants.AnnotationTypeNameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "name",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.AnnotationTypeCreatorIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "creator_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.AnnotationTypeScoreEnabledField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "score_enabled",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.AnnotationTypeClassificationEnabledField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "classification_enabled",
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

func (atr *AnnotationTypeRepositoryImpl) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// AnnotationType does not have an owner field; no action needed

	return nil
}
