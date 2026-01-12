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

func NewTagTypeFromString(s string) (TagType, error) {
	tagType := TagType(s)
	if !tagType.IsValid() {
		details := map[string]any{"tag_type": s}
		return "", errors.NewValidationError("invalid tag type", details)
	}
	return tagType, nil
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
