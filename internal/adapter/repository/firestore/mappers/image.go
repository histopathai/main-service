package mappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ImageMapper struct {
	*EntityMapper[*model.Image]
}

func NewImageMapper() *ImageMapper {
	return &ImageMapper{
		EntityMapper: NewEntityMapper[*model.Image](),
	}
}

func (im *ImageMapper) ToFirestoreMap(entity *model.Image) map[string]interface{} {

	m := im.EntityMapper.ToFirestoreMap(entity)

	// Image specific fields
	m["content_type"] = entity.ContentType
	m["format"] = entity.Format
	m["origin_path"] = entity.OriginPath
	m["status"] = entity.Status.String()
	m["retry_count"] = entity.RetryCount
	if entity.Width != nil {
		m["width"] = *entity.Width
	}
	if entity.Height != nil {
		m["height"] = *entity.Height
	}
	if entity.Size != nil {
		m["size"] = *entity.Size
	}
	if entity.ProcessedPath != nil {
		m["processed_path"] = *entity.ProcessedPath
	}
	if entity.FailureReason != nil {
		m["failure_reason"] = *entity.FailureReason
	}
	if entity.LastProcessedAt != nil {
		m["last_processed_at"] = *entity.LastProcessedAt
	}

	return m
}

func (im *ImageMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Image, error) {

	entity, err := im.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}

	image := &model.Image{
		Entity: *entity,
	}

	data := doc.Data()

	image.ContentType = data["content_type"].(string)
	image.Format = data["format"].(string)
	image.OriginPath = data["origin_path"].(string)
	image.RetryCount = int(data["retry_count"].(int64))
	image.Status, err = model.NewImageStatusFromString(data["status"].(string))
	if err != nil {
		return nil, err
	}

	if v, ok := data["width"].(int64); ok {
		width := int(v)
		image.Width = &width
	}
	if v, ok := data["height"].(int64); ok {
		height := int(v)
		image.Height = &height
	}
	if v, ok := data["size"].(int64); ok {
		size := int64(v)
		image.Size = &size
	}

	if v, ok := data["processed_path"].(string); ok {
		image.ProcessedPath = &v
	}

	if v, ok := data["failure_reason"].(string); ok {
		image.FailureReason = &v
	}

	if v, ok := data["last_processed_at"].(time.Time); ok {
		image.LastProcessedAt = &v
	}

	return image, nil
}

func (im *ImageMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates, err := im.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	// Image specific updates
	for k, v := range updates {
		switch k {
		case constants.ImageWidthField:
			if width, ok := v.(*int); ok {
				mappedUpdates["width"] = *width
			} else if widthInt, ok := v.(int); ok {
				mappedUpdates["width"] = widthInt
			} else {
				return nil, errors.NewValidationError("invalid type for width field", nil)
			}

		case constants.ImageHeightField:
			if height, ok := v.(*int); ok {
				mappedUpdates["height"] = *height
			} else if heightInt, ok := v.(int); ok {
				mappedUpdates["height"] = heightInt
			} else {
				return nil, errors.NewValidationError("invalid type for height field", nil)
			}

		case constants.ImageSizeField:
			if size, ok := v.(*int64); ok {
				mappedUpdates["size"] = *size
			} else if sizeInt64, ok := v.(int64); ok {
				mappedUpdates["size"] = sizeInt64
			} else {
				return nil, errors.NewValidationError("invalid type for size field", nil)
			}

		case constants.ImageProcessedPathField:
			if processedPath, ok := v.(*string); ok {
				mappedUpdates["processed_path"] = *processedPath
			} else if processedPathStr, ok := v.(string); ok {
				mappedUpdates["processed_path"] = processedPathStr
			} else {
				return nil, errors.NewValidationError("invalid type for processed_path field", nil)
			}

		case constants.ImageStatusField:
			if status, ok := v.(model.ImageStatus); ok {
				mappedUpdates["status"] = status.String()
			} else if statusStr, ok := v.(string); ok {
				mappedUpdates["status"] = statusStr
			} else {
				return nil, errors.NewValidationError("invalid type for status field", nil)
			}

		case constants.ImageFailureReasonField:
			if failureReason, ok := v.(*string); ok {
				mappedUpdates["failure_reason"] = *failureReason
			} else if failureReasonStr, ok := v.(string); ok {
				mappedUpdates["failure_reason"] = failureReasonStr
			} else {
				return nil, errors.NewValidationError("invalid type for failure_reason field", nil)
			}

		case constants.ImageRetryCountField:
			if retryCount, ok := v.(*int); ok {
				mappedUpdates["retry_count"] = *retryCount
			} else if retryCountInt, ok := v.(int); ok {
				mappedUpdates["retry_count"] = retryCountInt
			} else {
				return nil, errors.NewValidationError("invalid type for retry_count field", nil)
			}

		case constants.ImageLastProcessedAtField:
			if lastProcessedAt, ok := v.(*time.Time); ok {
				mappedUpdates["last_processed_at"] = *lastProcessedAt
			} else if lastProcessedAtTime, ok := v.(time.Time); ok {
				mappedUpdates["last_processed_at"] = lastProcessedAtTime
			} else {
				return nil, errors.NewValidationError("invalid type for last_processed_at field", nil)
			}
		}
	}

	return mappedUpdates, nil
}

func (im *ImageMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	firestoreFilters, err := im.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	// Image specific filters
	for _, f := range filters {
		switch f.Field {
		case constants.ImageStatusField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "status",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ImageFormatField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "format",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ImageWidthField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "width",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ImageHeightField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "height",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ImageSizeField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "size",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ImageFailureReasonField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "failure_reason",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ImageRetryCountField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "retry_count",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ImageLastProcessedAtField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "last_processed_at",
				Operator: f.Operator,
				Value:    f.Value,
			})
		}

	}

	return firestoreFilters, nil
}
