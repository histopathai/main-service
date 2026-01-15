package entityspecific

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateAnnotationType struct {
	Name       string
	EntityType string
	CreatorID  string
	TagType    string
	IsGlobal   bool
	IsRequired bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *CreateAnnotationType) Validate() error {
	details := make(map[string]interface{})

	if c.Name == "" {
		details["name"] = "Name is required"
	}

	if c.EntityType == "" {
		details["entity_type"] = "EntityType is required"
	}

	if c.CreatorID == "" {
		details["creator_id"] = "CreatorID is required"
	}
	if c.TagType == "" {
		details["tag_type"] = "TagType is required"
	}

	entity_type, err := vobj.NewEntityTypeFromString(c.EntityType)
	if err != nil {
		details["entity_type"] = "Invalid EntityType"
	} else if entity_type != vobj.EntityTypeAnnotationType {
		details["entity_type"] = "EntityType must be ANNOTATION_TYPE"
	}

	tag_type, err := vobj.NewTagTypeFromString(c.TagType)
	if err != nil {
		details["tag_type"] = "Invalid TagType"
	}

	if tag_type == vobj.NumberTag && (c.Min == nil || c.Max == nil) {
		details["min_max"] = "Min Max must be provided for Number TagType"
	} else if tag_type == vobj.NumberTag && c.Min != nil && c.Max != nil && *c.Min >= *c.Max {
		details["min_max"] = "Min must be less than Max"
	} else if tag_type == vobj.BooleanTag && (len(c.Options) != 0 || c.Min != nil || c.Max != nil) {
		details["options_min_max"] = "Options, Min, Max should not be provided for Boolean TagType"
	} else if tag_type == vobj.MultiSelectTag || tag_type == vobj.SelectTag {
		if len(c.Options) < 1 {
			details["options"] = "At least one option must be provided for MultiSelect or SingleSelect TagType"
		}
	} else if tag_type == vobj.TextTag && (len(c.Options) != 0 || c.Min != nil || c.Max != nil) {
		details["options_min_max"] = "Options, Min, Max should not be provided for Text TagType"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation error", details)
	}
	return nil
}

func (c *CreateAnnotationType) ToEntity() (interface{}, error) {
	if ok := c.Validate(); ok != nil {
		return nil, ok
	}

	entity_type, _ := vobj.NewEntityTypeFromString(c.EntityType)
	parent, _ := vobj.NewParentRef("", vobj.ParentTypeNone)
	entity, err := vobj.NewEntity(
		entity_type,
		&c.Name,
		c.CreatorID,
		parent)

	if err != nil {
		return nil, err
	}

	tag_type, _ := vobj.NewTagTypeFromString(c.TagType)

	annotationType := model.AnnotationType{
		Entity:     *entity,
		TagType:    tag_type,
		IsGlobal:   c.IsGlobal,
		IsRequired: c.IsRequired,
		Options:    c.Options,
		Min:        c.Min,
		Max:        c.Max,
		Color:      c.Color,
	}

	return annotationType, nil

}

type UpdateAnnotationType struct {
	ID         string
	CreatorID  *string
	IsGlobal   *bool
	IsRequired *bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *UpdateAnnotationType) Validate() error {
	details := make(map[string]interface{})

	if c.ID == "" {
		details["id"] = "ID is required"
	}
	if c.Min == nil || c.Max == nil {
		details["min_max"] = "Min Max must be provided for Number TagType"
	} else if c.Min != nil && c.Max != nil && *c.Min >= *c.Max {
		details["min_max"] = "Min must be less than Max"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation error", details)
	}
	return nil
}

func (c *UpdateAnnotationType) GetID() string {
	return c.ID
}

func (c *UpdateAnnotationType) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})

	if ok := c.Validate(); ok != nil {
		return nil
	}

	if c.CreatorID != nil {
		updates["creator_id"] = *c.CreatorID
	}
	if c.IsGlobal != nil {
		updates["is_global"] = *c.IsGlobal
	}
	if c.IsRequired != nil {
		updates["is_required"] = *c.IsRequired
	}
	if c.Options != nil {
		updates["options"] = c.Options
	}
	if c.Min != nil {
		updates["min"] = *c.Min
	}
	if c.Max != nil {
		updates["max"] = *c.Max
	}
	if c.Color != nil {
		updates["color"] = *c.Color
	}

	return updates
}
