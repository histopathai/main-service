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

type ImageRepositoryImpl struct {
	*GenericRepositoryImpl[*model.Image]
	_ port.ImageRepository // ensure interface compliance
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
	firestoreUpdates, err := EntityMapUpdates(updates)
	if err != nil {
		return nil, err
	}
	for key, value := range updates {
		if EntityFields[key] {
			continue
		}
		switch key {
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
		case constants.ImageFailureReasonField:
			firestoreUpdates["failure_reason"] = value
		case constants.ImageRetryCountField:
			firestoreUpdates["retry_count"] = value
		case constants.ImageLastProcessedAtField:
			firestoreUpdates["last_processed_at"] = value
		default:
			return nil, fmt.Errorf("unknown update field: %s", key)
		}
	}
	return firestoreUpdates, nil
}

func imageToFirestoreMap(i *model.Image) map[string]interface{} {
	m := EntityToFirestoreMap(&i.Entity)

	m["format"] = i.Format
	m["width"] = i.Width
	m["height"] = i.Height
	m["size"] = i.Size
	m["origin_path"] = i.OriginPath
	m["status"] = i.Status
	m["retry_count"] = i.RetryCount

	if i.ProcessedPath != nil {
		m["processed_path"] = *i.ProcessedPath
	}
	if i.FailureReason != nil {
		m["failure_reason"] = *i.FailureReason
	}
	if i.LastProcessedAt != nil {
		m["last_processed_at"] = *i.LastProcessedAt
	}

	return m
}

func imageFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Image, error) {
	ir := &model.Image{}
	data := doc.Data()

	entity, err := EntityFromFirestore(doc)
	if err != nil {
		return nil, err
	}

	ir.Entity = *entity

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

	if v, ok := data["failure_reason"].(string); ok {
		ir.FailureReason = &v
	}
	if v, ok := data["retry_count"].(int64); ok {
		ir.RetryCount = int(v)
	}
	if v, ok := data["last_processed_at"].(time.Time); ok {
		ir.LastProcessedAt = &v
	}

	ir.Status = model.ImageStatus(data["status"].(string))

	return ir, nil
}

func imageMapFilters(filters []query.Filter) ([]query.Filter, error) {
	firestoreFilters, err := EntityMapFilter(filters)
	if err != nil {
		return nil, err
	}
	for _, filter := range filters {
		if EntityFields[filter.Field] {
			continue
		}
		switch filter.Field {
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
	// Note: Naming convention might differ based on requirements, assuming ID is sufficient uniqueness
	// imageCopy.Name = fmt.Sprintf("%s-%s", imageCopy.ID, imageCopy.Name)

	return ir.GenericRepositoryImpl.Create(ctx, &imageCopy)
}

func (ir *ImageRepositoryImpl) Transfer(ctx context.Context, imageID string, newPatientID string) error {
	updates := map[string]interface{}{
		constants.ParentIDField: newPatientID,
	}
	return ir.GenericRepositoryImpl.Update(ctx, imageID, updates)
}
