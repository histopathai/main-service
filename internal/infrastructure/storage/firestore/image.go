package firestore

import (
	"context"
	"fmt"
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"

	"cloud.google.com/go/firestore"
)

type ImageRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Image]
	_ repository.ImageRepository // ensure interface compliance
}

func NewImageRepositoryImpl(client *firestore.Client, hasUniqueName bool) *ImageRepositoryImpl {
	return &ImageRepositoryImpl{
		GenericRepositoryImpl: NewGenericRepositoryImpl[*model.Image](
			client,
			constants.ImagesCollection,
			hasUniqueName,
			imageFromFirestoreDoc,
			imageToFirestoreMap,
			imageMapUpdates,
			imageMapFilters,
		),
	}
}

func imageMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {

		case constants.ImageFormatField:
			firestoreUpdates["format"] = value
		case constants.ImageCreatorIDField:
			firestoreUpdates["creator_id"] = value
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
		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}
	}
	return firestoreUpdates, nil
}
func imageToFirestoreMap(i *model.Image) map[string]interface{} {
	m := map[string]interface{}{
		"patient_id":  i.PatientID,
		"creator_id":  i.CreatorID,
		"name":        i.Name,
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
	ir.Name = data["name"].(string)
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

func imageMapFilters(filters []query.Filter) ([]query.Filter, error) {
	var firestoreFilters []query.Filter
	for _, filter := range filters {
		switch filter.Field {
		case constants.ImagePatientIDField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "patient_id",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.ImageCreatorIDField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "creator_id",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.ImageNameField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "name",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.ImageStatusField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "status",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.ImageFormatField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "format",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.CreatedAtField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "created_at",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.UpdatedAtField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "updated_at",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field: %s", filter.Field)
		}
	}
	return firestoreFilters, nil
}

func (ir *ImageRepositoryImpl) Create(ctx context.Context, entity *model.Image) (*model.Image, error) {

	if entity == nil {
		return nil, ErrInvalidInput
	}
	imageCopy := *entity
	if imageCopy.ID == "" {
		imageCopy.ID = ir.client.Collection(ir.collection).NewDoc().ID
	}
	imageCopy.Name = fmt.Sprintf("%s-%s", imageCopy.ID, imageCopy.Name)

	return ir.GenericRepositoryImpl.Create(ctx, &imageCopy)
}

func (ir *ImageRepositoryImpl) Transfer(ctx context.Context, imageID string, newPatientID string) error {
	updates := map[string]interface{}{
		constants.ImagePatientIDField: newPatientID,
	}
	return ir.GenericRepositoryImpl.Update(ctx, imageID, updates)
}
