package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
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

	m["tag_type"] = entity.TagType.String()
	m["is_global"] = entity.IsGlobal
	m["is_required"] = entity.IsRequired

	if len(entity.Options) > 0 {
		m["options"] = entity.Options
	}
	if entity.Min != nil {
		m["min"] = *entity.Min
	}
	if entity.Max != nil {
		m["max"] = *entity.Max
	}
	if entity.Color != nil {
		m["color"] = *entity.Color
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

	if tagTypeStr, ok := data["tag_type"].(string); ok {
		annotationType.TagType, err = vobj.NewTagTypeFromString(tagTypeStr)
		if err != nil {
			return nil, err
		}
	}

	if isGlobal, ok := data["is_global"].(bool); ok {
		annotationType.IsGlobal = isGlobal
	}

	if isRequired, ok := data["is_required"].(bool); ok {
		annotationType.IsRequired = isRequired
	}

	if optionsRaw, ok := data["options"].([]interface{}); ok {
		options := make([]string, 0, len(optionsRaw))
		for _, opt := range optionsRaw {
			if optStr, ok := opt.(string); ok {
				options = append(options, optStr)
			}
		}
		annotationType.Options = options
	}

	if min, ok := data["min"].(float64); ok {
		annotationType.Min = &min
	}

	if max, ok := data["max"].(float64); ok {
		annotationType.Max = &max
	}

	if color, ok := data["color"].(string); ok {
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
		switch k {
		case constants.TagTypeField:
			if tagType, ok := v.(vobj.TagType); ok {
				mappedUpdates["tag_type"] = tagType.String()
			} else if tagTypeStr, ok := v.(string); ok {
				mappedUpdates["tag_type"] = tagTypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for tag_type field", nil)
			}

		case constants.TagGlobalField:
			if isGlobal, ok := v.(bool); ok {
				mappedUpdates["is_global"] = isGlobal
			} else {
				return nil, errors.NewValidationError("invalid type for is_global field", nil)
			}

		case constants.TagRequiredField:
			if isRequired, ok := v.(bool); ok {
				mappedUpdates["is_required"] = isRequired
			} else {
				return nil, errors.NewValidationError("invalid type for is_required field", nil)
			}

		case constants.TagOptionsField:
			if options, ok := v.([]string); ok {
				mappedUpdates["options"] = options
			} else {
				return nil, errors.NewValidationError("invalid type for options field", nil)
			}

		case constants.TagMinField:
			if min, ok := v.(*float64); ok {
				mappedUpdates["min"] = *min
			} else if minFloat, ok := v.(float64); ok {
				mappedUpdates["min"] = minFloat
			} else {
				return nil, errors.NewValidationError("invalid type for min field", nil)
			}

		case constants.TagMaxField:
			if max, ok := v.(*float64); ok {
				mappedUpdates["max"] = *max
			} else if maxFloat, ok := v.(float64); ok {
				mappedUpdates["max"] = maxFloat
			} else {
				return nil, errors.NewValidationError("invalid type for max field", nil)
			}

		case constants.TagColorField:
			if color, ok := v.(*string); ok {
				mappedUpdates["color"] = *color
			} else if colorStr, ok := v.(string); ok {
				mappedUpdates["color"] = colorStr
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
		switch f.Field {
		case constants.TagTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_type",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagGlobalField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_global",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagRequiredField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_required",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagOptionsField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "options",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagMinField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "min",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagMaxField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "max",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagColorField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "color",
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return mappedFilters, nil
}
