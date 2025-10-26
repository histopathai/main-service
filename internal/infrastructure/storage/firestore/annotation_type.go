package firestore

import (
	"context"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type AnnotationTypeRepositoryImpl struct {
	client     *firestore.Client
	collection string
}

func NewAnnotationTypeRepositoryImpl(client *firestore.Client) *AnnotationTypeRepositoryImpl {
	return &AnnotationTypeRepositoryImpl{
		client:     client,
		collection: constants.AnnotationTypesCollection,
	}
}

func (atr *AnnotationTypeRepositoryImpl) toFirestoreMap(at *model.AnnotationType) map[string]interface{} {

	m := map[string]interface{}{
		"name":                   at.Name,
		"score_enabled":          at.ScoreEnabled,
		"classification_enabled": at.ClassificationEnabled,
	}
	if at.ScoreRange != nil {
		m["score_range"] = *at.ScoreRange
	}
	if at.ScoreName != nil {
		m["score_name"] = *at.ScoreName
	}
	if at.ClassList != nil {
		m["class_list"] = *at.ClassList
	}
	m["created_at"] = at.CreatedAt
	m["updated_at"] = at.UpdatedAt
	return m
}

func (atr *AnnotationTypeRepositoryImpl) fromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationType, error) {
	at := &model.AnnotationType{}
	data := doc.Data()

	at.ID = doc.Ref.ID
	at.Name = data["name"].(string)
	at.ScoreEnabled = data["score_enabled"].(bool)
	at.ClassificationEnabled = data["classification_enabled"].(bool)

	if val, ok := data["score_range"]; ok {
		scoreRange := val.([]interface{})
		if len(scoreRange) == 2 {
			var arr [2]float64
			for i, v := range scoreRange {
				arr[i] = v.(float64)
			}
			at.ScoreRange = &arr
		}
	}

	if val, ok := data["score_name"]; ok {
		scoreName := val.(string)
		at.ScoreName = &scoreName
	}
	if val, ok := data["class_list"]; ok {
		classListInterface := val.([]interface{})
		classList := make([]string, len(classListInterface))
		for i, v := range classListInterface {
			classList[i] = v.(string)
		}
		at.ClassList = &classList
	}

	at.CreatedAt = data["created_at"].(time.Time)
	at.UpdatedAt = data["updated_at"].(time.Time)

	return at, nil
}

func (atr *AnnotationTypeRepositoryImpl) Create(ctx context.Context, entity *model.AnnotationType) (*model.AnnotationType, error) {

	if entity == nil {
		return nil, errors.NewValidationError("annotation type entity cannot be nil", nil)
	}

	if entity.ID == "" {
		entity.ID = atr.client.Collection(atr.collection).NewDoc().ID
	}

	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	data := atr.toFirestoreMap(entity)

	_, err := atr.client.Collection(atr.collection).Doc(entity.ID).Set(ctx, data)

	if err != nil {
		return nil, errors.NewInternalError("failed to create annotation type", err)
	}

	return entity, nil
}
func (atr *AnnotationTypeRepositoryImpl) GetByID(ctx context.Context, id string) (*model.AnnotationType, error) {
	docSnap, err := atr.client.Collection(atr.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, errors.NewNotFoundError("annotation type not found")
	}
	annotationType, err := atr.fromFirestoreDoc(docSnap)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse annotation type data", err)
	}
	return annotationType, nil
}

func (atr *AnnotationTypeRepositoryImpl) Update(ctx context.Context, id string, updates map[string]interface{}) error {

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
	firestoreUpdates["updated_at"] = time.Now()

	_, err := atr.client.Collection(atr.collection).Doc(id).Set(ctx, firestoreUpdates, firestore.MergeAll)
	if err != nil {
		return errors.NewInternalError("failed to update annotation type", err)
	}
	return nil
}

func (atr *AnnotationTypeRepositoryImpl) WithTx(ctx context.Context, fn func(ctx context.Context, tx repository.Transaction) error) error {
	err := atr.client.RunTransaction(ctx, func(ctx context.Context, fstx *firestore.Transaction) error {

		tx := NewFirestoreTransaction(atr.client, fstx)

		return fn(ctx, tx)
	})

	if err != nil {
		return errors.NewInternalError("firestore transaction failed", err)
	}

	return nil
}

func (atr *AnnotationTypeRepositoryImpl) GetByCriteria(ctx context.Context, filter []sharedQuery.Filter, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.AnnotationType], error) {
	query := atr.client.Collection(atr.collection).Query

	for _, f := range filter {
		query = query.Where(f.Field, string(f.Operator), f.Value)
	}

	// Apply pagination
	if paginationOpts == nil {
		paginationOpts = &sharedQuery.Pagination{
			Limit:  10,
			Offset: 0,
		}
	}

	query = query.Limit(paginationOpts.Limit + 1).Offset(paginationOpts.Offset)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var results []*model.AnnotationType

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.NewInternalError("failed to iterate annotation types", err)
		}
		at, err := atr.fromFirestoreDoc(doc)
		if err != nil {
			continue
		}
		results = append(results, at)
	}

	hasMore := false
	if len(results) > paginationOpts.Limit {
		hasMore = true
		results = results[:len(results)-1]
	}

	return &sharedQuery.Result[model.AnnotationType]{
		Data:    results,
		Total:   0,
		Limit:   paginationOpts.Limit,
		Offset:  paginationOpts.Offset,
		HasMore: hasMore,
	}, nil
}
