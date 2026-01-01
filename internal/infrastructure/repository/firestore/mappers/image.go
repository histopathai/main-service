package firestoreMappers

import (
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ImageMapper struct{}

func (im *ImageMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Image, error) {
	ir := &model.Image{}

	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	beMapper := &BaseEntityMapper{}
	baseEntity, _ := beMapper.FromFirestoreDoc(doc)

	if baseEntity == nil {
		return nil, fmt.Errorf("failed to map base entity from firestore document")
	}

	ir.BaseEntity = *baseEntity
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

func (im *ImageMapper) ToFirestoreMap(i *model.Image) map[string]interface{} {
	beMapper := &BaseEntityMapper{}
	m := beMapper.ToFirestoreMap(&i.BaseEntity)

	m["format"] = i.Format
	m["origin_path"] = i.OriginPath
	m["status"] = i.Status
	m["retry_count"] = i.RetryCount

	if i.Width != nil {
		m["width"] = *i.Width
	}
	if i.Height != nil {
		m["height"] = *i.Height
	}
	if i.Size != nil {
		m["size"] = *i.Size
	}
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

func (im *ImageMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	beMapper := &BaseEntityMapper{}
	firestoreUpdates, _ := beMapper.MapUpdates(updates)

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
		case constants.ImageFailureReasonField:
			firestoreUpdates["failure_reason"] = value
		case constants.ImageRetryCountField:
			firestoreUpdates["retry_count"] = value
		case constants.ImageLastProcessedAtField:
			firestoreUpdates["last_processed_at"] = value
		default:
			return nil, fmt.Errorf("unknown field in image updates: %s", key)
		}
	}

	return firestoreUpdates, nil
}

func (im *ImageMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	beMapper := &BaseEntityMapper{}
	mappedFilters, _ := beMapper.MapFilters(filters)

	for _, filter := range filters {
		switch filter.Field {
		case constants.ImageStatusField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "status",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		case constants.ImageFormatField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "format",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
		default:
			return nil, fmt.Errorf("unknown filter field for image: %s", filter.Field)
		}
	}

	return mappedFilters, nil
}
