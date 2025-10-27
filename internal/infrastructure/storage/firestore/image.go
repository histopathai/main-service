package firestore

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type ImageRepositoryImpl struct {
	client     *firestore.Client
	collection string
}

func NewImageRepositoryImpl(client *firestore.Client) *ImageRepositoryImpl {
	return &ImageRepositoryImpl{
		client:     client,
		collection: constants.ImagesCollection,
	}
}

func (ir *ImageRepositoryImpl) toFirestoreMap(i *model.Image) map[string]interface{} {
	m := map[string]interface{}{
		"patient_id":  i.PatientID,
		"creator_id":  i.CreatorID,
		"file_name":   i.FileName,
		"format":      i.Format,
		"width":       i.Width,
		"height":      i.Height,
		"size":        i.Size,
		"origin_path": i.OriginPath}

	if i.ProcessedPath != nil {
		m["processed_path"] = *i.ProcessedPath
	}
	m["status"] = i.Status
	m["created_at"] = i.CreatedAt
	m["updated_at"] = i.UpdatedAt
	return m
}

func (ir *ImageRepositoryImpl) fromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Image, error) {
	i := &model.Image{}
	data := doc.Data()

	i.ID = doc.Ref.ID
	i.PatientID = data["patient_id"].(string)
	i.CreatorID = data["creator_id"].(string)
	i.FileName = data["file_name"].(string)
	i.Format = data["format"].(string)

	i.OriginPath = data["origin_path"].(string)

	if v, ok := data["width"].(int64); ok {
		width := int(v)
		i.Width = &width
	}
	if v, ok := data["height"].(int64); ok {
		height := int(v)
		i.Height = &height
	}
	if v, ok := data["size"].(int64); ok {
		size := int64(v)
		i.Size = &size
	}

	if v, ok := data["processed_path"].(string); ok {
		i.ProcessedPath = &v
	}

	i.Status = model.ImageStatus(data["status"].(string))
	i.CreatedAt = data["created_at"].(time.Time)
	i.UpdatedAt = data["updated_at"].(time.Time)

	return i, nil
}

func (ir *ImageRepositoryImpl) Create(ctx context.Context, entity *model.Image) (*model.Image, error) {

	if entity == nil {
		return nil, errors.NewValidationError("image entity cannot be nil", nil)
	}

	if entity.ID == "" {
		entity.ID = ir.client.Collection(ir.collection).NewDoc().ID
	}
	filename := entity.FileName
	entity.FileName = fmt.Sprintf("%s-%s", entity.ID, filename)
	entity.CreatedAt = time.Now()
	entity.UpdatedAt = time.Now()

	data := ir.toFirestoreMap(entity)

	_, err := ir.client.Collection(ir.collection).Doc(entity.ID).Set(ctx, data)
	if err != nil {
		return nil, errors.FromExternalError(err, "firestore")
	}

	return entity, nil
}

func (ir *ImageRepositoryImpl) GetByID(ctx context.Context, id string) (*model.Image, error) {
	docSnap, err := ir.client.Collection(ir.collection).Doc(id).Get(ctx)

	if err != nil {
		return nil, errors.FromExternalError(err, "firestore")
	}
	image, err := ir.fromFirestoreDoc(docSnap)
	if err != nil {
		return nil, errors.NewInternalError("failed to parse image data", err)
	}
	return image, nil
}

func (ir *ImageRepositoryImpl) Update(ctx context.Context, id string, updates map[string]interface{}) error {

	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.ImageFileNameField:
			firestoreUpdates["file_name"] = value
		case constants.ImageFormatField:
			firestoreUpdates["format"] = value
		case constants.ImageWidthField:
			firestoreUpdates["width"] = value
		case constants.ImageHeightField:
			firestoreUpdates["height"] = value
		case constants.ImageSizeField:
			firestoreUpdates["size"] = value
		case constants.ImageOriginPathField:
			firestoreUpdates["origin_path"] = value
		case constants.ImageProcessedPathField:
			firestoreUpdates["processed_path"] = value
		case constants.ImageStatusField:
			firestoreUpdates["status"] = value
		}
	}
	firestoreUpdates["updated_at"] = time.Now()

	_, err := ir.client.Collection(ir.collection).Doc(id).Set(ctx, firestoreUpdates, firestore.MergeAll)
	if err != nil {
		return errors.FromExternalError(err, "firestore")
	}

	return nil
}

func (ir *ImageRepositoryImpl) WithTx(ctx context.Context, fn func(ctx context.Context, tx repository.Transaction) error) error {
	err := ir.client.RunTransaction(ctx, func(ctx context.Context, fstx *firestore.Transaction) error {

		tx := NewFirestoreTransaction(ir.client, fstx)

		return fn(ctx, tx)
	})

	if err != nil {
		return errors.FromExternalError(err, "firestore")
	}

	return nil
}

func (ir *ImageRepositoryImpl) GetByCriteria(ctx context.Context, filter []sharedQuery.Filter, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Image], error) {
	query := ir.client.Collection(ir.collection).Query

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

	images := []*model.Image{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, errors.FromExternalError(err, "firestore")
		}

		img, err := ir.fromFirestoreDoc(doc)
		if err != nil {
			continue
		}
		images = append(images, img)
	}

	hasmore := false
	if len(images) > paginationOpts.Limit {
		hasmore = true
		images = images[:len(images)-1]
	}

	return &sharedQuery.Result[model.Image]{
		Data:    images,
		Total:   0, // Total count can be implemented if needed
		Limit:   paginationOpts.Limit,
		Offset:  paginationOpts.Offset,
		HasMore: hasmore,
	}, nil
}

func (ir *ImageRepositoryImpl) GetByPatientID(ctx context.Context, patientID string, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.Image], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    "patient_id",
			Operator: sharedQuery.OpEqual,
			Value:    patientID,
		},
	}
	return ir.GetByCriteria(ctx, filters, paginationOpts)
}
