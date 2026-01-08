package firestoreMappers

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeMapper struct {
	entityMapper *EntityMapper
}

func NewAnnotationTypeMapper() *AnnotationTypeMapper {
	return &AnnotationTypeMapper{
		entityMapper: &EntityMapper{},
	}
}

func (atm *AnnotationTypeMapper) FromFirestoreDoc(doc *firestore.DocumentSnapshot) (*model.AnnotationType, error) {
	// Parse base entity
	entity, err := atm.entityMapper.FromFirestoreDoc(doc)
	if err != nil {
		return nil, err
	}

	data := doc.Data()

	// Parse tags
	var tags []vobj.Tag
	if tagsData, ok := data["tags"].([]interface{}); ok {
		tags = make([]vobj.Tag, 0, len(tagsData))
		for _, tagInterface := range tagsData {
			tagMap, ok := tagInterface.(map[string]interface{})
			if !ok {
				continue
			}

			tag, err := atm.parseTag(tagMap)
			if err != nil {
				return nil, err
			}
			tags = append(tags, *tag)
		}
	}

	return &model.AnnotationType{
		Entity: entity,
		Tags:   tags,
	}, nil
}

func (atm *AnnotationTypeMapper) parseTag(tagMap map[string]interface{}) (*vobj.Tag, error) {
	name := tagMap["name"].(string)
	tagTypeStr := tagMap["type"].(string)
	tagType := vobj.TagType(tagTypeStr)

	// Parse options
	var options []string
	if optionsData, ok := tagMap["options"].([]interface{}); ok {
		options = make([]string, len(optionsData))
		for i, opt := range optionsData {
			options[i] = opt.(string)
		}
	}

	global := false
	if g, ok := tagMap["global"].(bool); ok {
		global = g
	}

	required := false
	if r, ok := tagMap["required"].(bool); ok {
		required = r
	}

	// Min/Max değerleri
	var min, max *float64
	if minVal, ok := tagMap["min"].(float64); ok {
		min = &minVal
	}
	if maxVal, ok := tagMap["max"].(float64); ok {
		max = &maxVal
	}

	// Color
	var color *string
	if colorVal, ok := tagMap["color"].(string); ok {
		color = &colorVal
	}

	return vobj.NewTag(name, tagType, options, global, required, min, max, color)
}

func (atm *AnnotationTypeMapper) ToFirestoreMap(annotationType *model.AnnotationType) map[string]interface{} {
	m := atm.entityMapper.ToFirestoreMap(annotationType.Entity)

	if len(annotationType.Tags) > 0 {
		tagsData := make([]map[string]interface{}, len(annotationType.Tags))
		for i, tag := range annotationType.Tags {
			tagsData[i] = atm.tagToMap(&tag)
		}
		m["tags"] = tagsData
	} else {
		m["tags"] = []map[string]interface{}{}
	}

	return m
}

func (atm *AnnotationTypeMapper) tagToMap(tag *vobj.Tag) map[string]interface{} {
	m := map[string]interface{}{
		"name":     tag.Name,
		"type":     tag.Type.String(),
		"global":   tag.Global,
		"required": tag.Required,
	}

	if len(tag.Options) > 0 {
		m["options"] = tag.Options
	}

	if tag.Min != nil {
		m["min"] = *tag.Min
	}
	if tag.Max != nil {
		m["max"] = *tag.Max
	}

	if tag.Color != nil {
		m["color"] = *tag.Color
	}

	return m
}

func (atm *AnnotationTypeMapper) MapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	firestoreUpdates, err := atm.entityMapper.MapUpdates(updates)
	if err != nil {
		return nil, err
	}

	if tagsValue, ok := updates[constants.TagsField]; ok {
		tags, ok := tagsValue.([]vobj.Tag)
		if !ok {
			return nil, errors.NewValidationError("invalid tags type", nil)
		}

		tagsData := make([]map[string]interface{}, len(tags))
		for i, tag := range tags {
			tagsData[i] = atm.tagToMap(&tag)
		}
		firestoreUpdates["tags"] = tagsData
		delete(updates, constants.TagsField)
	}

	return firestoreUpdates, nil
}

func (atm *AnnotationTypeMapper) MapFilters(filters []query.Filter) ([]query.Filter, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	// Mapped base entity filters
	mappedFilters, err := atm.entityMapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	// Annotation type specific filters
	// Note: Due to limitations of nested array queries in Firestore,
	// special approaches may be needed for tag-based filtering

	return mappedFilters, nil
}
