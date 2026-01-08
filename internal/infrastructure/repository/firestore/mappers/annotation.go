package firestoreMappers

import (
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationMapper struct {
	entityMapper *EntityMapper
}

func NewAnnotationMapper() *AnnotationMapper {
	return &AnnotationMapper{
		entityMapper: &EntityMapper{},
	}
}

func (am *AnnotationMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.Annotation, error) {
	data := doc.Data()

	if data == nil {
		return nil, fmt.Errorf("firestore document data is nil")
	}

	// Base entity'yi parse et
	entity, err := am.entityMapper.FromFirestoreDoc(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to map entity from firestore document: %w", err)
	}

	// Polygon'u parse et
	polygon := []vobj.Point{}
	if pts, ok := data["polygon"].([]interface{}); ok {
		polygon = make([]vobj.Point, 0, len(pts))
		for _, pt := range pts {
			if pointMap, ok := pt.(map[string]interface{}); ok {
				x, xOk := pointMap["x"].(float64)
				y, yOk := pointMap["y"].(float64)
				if xOk && yOk {
					polygon = append(polygon, vobj.Point{X: x, Y: y})
				}
			}
		}
	}

	// Tag'i parse et
	tag, err := am.parseTag(data)
	if err != nil {
		return nil, err
	}

	return &model.Annotation{
		Entity:  entity,
		Polygon: polygon,
		Tag:     *tag,
	}, nil
}

func (am *AnnotationMapper) parseTag(data map[string]interface{}) (*vobj.TagValue, error) {
	tagTypeStr, ok := data["tag_type"].(string)
	if !ok {
		return nil, errors.NewValidationError("tag_type is required", nil)
	}
	tagType := vobj.TagType(tagTypeStr)

	tagName, ok := data["tag_name"].(string)
	if !ok {
		return nil, errors.NewValidationError("tag_name is required", nil)
	}

	value := data["tag_value"]

	var color *string
	if colorVal, ok := data["tag_color"].(string); ok && colorVal != "" {
		color = &colorVal
	}

	global := false
	if g, ok := data["tag_global"].(bool); ok {
		global = g
	}

	return vobj.NewTagValue(tagType, tagName, value, color, global)
}

func (am *AnnotationMapper) ToFirestoreMap(annotation *model.Annotation) map[string]interface{} {
	// Base entity map'ini al
	m := am.entityMapper.ToFirestoreMap(annotation.Entity)

	// Polygon'u serialize et
	if len(annotation.Polygon) > 0 {
		points := make([]map[string]float64, len(annotation.Polygon))
		for i, p := range annotation.Polygon {
			points[i] = map[string]float64{
				"x": p.X,
				"y": p.Y,
			}
		}
		m["polygon"] = points
	} else {
		m["polygon"] = []map[string]float64{}
	}

	// Tag'i serialize et
	m["tag_type"] = annotation.Tag.TagType.String()
	m["tag_name"] = annotation.Tag.TagName
	m["tag_value"] = annotation.Tag.Value
	m["tag_global"] = annotation.Tag.Global
	if annotation.Tag.Color != nil {
		m["tag_color"] = *annotation.Tag.Color
	}

	return m
}

func (am *AnnotationMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	// Base entity updates'lerini al
	firestoreUpdates, err := am.entityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	for k, v := range updates {
		switch k {
		case constants.PolygonField:
			points, ok := v.([]vobj.Point)
			if !ok {
				return nil, errors.NewValidationError("invalid polygon type", nil)
			}
			firestorePoints := make([]map[string]float64, len(points))
			for i, p := range points {
				firestorePoints[i] = map[string]float64{
					"x": p.X,
					"y": p.Y,
				}
			}
			firestoreUpdates["polygon"] = firestorePoints
			delete(updates, constants.PolygonField)

		case constants.TagField:
			tagValue, ok := v.(vobj.TagValue)
			if !ok {
				return nil, errors.NewValidationError("invalid tag type", nil)
			}
			firestoreUpdates["tag_type"] = tagValue.TagType.String()
			firestoreUpdates["tag_name"] = tagValue.TagName
			firestoreUpdates["tag_value"] = tagValue.Value
			firestoreUpdates["tag_global"] = tagValue.Global
			if tagValue.Color != nil {
				firestoreUpdates["tag_color"] = *tagValue.Color
			}
			delete(updates, constants.TagField)

		case constants.TagValueField:
			// Tag value'yu tek başına güncellemek için
			firestoreUpdates["tag_value"] = v
			delete(updates, constants.TagValueField)
		}
	}

	return firestoreUpdates, nil
}

func (am *AnnotationMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	// Base entity filters'ları map'le
	mappedFilters, err := am.entityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	unprocessedIdx := 0
	for i, filter := range filters {
		processed := false

		switch filter.Field {
		case constants.TagNameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_name",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.TagTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_type",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.TagValueField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_value",
				Operator: filter.Operator,
				Value:    filter.Value,
			})
			processed = true

		case constants.TagGlobalField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_global",
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

	// İşlenmemiş filtreleri temizle
	for i := unprocessedIdx; i < len(filters); i++ {
		filters[i] = query.Filter{}
	}

	return mappedFilters, nil
}
