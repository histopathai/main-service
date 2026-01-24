package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ContentMapper struct {
	*EntityMapper[*model.Content]
}

func NewContentMapper() *ContentMapper {
	return &ContentMapper{
		EntityMapper: NewEntityMapper[*model.Content](),
	}
}

func (cm *ContentMapper) ToFirestoreMap(entity *model.Content) map[string]interface{} {
	m := cm.EntityMapper.ToFirestoreMap(entity)

	m["provider"] = entity.Provider.String()
	m["path"] = entity.Path
	m["content_type"] = entity.ContentType.String()
	m["size"] = entity.Size

	return m
}

func (cm *ContentMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Content, error) {
	entity, err := cm.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}

	content := &model.Content{
		Entity: *entity,
	}

	data := doc.Data()

	if v, ok := data["provider"].(string); ok {
		content.Provider = vobj.ContentProvider(v)
	}

	if v, ok := data["path"].(string); ok {
		content.Path = v
	}

	if v, ok := data["content_type"].(string); ok {
		content.ContentType = vobj.ContentType(v)
	}

	if v, ok := data["size"].(int64); ok {
		content.Size = v
	}

	return content, nil
}

func (cm *ContentMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates, err := cm.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		switch k {
		case constants.ContentProviderField:
			if provider, ok := v.(vobj.ContentProvider); ok {
				mappedUpdates["provider"] = provider.String()
			} else if providerStr, ok := v.(string); ok {
				mappedUpdates["provider"] = providerStr
			} else {
				return nil, errors.NewValidationError("invalid type for provider field", nil)
			}

		case constants.ContentPathField:
			if path, ok := v.(*string); ok {
				mappedUpdates["path"] = *path
			} else if pathStr, ok := v.(string); ok {
				mappedUpdates["path"] = pathStr
			} else {
				return nil, errors.NewValidationError("invalid type for path field", nil)
			}

		case constants.ContentTypeField:
			if contentType, ok := v.(vobj.ContentType); ok {
				mappedUpdates["content_type"] = contentType.String()
			} else if contentTypeStr, ok := v.(string); ok {
				mappedUpdates["content_type"] = contentTypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for content_type field", nil)
			}

		case constants.ContentSizeField:
			if size, ok := v.(*int64); ok {
				mappedUpdates["size"] = *size
			} else if sizeInt, ok := v.(int64); ok {
				mappedUpdates["size"] = sizeInt
			} else {
				return nil, errors.NewValidationError("invalid type for size field", nil)
			}
		}
	}

	return mappedUpdates, nil
}

func (cm *ContentMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	firestoreFilters, err := cm.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	for _, f := range filters {
		switch f.Field {
		case constants.ContentProviderField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "provider",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ContentTypeField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "content_type",
				Operator: f.Operator,
				Value:    f.Value,
			})

		case constants.ContentSizeField:
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    "size",
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return firestoreFilters, nil
}
