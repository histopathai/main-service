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
	if s == "" {
		details := map[string]any{"value": s}
		return "", errors.NewValidationError("tag type cannot be empty", details)
	}
	tagType := TagType(s)
	if !tagType.IsValid() {
		details := map[string]any{"tag_type": s}
		return "", errors.NewValidationError("invalid tag type", details)
	}
	return tagType, nil
}
