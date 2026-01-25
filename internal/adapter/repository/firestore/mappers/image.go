package mappers

import (
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
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

	// Basic fields
	m[fields.ImageWsID.FirestoreName()] = entity.WsID
	m[fields.ImageFormat.FirestoreName()] = entity.Format

	// Optional basic fields
	if entity.Width != nil {
		m[fields.ImageWidth.FirestoreName()] = *entity.Width
	}
	if entity.Height != nil {
		m[fields.ImageHeight.FirestoreName()] = *entity.Height
	}

	// Magnification (nested object or null)
	if entity.Magnification != nil {
		magMap := make(map[string]interface{})
		if entity.Magnification.Objective != nil {
			magMap["objective"] = *entity.Magnification.Objective
		}
		if entity.Magnification.NativeLevel != nil {
			magMap["native_level"] = *entity.Magnification.NativeLevel
		}
		if entity.Magnification.ScanMagnification != nil {
			magMap["scan_magnification"] = *entity.Magnification.ScanMagnification
		}
		if len(magMap) > 0 {
			m["magnification"] = magMap
		}
	}

	// Origin Content ID
	if entity.OriginContentID != nil {
		m[fields.ImageOriginContentID.FirestoreName()] = *entity.OriginContentID
	}

	// Processed Content IDs
	if entity.ThumbnailContentID != nil {
		m[fields.ImageThumbnailContentID.FirestoreName()] = *entity.ThumbnailContentID
	}
	if entity.DziContentID != nil {
		m[fields.ImageDziContentID.FirestoreName()] = *entity.DziContentID
	}
	if entity.IndexmapContentID != nil {
		m[fields.ImageIndexmapContentID.FirestoreName()] = *entity.IndexmapContentID
	}
	if entity.TilesContentID != nil {
		m[fields.ImageTilesContentID.FirestoreName()] = *entity.TilesContentID
	}
	if entity.ZipTilesContentID != nil {
		m[fields.ImageZipTilesContentID.FirestoreName()] = *entity.ZipTilesContentID
	}

	// Processing info
	processingMap := make(map[string]interface{})
	processingMap["status"] = entity.Processing.Status.String()
	processingMap["version"] = entity.Processing.Version.String()
	processingMap["retry_count"] = entity.Processing.RetryCount
	if entity.Processing.FailureReason != nil {
		processingMap["failure_reason"] = *entity.Processing.FailureReason
	}
	if !entity.Processing.LastProcessedAt.IsZero() {
		processingMap["last_processed_at"] = entity.Processing.LastProcessedAt
	}
	m["processing"] = processingMap

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

	// Basic fields
	if v, ok := data[fields.ImageWsID.FirestoreName()].(string); ok {
		image.WsID = v
	}
	if v, ok := data[fields.ImageFormat.FirestoreName()].(string); ok {
		image.Format = v
	}

	// Optional basic fields
	if v, ok := data[fields.ImageWidth.FirestoreName()].(int64); ok {
		width := int(v)
		image.Width = &width
	}
	if v, ok := data[fields.ImageHeight.FirestoreName()].(int64); ok {
		height := int(v)
		image.Height = &height
	}

	// Magnification
	if magData, ok := data["magnification"].(map[string]interface{}); ok {
		mag := &vobj.OpticalMagnification{}
		if v, ok := magData["objective"].(float64); ok {
			mag.Objective = &v
		}
		if v, ok := magData["native_level"].(int64); ok {
			level := int(v)
			mag.NativeLevel = &level
		}
		if v, ok := magData["scan_magnification"].(float64); ok {
			mag.ScanMagnification = &v
		}
		image.Magnification = mag
	}

	// Origin content
	if v, ok := data[fields.ImageOriginContentID.FirestoreName()].(string); ok {
		image.OriginContentID = &v
	}

	// Processed content IDs
	if v, ok := data[fields.ImageThumbnailContentID.FirestoreName()].(string); ok {
		image.ThumbnailContentID = &v
	}
	if v, ok := data[fields.ImageDziContentID.FirestoreName()].(string); ok {
		image.DziContentID = &v
	}
	if v, ok := data[fields.ImageIndexmapContentID.FirestoreName()].(string); ok {
		image.IndexmapContentID = &v
	}
	if v, ok := data[fields.ImageTilesContentID.FirestoreName()].(string); ok {
		image.TilesContentID = &v
	}
	if v, ok := data[fields.ImageZipTilesContentID.FirestoreName()].(string); ok {
		image.ZipTilesContentID = &v
	}

	// Processing info
	if procInfo, ok := data["processing"].(map[string]interface{}); ok {
		if statusStr, ok := procInfo["status"].(string); ok {
			image.Processing.Status, err = vobj.NewImageStatusFromString(statusStr)
			if err != nil {
				return nil, err
			}
		}

		if versionStr, ok := procInfo["version"].(string); ok {
			image.Processing.Version = vobj.ProcessingVersion(versionStr)
		}

		if retryCount, ok := procInfo["retry_count"].(int64); ok {
			image.Processing.RetryCount = int(retryCount)
		}

		if failureReason, ok := procInfo["failure_reason"].(string); ok {
			image.Processing.FailureReason = &failureReason
		}

		if lastProcessedAt, ok := procInfo["last_processed_at"].(time.Time); ok {
			image.Processing.LastProcessedAt = lastProcessedAt
		}
	}

	return image, nil
}

