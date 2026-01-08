package firestoreMappers

import (
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ImageMapper struct {
	entityMapper *EntityMapper
}

func NewImageMapper() *ImageMapper {
	return &ImageMapper{
		entityMapper: &EntityMapper{},
	}
}

func (im *ImageMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Image, error) {
	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	entity, err := im.entityMapper.FromFirestoreDoc(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity from firestore document: %w", err)
	}

	image := &model.Image{
		Entity:     entity,
		Format:     data["format"].(string),
		OriginPath: data["origin_path"].(string),
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

	if v, ok := data["processed_path"].(string); ok && v != "" {
		image.ProcessedPath = &v
	}

	processReport, err := im.parseProcessReport(data)
	if err != nil {
		return nil, err
	}
	image.ProcessReport = processReport

	return image, nil
}

func (im *ImageMapper) parseProcessReport(data map[string]interface{}) (*vobj.ImageProcessReport, error) {
	statusStr, ok := data["status"].(string)
	if !ok {
		return nil, errors.NewValidationError("status is required", nil)
	}

	status, err := vobj.NewImageStatusFromString(statusStr)
	if err != nil {
		return nil, err
	}

	report := &vobj.ImageProcessReport{
		Status:     status,
		RetryCount: 0,
	}

	if v, ok := data["failure_reason"].(string); ok && v != "" {
		report.FailureReason = &v
	}

	if v, ok := data["retry_count"].(int64); ok {
		report.RetryCount = int(v)
	}

	if v, ok := data["last_processed_at"].(time.Time); ok {
		report.LastProcessedAt = &v
	}

	return report, nil
}

func (im *ImageMapper) ToFirestoreMap(image *model.Image) map[string]interface{} {
	m := im.entityMapper.ToFirestoreMap(image.Entity)

	m["format"] = image.Format
	m["origin_path"] = image.OriginPath

	if image.Width != nil {
		m["width"] = *image.Width
	}

	if image.Height != nil {
		m["height"] = *image.Height
	}

	if image.Size != nil {
		m["size"] = *image.Size
	}

	if image.ProcessedPath != nil {
		m["processed_path"] = *image.ProcessedPath
	}

	m["status"] = image.ProcessReport.Status.String()
	m["retry_count"] = image.ProcessReport.RetryCount

	if image.ProcessReport.FailureReason != nil {
		m["failure_reason"] = *image.ProcessReport.FailureReason
	}

	if image.ProcessReport.LastProcessedAt != nil {
		m["last_processed_at"] = *image.ProcessReport.LastProcessedAt
	}

	return m
}

func (im *ImageMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	firestoreUpdates, err := im.entityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for key, value := range updates {
		switch key {
		case constants.FormatField:
			firestoreUpdates["format"] = value
			delete(updates, constants.FormatField)

		case constants.WidthField:
			firestoreUpdates["width"] = value
			delete(updates, constants.WidthField)

		case constants.HeightField:
			firestoreUpdates["height"] = value
			delete(updates, constants.HeightField)

		case constants.SizeField:
			firestoreUpdates["size"] = value
			delete(updates, constants.SizeField)

		case constants.OriginPathField:
			firestoreUpdates["origin_path"] = value
			delete(updates, constants.OriginPathField)

		case constants.ProcessedPathField:
			firestoreUpdates["processed_path"] = value
			delete(updates, constants.ProcessedPathField)

		case constants.StatusField:
			firestoreUpdates["status"] = value
			delete(updates, constants.StatusField)

		case constants.FailureReasonField:
			firestoreUpdates["failure_reason"] = value
			delete(updates, constants.FailureReasonField)

		case constants.RetryCountField:
			firestoreUpdates["retry_count"] = value
			delete(updates, constants.RetryCountField)

		case constants.LastProcessedAtField:
			firestoreUpdates["last_processed_at"] = value
			delete(updates, constants.LastProcessedAtField)
		}
	}

	return firestoreUpdates, nil
}

func (im *ImageMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	mappedFilters, err := im.entityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	unprocessedIdx := 0
	for i, filter := range filters {
		processed := false

		switch filter.Field {
		case constants.FormatField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "format",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.WidthField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "width",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.HeightField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "height",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.SizeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "size",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.StatusField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "status",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.RetryCountField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "retry_count",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.LastProcessedAtField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "last_processed_at",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true
		}

		if !processed {
			filters[unprocessedIdx] = filters[i]
			unprocessedIdx++
		}
	}

	for i := unprocessedIdx; i < len(filters); i++ {
		filters[i] = query.Filter{}
	}

	return mappedFilters, nil
}
