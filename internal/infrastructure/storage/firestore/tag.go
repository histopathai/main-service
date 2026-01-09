package firestore

import (
	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

func TagToFirestoreMap(t *vobj.Tag) map[string]interface{} {
	m := make(map[string]interface{})
	m["tag_type"] = t.Type
	m["tag_name"] = t.Name
	m["tag_options"] = t.Options
	m["is_global"] = t.Global
	m["is_required"] = t.Required
	if t.Type == vobj.NumberTag {
		if t.Min != nil {
			m["min"] = *t.Min
		}
		if t.Max != nil {
			m["max"] = *t.Max
		}
	}
	if t.Color != nil {
		m["tag_color"] = *t.Color
	}
	return m
}

func TagFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*vobj.Tag, error) {
	data := doc.Data()

	t := &vobj.Tag{}

	if v, ok := data["tag_name"].(string); ok {
		t.Name = v
	}

	if v, ok := data["tag_type"].(vobj.TagType); ok {
		t.Type = v
	}

	if v, ok := data["tag_options"].([]string); ok {
		t.Options = v
	} else if v, ok := data["tag_options"].([]interface{}); ok {
		for _, item := range v {
			if strItem, ok := item.(string); ok {
				t.Options = append(t.Options, strItem)
			}
		}
	}

	if v, ok := data["is_global"].(bool); ok {
		t.Global = v
	}

	if v, ok := data["is_required"].(bool); ok {
		t.Required = v
	}

	if v, ok := data["min"].(float64); ok {
		t.Min = &v
	}

	if v, ok := data["max"].(float64); ok {
		t.Max = &v
	}

	if v, ok := data["tag_color"].(string); ok {
		t.Color = &v
	}

	return t, nil
}

func TagMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.TagNameField:
			firestoreUpdates["tag_name"] = value
			delete(updates, key)
		case constants.TagTypeField:
			firestoreUpdates["tag_type"] = value
			delete(updates, key)
		case constants.TagOptionsField:
			firestoreUpdates["tag_options"] = value
			delete(updates, key)
		case constants.TagGlobalField:
			firestoreUpdates["is_global"] = value
			delete(updates, key)
		case constants.TagRequiredField:
			firestoreUpdates["is_required"] = value
			delete(updates, key)
		case constants.TagMinField:
			firestoreUpdates["min"] = value
			delete(updates, key)
		case constants.TagMaxField:
			firestoreUpdates["max"] = value
			delete(updates, key)
		case constants.TagColorField:
			firestoreUpdates["tag_color"] = value
			delete(updates, key)
		default:
			continue
			//ignore unknown fields
		}

	}
	return firestoreUpdates, nil
}

func TagMapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))
	processedIndices := make(map[int]bool)
	for i, f := range filters {
		switch f.Field {
		case constants.TagNameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_name",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagOptionsField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_options",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_type",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagGlobalField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_global",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagRequiredField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_required",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagColorField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_color",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true

		case constants.TagMaxField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "max",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagMinField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "min",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		default:
			continue
			//ignore unknown fields
		}
	}
	// Remove processed filters from the original slice
	for i := len(filters) - 1; i >= 0; i-- {
		if processedIndices[i] {
			filters = append(filters[:i], filters[i+1:]...)
		}
	}
	return mappedFilters, nil
}

func TagValueToFirestoreMap(tv *vobj.TagValue) map[string]interface{} {
	m := make(map[string]interface{})
	m["tag_name"] = tv.TagName
	m["tag_type"] = tv.TagType
	m["tag_value"] = tv.Value
	if tv.Color != nil {
		m["tag_color"] = *tv.Color
	}
	m["is_global"] = tv.Global
	return m
}

func TagValueFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*vobj.TagValue, error) {
	data := doc.Data()

	a := vobj.TagValue{}

	if v, ok := data["tag_name"].(string); ok {
		a.TagName = v
	}

	if v, ok := data["tag_type"].(vobj.TagType); ok {
		a.TagType = v
	}

	if v, ok := data["tag_value"]; ok {
		a.Value = v
	}

	if v, ok := data["tag_color"].(string); ok {
		a.Color = &v
	}

	if v, ok := data["is_global"].(bool); ok {
		a.Global = v
	}

	return &a, nil
}

func TagValueMapUpdates(updates map[string]interface{}) (map[string]interface{}, error) {
	firestoreUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case constants.TagNameField:
			firestoreUpdates["tag_name"] = value
			delete(updates, key)
		case constants.TagTypeField:
			firestoreUpdates["tag_type"] = value
			delete(updates, key)
		case constants.TagValueField:
			firestoreUpdates["tag_value"] = value
			delete(updates, key)
		case constants.TagColorField:
			firestoreUpdates["tag_color"] = value
			delete(updates, key)
		case constants.TagGlobalField:
			firestoreUpdates["is_global"] = value
			delete(updates, key)
		default:
			continue
			//ignore unknown fields
		}

	}
	return firestoreUpdates, nil
}

func TagValueMapFilters(filters []query.Filter) ([]query.Filter, error) {
	mappedFilters := make([]query.Filter, 0, len(filters))
	processedIndices := make(map[int]bool)
	for i, f := range filters {
		switch f.Field {
		case constants.TagNameField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_name",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagTypeField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_type",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagValueField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_value",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagColorField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "tag_color",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		case constants.TagGlobalField:
			mappedFilters = append(mappedFilters, query.Filter{
				Field:    "is_global",
				Operator: f.Operator,
				Value:    f.Value,
			})
			processedIndices[i] = true
		default:
			continue
			//ignore unknown fields
		}
	}
	// Remove processed filters from the original slice
	for i := len(filters) - 1; i >= 0; i-- {
		if processedIndices[i] {
			filters = append(filters[:i], filters[i+1:]...)
		}
	}
	return mappedFilters, nil
}
