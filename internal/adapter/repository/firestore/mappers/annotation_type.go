package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeMapper struct {
	*EntityMapper[*model.AnnotationType]
}

func NewAnnotationTypeMapper() *AnnotationTypeMapper {
	return &AnnotationTypeMapper{
		EntityMapper: NewEntityMapper[*model.AnnotationType](),
	}
}

func (atm *AnnotationTypeMapper) ToFirestoreMap(entity *model.AnnotationType) map[string]interface{} {
	m := atm.EntityMapper.ToFirestoreMap(entity)

	m[fields.AnnotationTypeTagType.FirestoreName()] = entity.TagType.String()
	m[fields.AnnotationTypeIsGlobal.FirestoreName()] = entity.IsGlobal
	m[fields.AnnotationTypeIsRequired.FirestoreName()] = entity.IsRequired

	if len(entity.Options) > 0 {
		m[fields.AnnotationTypeOptions.FirestoreName()] = entity.Options
	}
	if entity.Min != nil {
		m[fields.AnnotationTypeMin.FirestoreName()] = *entity.Min
	}
	if entity.Max != nil {
		m[fields.AnnotationTypeMax.FirestoreName()] = *entity.Max
	}
	if entity.Color != nil {
		m[fields.AnnotationTypeColor.FirestoreName()] = *entity.Color
	}

	return m
}

func (atm *AnnotationTypeMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationType, error) {
	entity, err := atm.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}

	annotationType := &model.AnnotationType{
		Entity: *entity,
	}

	data := doc.Data()

	if tagTypeStr, ok := data[fields.AnnotationTypeTagType.FirestoreName()].(string); ok {
		annotationType.TagType, err = vobj.NewTagTypeFromString(tagTypeStr)
		if err != nil {
			return nil, err
		}
	}

	if isGlobal, ok := data[fields.AnnotationTypeIsGlobal.FirestoreName()].(bool); ok {
		annotationType.IsGlobal = isGlobal
	}

	if isRequired, ok := data[fields.AnnotationTypeIsRequired.FirestoreName()].(bool); ok {
		annotationType.IsRequired = isRequired
	}

	if optionsRaw, ok := data[fields.AnnotationTypeOptions.FirestoreName()].([]interface{}); ok {
		options := make([]string, 0, len(optionsRaw))
		for _, opt := range optionsRaw {
			if optStr, ok := opt.(string); ok {
				options = append(options, optStr)
			}
		}
		annotationType.Options = options
	}

	if min, ok := data[fields.AnnotationTypeMin.FirestoreName()].(float64); ok {
		annotationType.Min = &min
	}

	if max, ok := data[fields.AnnotationTypeMax.FirestoreName()].(float64); ok {
		annotationType.Max = &max
	}

	if color, ok := data[fields.AnnotationTypeColor.FirestoreName()].(string); ok {
		annotationType.Color = &color
	}

	return annotationType, nil
}

func (atm *AnnotationTypeMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates, err := atm.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		firestoreField := fields.MapToFirestore(k)

		switch k {
		case fields.AnnotationTypeTagType.DomainName():
			if tagType, ok := v.(vobj.TagType); ok {
				mappedUpdates[firestoreField] = tagType.String()
			} else if tagTypeStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = tagTypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for tag_type field", nil)
			}

		case fields.AnnotationTypeIsGlobal.DomainName():
			if isGlobal, ok := v.(bool); ok {
				mappedUpdates[firestoreField] = isGlobal
			} else {
				return nil, errors.NewValidationError("invalid type for is_global field", nil)
			}

		case fields.AnnotationTypeIsRequired.DomainName():
			if isRequired, ok := v.(bool); ok {
				mappedUpdates[firestoreField] = isRequired
			} else {
				return nil, errors.NewValidationError("invalid type for is_required field", nil)
			}

		case fields.AnnotationTypeOptions.DomainName():
			if options, ok := v.([]string); ok {
				mappedUpdates[firestoreField] = options
			} else {
				return nil, errors.NewValidationError("invalid type for options field", nil)
			}

		case fields.AnnotationTypeMin.DomainName():
			if min, ok := v.(*float64); ok {
				mappedUpdates[firestoreField] = *min
			} else if minFloat, ok := v.(float64); ok {
				mappedUpdates[firestoreField] = minFloat
			} else {
				return nil, errors.NewValidationError("invalid type for min field", nil)
			}

		case fields.AnnotationTypeMax.DomainName():
			if max, ok := v.(*float64); ok {
				mappedUpdates[firestoreField] = *max
			} else if maxFloat, ok := v.(float64); ok {
				mappedUpdates[firestoreField] = maxFloat
			} else {
				return nil, errors.NewValidationError("invalid type for max field", nil)
			}

		case fields.AnnotationTypeColor.DomainName():
			if color, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *color
			} else if colorStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = colorStr
			} else {
				return nil, errors.NewValidationError("invalid type for color field", nil)
			}
		}
	}

	return mappedUpdates, nil
}

func (atm *AnnotationTypeMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters, err := atm.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	for _, f := range filters {
		firestoreField := fields.MapToFirestore(f.Field)
		if fields.AnnotationTypeField(f.Field).IsValid() {
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    firestoreField,
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return mappedFilters, nil
}
