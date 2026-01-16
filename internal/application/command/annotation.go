package command

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CommandPoint struct {
	X float64
	Y float64
}

type CreateAnnotationCommand struct {
	CreateEntityCommand
	TagType  string
	Value    any
	Color    *string
	IsGlobal bool
	Points   []CommandPoint
}

func (c *CreateAnnotationCommand) Validate() error {
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
	} else if _, err := vobj.NewTagTypeFromString(c.TagType); err != nil {
		details["tag_type"] = "Invalid TagType"
	}

	if c.Value == nil {
		details["value"] = "Value is required"
	}

	if len(c.Points) < 3 && !c.IsGlobal {
		details["points"] = "At least 3 points required for non-global annotation"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}

	return nil
}

func (c *CreateAnnotationCommand) ToEntity() (interface{}, error) {
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

	points := make([]vobj.Point, len(c.Points))
	for i, p := range c.Points {
		points[i] = vobj.Point{X: p.X, Y: p.Y}
	}

	return &model.Annotation{
		Entity:   *entity,
		Polygon:  &points,
		Value:    c.Value,
		TagType:  tagType,
		IsGlobal: c.IsGlobal,
		Color:    c.Color,
	}, nil
}

type UpdateAnnotationCommand struct {
	UpdateEntityCommand
	Value    interface{}
	Color    *string
	IsGlobal *bool
	Points   []CommandPoint
}

func (c *UpdateAnnotationCommand) Validate() error {
	details := make(map[string]interface{})

	if err := c.UpdateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	if c.Points != nil && len(c.Points) < 3 && (c.IsGlobal == nil || !*c.IsGlobal) {
		details["points"] = "At least 3 points required for non-global annotation"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}

	return nil
}

func (c *UpdateAnnotationCommand) GetUpdates() map[string]interface{} {
	if err := c.Validate(); err != nil {
		return nil
	}

	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	if c.Value != nil {
		updates["tag_value"] = c.Value
	}
	if c.Color != nil {
		updates["color"] = *c.Color
	}
	if c.IsGlobal != nil {
		updates["is_global"] = *c.IsGlobal
	}
	if c.Points != nil {
		points := make([]vobj.Point, len(c.Points))
		for i, p := range c.Points {
			points[i] = vobj.Point{X: p.X, Y: p.Y}
		}
		updates["polygon"] = points
	}

	return updates
}
