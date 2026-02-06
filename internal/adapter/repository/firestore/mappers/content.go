package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
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

	m[fields.ContentProvider.FirestoreName()] = entity.Provider.String()
	m[fields.ContentPath.FirestoreName()] = entity.Path
	m[fields.ContentType.FirestoreName()] = entity.ContentType.String()
	m[fields.ContentSize.FirestoreName()] = entity.Size
	m[fields.ContentUploadPending.FirestoreName()] = entity.UploadPending

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

	if v, ok := data[fields.ContentProvider.FirestoreName()].(string); ok {
		content.Provider = vobj.ContentProvider(v)
	}

	if v, ok := data[fields.ContentPath.FirestoreName()].(string); ok {
		content.Path = v
	}

	if v, ok := data[fields.ContentType.FirestoreName()].(string); ok {
		content.ContentType = vobj.ContentType(v)
	}

	if v, ok := data[fields.ContentSize.FirestoreName()].(int64); ok {
		content.Size = v
	}
	if v, ok := data[fields.ContentUploadPending.FirestoreName()].(bool); ok {
		content.UploadPending = v
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
		case fields.ContentProvider.DomainName():
			if provider, ok := v.(vobj.ContentProvider); ok {
				mappedUpdates[fields.ContentProvider.FirestoreName()] = provider.String()
			} else if providerStr, ok := v.(string); ok {
				mappedUpdates[fields.ContentProvider.FirestoreName()] = providerStr
			} else {
				return nil, errors.NewValidationError("invalid type for provider field", nil)
			}

		case fields.ContentPath.DomainName():
			if path, ok := v.(*string); ok {
				mappedUpdates[fields.ContentPath.FirestoreName()] = *path
			} else if pathStr, ok := v.(string); ok {
				mappedUpdates[fields.ContentPath.FirestoreName()] = pathStr
			} else {
				return nil, errors.NewValidationError("invalid type for path field", nil)
			}

		case fields.ContentType.DomainName():
			if contentType, ok := v.(vobj.ContentType); ok {
				mappedUpdates[fields.ContentType.FirestoreName()] = contentType.String()
			} else if contentTypeStr, ok := v.(string); ok {
				mappedUpdates[fields.ContentType.FirestoreName()] = contentTypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for content_type field", nil)
			}

		case fields.ContentSize.DomainName():
			if size, ok := v.(*int64); ok {
				mappedUpdates[fields.ContentSize.FirestoreName()] = *size
			} else if sizeInt, ok := v.(int64); ok {
				mappedUpdates[fields.ContentSize.FirestoreName()] = sizeInt
			} else {
				return nil, errors.NewValidationError("invalid type for size field", nil)
			}

		case fields.ContentUploadPending.DomainName():
			if uploadPending, ok := v.(bool); ok {
				mappedUpdates[fields.ContentUploadPending.FirestoreName()] = uploadPending
			} else {
				return nil, errors.NewValidationError("invalid type for upload_pending field", nil)
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
		firestoreField := fields.MapToFirestore(f.Field)
		if fields.ContentField(f.Field).IsValid() {
			firestoreFilters = append(firestoreFilters, query.Filter{
				Field:    firestoreField,
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return firestoreFilters, nil
}
