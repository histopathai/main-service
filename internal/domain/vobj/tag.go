package vobj

import "github.com/histopathai/main-service/internal/shared/errors"

const (
	NumberTag      TagType = "NUMBER"
	TextTag        TagType = "TEXT"
	BooleanTag     TagType = "BOOLEAN"
	SelectTag      TagType = "SELECT"
	MultiSelectTag TagType = "MULTI_SELECT"
)

type TagType string

func (t TagType) String() string {
	return string(t)
}

func (t TagType) IsValid() bool {
	switch t {
	case NumberTag, TextTag, BooleanTag, SelectTag, MultiSelectTag:
		return true
	default:
		return false
	}
}

type TagValue struct {
	TagType TagType
	TagName string
	Value   any
	Color   *string
	Global  bool
}

func NewTagValue(tagType TagType, tagName string, value any, color *string, global bool) (*TagValue, error) {
	if !tagType.IsValid() {
		details := map[string]any{"tag_type": tagType}
		return nil, errors.NewValidationError("invalid tag type", details)
	}

	if tagName == "" {
		return nil, errors.NewValidationError("tag name is required", nil)
	}

	return &TagValue{
		TagType: tagType,
		TagName: tagName,
		Value:   value,
		Color:   color,
		Global:  global,
	}, nil
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
	return tv.TagType
}

type Tag struct {
	Name     string
	Type     TagType
	Options  []string
	Global   bool
	Required bool
	Min      *float64
	Max      *float64
	Color    *string
}

func NewTag(name string, tagType TagType, options []string, global, required bool, min, max *float64, color *string) (*Tag, error) {
	if name == "" {
		return nil, errors.NewValidationError("tag name is required", nil)
	}

	if !tagType.IsValid() {
		details := map[string]any{"tag_type": tagType}
		return nil, errors.NewValidationError("invalid tag type", details)
	}

	if err := validateTagConstraints(tagType, options, min, max); err != nil {
		return nil, err
	}

	return &Tag{
		Name:     name,
		Type:     tagType,
		Options:  options,
		Global:   global,
		Required: required,
		Min:      min,
		Max:      max,
		Color:    color,
	}, nil
}

func (t *Tag) IsGlobal() bool {
	if t == nil {
		return false
	}
	return t.Global
}

func (t *Tag) IsRequired() bool {
	if t == nil {
		return false
	}
	return t.Required
}

func (t *Tag) IsNumeric() bool {
	if t == nil {
		return false
	}
	return t.Type == NumberTag
}

func (t *Tag) IsText() bool {
	if t == nil {
		return false
	}
	return t.Type == TextTag
}

func (t *Tag) IsBoolean() bool {
	if t == nil {
		return false
	}
	return t.Type == BooleanTag
}

func (t *Tag) IsSelect() bool {
	if t == nil {
		return false
	}
	return t.Type == SelectTag
}

func (t *Tag) IsMultiSelect() bool {
	if t == nil {
		return false
	}
	return t.Type == MultiSelectTag
}

func validateTagConstraints(tagType TagType, options []string, min, max *float64) error {
	if len(options) > 0 && (tagType != SelectTag && tagType != MultiSelectTag) {
		details := map[string]any{"tag_type": tagType, "options_count": len(options)}
		return errors.NewValidationError("options can only be set for select or multi-select tag types", details)
	}

	if (tagType == SelectTag || tagType == MultiSelectTag) && len(options) == 0 {
		details := map[string]any{"tag_type": tagType}
		return errors.NewValidationError("options are required for select and multi-select tag types", details)
	}

	if min != nil && tagType != NumberTag {
		details := map[string]any{"tag_type": tagType, "min": *min}
		return errors.NewValidationError("min can only be set for number tag type", details)
	}

	if max != nil && tagType != NumberTag {
		details := map[string]any{"tag_type": tagType, "max": *max}
		return errors.NewValidationError("max can only be set for number tag type", details)
	}

	if min != nil && max != nil && *min > *max {
		details := map[string]any{"min": *min, "max": *max}
		return errors.NewValidationError("min cannot be greater than max", details)
	}

	return nil
}