func (im *ImageMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates, err := im.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		firestoreField := fields.MapToFirestore(k)

		switch k {
		case fields.ImageWidth.DomainName():
			if width, ok := v.(*int); ok {
				mappedUpdates[firestoreField] = *width
			} else if widthInt, ok := v.(int); ok {
				mappedUpdates[firestoreField] = widthInt
			} else {
				return nil, errors.NewValidationError("invalid type for width field", nil)
			}

		case fields.ImageHeight.DomainName():
			if height, ok := v.(*int); ok {
				mappedUpdates[firestoreField] = *height
			} else if heightInt, ok := v.(int); ok {
				mappedUpdates[firestoreField] = heightInt
			} else {
				return nil, errors.NewValidationError("invalid type for height field", nil)
			}

		case fields.ImageSize.DomainName(): // Not in ImageField?
			if size, ok := v.(*int64); ok {
				mappedUpdates[firestoreField] = *size
			} else if sizeInt64, ok := v.(int64); ok {
				mappedUpdates[firestoreField] = sizeInt64
			} else {
				return nil, errors.NewValidationError("invalid type for size field", nil)
			}

		case fields.ImageMagnification.DomainName():
			if mag, ok := v.(*vobj.OpticalMagnification); ok && mag != nil {
				magMap := make(map[string]interface{})
				if mag.Objective != nil {
					magMap["objective"] = *mag.Objective
				}
				if mag.NativeLevel != nil {
					magMap["native_level"] = *mag.NativeLevel
				}
				if mag.ScanMagnification != nil {
					magMap["scan_magnification"] = *mag.ScanMagnification
				}
				mappedUpdates[firestoreField] = magMap
			} else {
				return nil, errors.NewValidationError("invalid type for magnification field", nil)
			}

		case fields.ImageOriginContentID.DomainName():
			if id, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *id
			} else if idStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = idStr
			} else {
				return nil, errors.NewValidationError("invalid type for origin_content_id field", nil)
			}

		case fields.ImageThumbnailContentID.DomainName():
			if id, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *id
			} else if idStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = idStr
			} else {
				return nil, errors.NewValidationError("invalid type for thumbnail_content_id field", nil)
			}
		case fields.ImageDziContentID.DomainName():
			if id, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *id
			} else if idStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = idStr
			} else {
				return nil, errors.NewValidationError("invalid type for dzi_content_id field", nil)
			}

		case fields.ImageIndexmapContentID.DomainName():
			if id, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *id
			} else if idStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = idStr
			} else {
				return nil, errors.NewValidationError("invalid type for indexmap_content_id field", nil)
			}

		case fields.ImageTilesContentID.DomainName(): // Typo fix: previous code had ImageZipTilesContentIDField separately
			if id, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *id
			} else if idStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = idStr
			} else {
				return nil, errors.NewValidationError("invalid type for tiles_content_id field", nil)
			}

		case fields.ImageZipTilesContentID.DomainName():
			if id, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *id
			} else if idStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = idStr
			} else {
				return nil, errors.NewValidationError("invalid type for ziptiles_content_id field", nil)
			}

		case fields.ImageProcessingStatus.DomainName():
			if status, ok := v.(vobj.ImageStatus); ok {
				mappedUpdates[firestoreField] = status.String()
			} else if statusStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = statusStr
			} else {
				return nil, errors.NewValidationError("invalid type for processing.status field", nil)
			}

		case fields.ImageProcessingVersion.DomainName():
			if version, ok := v.(vobj.ProcessingVersion); ok {
				mappedUpdates[firestoreField] = version.String()
			} else if versionStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = versionStr
			} else {
				return nil, errors.NewValidationError("invalid type for processing.version field", nil)
			}

		case fields.ImageProcessingFailureReason.DomainName():
			if reason, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *reason
			} else if reasonStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = reasonStr
			} else {
				return nil, errors.NewValidationError("invalid type for processing.failure_reason field", nil)
			}

		case fields.ImageProcessingRetryCount.DomainName():
			if retryCount, ok := v.(*int); ok {
				mappedUpdates[firestoreField] = *retryCount
			} else if retryCountInt, ok := v.(int); ok {
				mappedUpdates[firestoreField] = retryCountInt
			} else {
				return nil, errors.NewValidationError("invalid type for processing.retry_count field", nil)
			}

		case fields.ImageProcessingLastProcessedAt.DomainName():
			if lastProcessedAt, ok := v.(*time.Time); ok {
				mappedUpdates[firestoreField] = *lastProcessedAt
			} else if lastProcessedAtTime, ok := v.(time.Time); ok {
				mappedUpdates[firestoreField] = lastProcessedAtTime
			} else {
				return nil, errors.NewValidationError("invalid type for processing.last_processed_at field", nil)
			}

		case fields.ImageWsID.DomainName():
			if wsID, ok := v.(string); ok {
				mappedUpdates[firestoreField] = wsID
			} else {
				return nil, errors.NewValidationError("invalid type for ws_id field", nil)
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

	for _, f := range filters {
		firestoreField := fields.MapToFirestore(f.Field)
		if fields.ImageField(f.Field).IsValid() {
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    firestoreField,
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return firestoreFilters, nil
}
