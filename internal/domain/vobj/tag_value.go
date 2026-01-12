package vobj

import "github.com/histopathai/main-service/internal/shared/errors"

type TagValue struct {
	Type   TagType
	Value  any
	Color  *string
	Global bool
}

func (tv *TagValue) IsGlobal() bool {
	if tv == nil {
		return false
	}
	return tv.Global
}

func (tv *TagValue) GetType() TagType {
	if tv == nil {
		return ""
	}
	return tv.Type
}

func (tv *TagValue) GetValue() any {
	if tv == nil {
		return nil
	}
	return tv.Value
}

func (tv *TagValue) GetColor() *string {
	if tv == nil {
		return nil
	}
	return tv.Color
}

func NewTagValue(tagType TagType, value any, color *string, global bool) (*TagValue, error) {
	if !tagType.IsValid() {
		details := map[string]any{"tag_type": tagType}
		return nil, errors.NewValidationError("invalid tag type", details)
	}

	return &TagValue{
		Type:   tagType,
		Value:  value,
		Color:  color,
		Global: global,
	}, nil
}
