package command

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateAnnotationTypeCommand struct {
	CreateEntityCommand
	TagType    string
	IsGlobal   bool
	IsRequired bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *CreateAnnotationTypeCommand) Validate() error {
	details := make(map[string]interface{})

	if err := c.CreateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	if c.TagType == "" {
		details["tag_type"] = "TagType is required"
	}

	entityType, err := vobj.NewEntityTypeFromString(c.EntityType)
	if err != nil {
		details["entity_type"] = "Invalid EntityType"
	} else if entityType != vobj.EntityTypeAnnotationType {
		details["entity_type"] = "EntityType must be ANNOTATION_TYPE"
	}

	tagType, err := vobj.NewTagTypeFromString(c.TagType)
	if err != nil {
		details["tag_type"] = "Invalid TagType"
	} else {
		switch tagType {
		case vobj.NumberTag:
			if c.Min == nil || c.Max == nil {
				details["min_max"] = "Min and Max must be provided for Number TagType"
			} else if *c.Min >= *c.Max {
				details["min_max"] = "Min must be less than Max"
			}
		case vobj.BooleanTag, vobj.TextTag:
			if len(c.Options) != 0 || c.Min != nil || c.Max != nil {
				details["options_min_max"] = "Options, Min, Max should not be provided for " + string(tagType) + " TagType"
			}
		case vobj.MultiSelectTag, vobj.SelectTag:
			if len(c.Options) < 1 {
				details["options"] = "At least one option must be provided for MultiSelect or SingleSelect TagType"
			}
		}
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation error", details)
	}
	return nil
}

func (c *CreateAnnotationTypeCommand) ToEntity() (interface{}, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	entity, ok := baseEntity.(*vobj.Entity)
	if !ok {
		return nil, errors.NewValidationError("failed to cast to Entity", nil)
	}

	tagType, _ := vobj.NewTagTypeFromString(c.TagType)

	return &model.AnnotationType{
		Entity:     *entity,
		TagType:    tagType,
		IsGlobal:   c.IsGlobal,
		IsRequired: c.IsRequired,
		Options:    c.Options,
		Min:        c.Min,
		Max:        c.Max,
		Color:      c.Color,
	}, nil
}

type UpdateAnnotationTypeCommand struct {
	UpdateEntityCommand
	IsGlobal   *bool
	IsRequired *bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}

func (c *UpdateAnnotationTypeCommand) Validate() error {
	details := make(map[string]interface{})

	if err := c.UpdateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	if c.Min != nil && c.Max != nil && *c.Min >= *c.Max {
		details["min_max"] = "Min must be less than Max"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation error", details)
	}
	return nil
}

func (c *UpdateAnnotationTypeCommand) GetUpdates() map[string]interface{} {
	if err := c.Validate(); err != nil {
		return nil
	}

	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
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
