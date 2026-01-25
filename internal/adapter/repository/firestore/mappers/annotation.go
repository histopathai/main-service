package mappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
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
		m[fields.AnnotationPolygon.FirestoreName()] = vobj.ToJSONPoints(*entity.Polygon)
	}
	m[fields.AnnotationTagValue.FirestoreName()] = entity.Value
	m[fields.AnnotationTagType.FirestoreName()] = entity.TagType.String()
	m[fields.AnnotationIsGlobal.FirestoreName()] = entity.IsGlobal
	if entity.Color != nil {
		m[fields.AnnotationColor.FirestoreName()] = *entity.Color
	}
	// WsID is not in AnnotationFields, borrowing from ImageFields or using string literal.
	// Using ImageWsID for consistency if allowable, otherwise string "ws_id".
	// The file uses "ws_id". field_set.go showed ImageWsID = "ws_id".
	m[fields.ImageWsID.FirestoreName()] = entity.WsID

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

	if polygonRaw, ok := data[fields.AnnotationPolygon.FirestoreName()].([]interface{}); ok {
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

	annotation.Value = data[fields.AnnotationTagValue.FirestoreName()]

	if tagTypeStr, ok := data[fields.AnnotationTagType.FirestoreName()].(string); ok {
		annotation.TagType, err = vobj.NewTagTypeFromString(tagTypeStr)
		if err != nil {
			return nil, err
		}
	}

	if isGlobal, ok := data[fields.AnnotationIsGlobal.FirestoreName()].(bool); ok {
		annotation.IsGlobal = isGlobal
	}

	if color, ok := data[fields.AnnotationColor.FirestoreName()].(string); ok {
		annotation.Color = &color
	}

	if wsID, ok := data[fields.ImageWsID.FirestoreName()].(string); ok {
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
		firestoreField := fields.MapToFirestore(k)

		switch k {
		case fields.AnnotationPolygon.DomainName():
			if polygon, ok := v.(*[]vobj.Point); ok {
				mappedUpdates[firestoreField] = vobj.ToJSONPoints(*polygon)
			} else if points, ok := v.([]vobj.Point); ok {
				mappedUpdates[firestoreField] = vobj.ToJSONPoints(points)
			} else {
				return nil, errors.NewValidationError("invalid type for polygon field", nil)
			}

		case fields.AnnotationTagValue.DomainName():
			mappedUpdates[firestoreField] = v

		case fields.AnnotationTagType.DomainName():
			if tagType, ok := v.(vobj.TagType); ok {
				mappedUpdates[firestoreField] = tagType.String()
			} else if tagTypeStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = tagTypeStr
			} else {
				return nil, errors.NewValidationError("invalid type for tag_type field", nil)
			}

		case fields.AnnotationIsGlobal.DomainName():
			if isGlobal, ok := v.(bool); ok {
				mappedUpdates[firestoreField] = isGlobal
			} else {
				return nil, errors.NewValidationError("invalid type for is_global field", nil)
			}

		case fields.AnnotationColor.DomainName():
			if color, ok := v.(*string); ok {
				mappedUpdates[firestoreField] = *color
			} else if colorStr, ok := v.(string); ok {
				mappedUpdates[firestoreField] = colorStr
			} else {
				return nil, errors.NewValidationError("invalid type for color field", nil)
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

func (am *AnnotationMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters, err := am.EntityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	for _, f := range filters {
		firestoreField := fields.MapToFirestore(f.Field)
		if fields.AnnotationField(f.Field).IsValid() {
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    firestoreField,
				Operator: f.Operator,
				Value:    f.Value,
			})
		} else if f.Field == fields.ImageWsID.APIName() || f.Field == fields.ImageWsID.DomainName() {
			// Handle ws_id explicitly if not in AnnotationFields
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    fields.ImageWsID.FirestoreName(),
				Operator: f.Operator,
				Value:    f.Value,
			})
		}
	}

	return mappedFilters, nil
}
