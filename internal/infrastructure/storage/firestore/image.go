package firestore

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"

	"cloud.google.com/go/firestore"
)

type ImageRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Image]
	_ repository.ImageRepository // ensure interface compliance
}

func NewImageRepositoryImpl(client *firestore.Client) *ImageRepositoryImpl {
	return &ImageRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl(
			client,
			constants.ImagesCollection,
			imageFromFirestoreDoc,
			imageToFirestoreMap,
			imageMapUpdates,
		),
	}
}

func imageMapUpdates(updates map[string]interface{}) map[string]interface{} {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {

		case constants.ImageFormatField:
			firestoreUpdates["format"] = value
		case constants.ImageWidthField:
			firestoreUpdates["width"] = value
		case constants.ImageHeightField:
			firestoreUpdates["height"] = value
		case constants.ImageSizeField:
			firestoreUpdates["size"] = value
		case constants.ImageProcessedPathField:
			firestoreUpdates["processed_path"] = value
		case constants.ImageStatusField:
			firestoreUpdates["status"] = value
		}
	}
	return firestoreUpdates
}
func imageToFirestoreMap(i *model.Image) map[string]interface{} {
	m := map[string]interface{}{
		"patient_id":  i.PatientID,
		"creator_id":  i.CreatorID,
		"file_name":   i.FileName,
		"format":      i.Format,
		"width":       i.Width,
		"height":      i.Height,
		"size":        i.Size,
		"origin_path": i.OriginPath,
	}

	if i.ProcessedPath != nil {
		m["processed_path"] = *i.ProcessedPath
	}
	m["status"] = i.Status
	m["created_at"] = i.CreatedAt
	m["updated_at"] = i.UpdatedAt
	return m
}

func imageFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Image, error) {
	ir := &model.Image{}
	data := doc.Data()

	ir.ID = doc.Ref.ID
	ir.PatientID = data["patient_id"].(string)
	ir.CreatorID = data["creator_id"].(string)
	ir.FileName = data["file_name"].(string)
	ir.Format = data["format"].(string)
	ir.OriginPath = data["origin_path"].(string)

	if v, ok := data["width"].(int64); ok {
		width := int(v)
		ir.Width = &width
	}
	if v, ok := data["height"].(int64); ok {
		height := int(v)
		ir.Height = &height
	}
	if v, ok := data["size"].(int64); ok {
		size := int64(v)
		ir.Size = &size
	}

	if v, ok := data["processed_path"].(string); ok {
		ir.ProcessedPath = &v
	}

	ir.Status = model.ImageStatus(data["status"].(string))
	ir.CreatedAt = data["created_at"].(time.Time)
	ir.UpdatedAt = data["updated_at"].(time.Time)

	return ir, nil
}

func (ir *ImageRepositoryImpl) Create(ctx context.Context, entity *model.Image) (*model.Image, error) {

	if entity == nil {
		return nil, errors.NewValidationError("image entity cannot be nil", nil)
	}

	if entity.ID == "" {
		entity.ID = ir.client.Collection(ir.collection).NewDoc().ID
	}
	entity.FileName = fmt.Sprintf("%s-%s", entity.ID, entity.FileName)

	return ir.GenericRepositoryImpl.Create(ctx, entity)
}

func (ir *ImageRepositoryImpl) Transfer(ctx context.Context, imageID string, newPatientID string) error {
	updates := map[string]interface{}{
		constants.ImagePatientIDField: newPatientID,
	}
	return ir.GenericRepositoryImpl.Update(ctx, imageID, updates)
}
