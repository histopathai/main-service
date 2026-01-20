package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationMapper struct {
	*EntityMapper[*model.Annotation]
}

func NewAnnotationMapper() *AnnotationMapper {
	return &AnnotationMapper{
		EntityMapper: NewEntityMapper[*model.Annotation](),
	}
}

func (am *AnnotationMapper) ToFirestoreMap(entity *model.Annotation) map[string]interface{} {
	m := am.EntityMapper.ToFirestoreMap(entity)

	if entity.Polygon != nil {
		m["polygon"] = vobj.ToJSONPoints(*entity.Polygon)
	}
	m["value"] = entity.Value
	m["tag_type"] = entity.TagType.String()
	m["is_global"] = entity.IsGlobal
	if entity.Color != nil {
		m["color"] = *entity.Color
	}
	m["ws_id"] = entity.WsID

	return m
}

func (am *AnnotationMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Annotation, error) {
	entity, err := am.EntityMapper.ParseEntity(doc)
	if err != nil {
		return nil, err
	}

	annotation := &model.Annotation{
		Entity: *entity,
	}

	data := doc.Data()

	if polygonRaw, ok := data["polygon"].([]interface{}); ok {
		jsonPoints := make([]map[string]float64, 0, len(polygonRaw))
		for _, p := range polygonRaw {
			if pointMap, ok := p.(map[string]interface{}); ok {
				jsonPoint := make(map[string]float64)
				if x, ok := pointMap["X"].(float64); ok {
					jsonPoint["X"] = x
				}
				if y, ok := pointMap["Y"].(float64); ok {
					jsonPoint["Y"] = y
				}
				jsonPoints = append(jsonPoints, jsonPoint)
			}
		}
		points := vobj.FromJSONPoints(jsonPoints)
		annotation.Polygon = &points
	}

	annotation.Value = data["value"]

	if tagTypeStr, ok := data["tag_type"].(string); ok {
		annotation.TagType, err = vobj.NewTagTypeFromString(tagTypeStr)
		if err != nil {
			return nil, err
		}
	}

	if isGlobal, ok := data["is_global"].(bool); ok {
		annotation.IsGlobal = isGlobal
	}

	if color, ok := data["color"].(string); ok {
		annotation.Color = &color
	}

	if wsID, ok := data["ws_id"].(string); ok {
		annotation.WsID = wsID
	}

	return annotation, nil
}

func (am *AnnotationMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	mappedUpdates, err := am.EntityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		switch k {
		case constants.PolygonField:
			if polygon, ok := v.(*[]vobj.Point); ok {
				mappedUpdates["polygon"] = vobj.ToJSONPoints(*polygon)
			} else if points, ok := v.([]vobj.Point); ok {
				mappedUpdates["polygon"] = vobj.ToJSONPoints(points)
			} else {
				return nil, errors.NewValidationError("invalid type for polygon field", nil)
			}

		case constants.TagValueField:
			mappedUpdates["value"] = v

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

		case constants.TagColorField:
			if color, ok := v.(*string); ok {
				mappedUpdates["color"] = *color
			} else if colorStr, ok := v.(string); ok {
				mappedUpdates["color"] = colorStr
			} else {
				return nil, errors.NewValidationError("invalid type for color field", nil)
			}
		case constants.WsIDField:
			if wsID, ok := v.(string); ok {
				mappedUpdates["ws_id"] = wsID
			} else {
				return nil, errors.NewValidationError("invalid type for ws_id field", nil)
			}
		}
	}

	return mappedUpdates, nil
}

func (am *AnnotationMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters, err := am.EntityMapper.MapFilters(filters)
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
		case constants.TagColorField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "color",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.TagValueField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "value",
				Operator: f.Operator,
				Value:    f.Value,
			})
		case constants.WsIDField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "ws_id",
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return mappedFilters, nil
}
