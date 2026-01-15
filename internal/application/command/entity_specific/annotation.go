package entityspecific

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
	Name       string
	EntityType string
	CreatorID  string
	ParentID   string
	ParentType string
	TagType    string
	Value      any
	Color      *string
	IsGlobal   bool
	Points     []CommandPoint
}

func (c *CreateAnnotationCommand) Validate() error {
	details := make(map[string]interface{})
	if c.Name == "" {
		details["name"] = "Name is required"
	}
	if c.CreatorID == "" {
		details["creator_id"] = "CreatorID is required"
	}
	if c.ParentID == "" {
		details["parent_id"] = "ParentID is required"
	}
	if c.ParentType == "" {
		details["parent_type"] = "ParentType is required"
		if vobj.ParentTypeImage.String() != c.ParentType {
			details["parent_type"] = "ParentType must be IMAGE"
		}
	}
	if c.TagType == "" {
		details["tag_type"] = "TagType is required"

	}
	if c.Value == nil {
		details["value"] = "Value is required"
	}

	if len(c.Points) < 3 && !c.IsGlobal {
		details["points"] = "At least 3 points are required to form a polygon if not global"
	}

	_, err := vobj.NewTagTypeFromString(c.TagType)
	if err != nil {
		details["tag_type"] = "Invalid TagType"
	}

	_, err = vobj.NewParentTypeFromString(c.ParentType)
	if err != nil {
		details["parent_type"] = "Invalid ParentType"

	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}

	return nil

}

func (c *CreateAnnotationCommand) ToEntity() (interface{}, error) {
	if ok := c.Validate(); ok != nil {
		return nil, ok
	}

	entity_type, _ := vobj.NewEntityTypeFromString(c.EntityType)
	parent_type, _ := vobj.NewParentTypeFromString(c.ParentType)
	parent, _ := vobj.NewParentRef(c.ParentID, parent_type)

	entity, _ := vobj.NewEntity(
		entity_type,
		&c.Name,
		c.CreatorID,
		parent)

	tag_type, _ := vobj.NewTagTypeFromString(c.TagType)

	points := make([]vobj.Point, len(c.Points))
	for i, p := range c.Points {
		points[i] = vobj.Point{
			X: p.X,
			Y: p.Y,
		}
	}

	annotation := model.Annotation{
		Entity:   *entity,
		Polygon:  &points,
		Value:    c.Value,
		TagType:  tag_type,
		IsGlobal: c.IsGlobal,
		Color:    c.Color,
	}

	return &annotation, nil
}

type UpdateAnnotationCommand struct {
	ID        string
	CreatorID *string
	Value     *any
	Color     *string
	IsGlobal  *bool
	Points    []CommandPoint
}

func (c *UpdateAnnotationCommand) Validate() error {
	details := make(map[string]interface{})

	if c.ID == "" {
		details["id"] = "ID is required"
	}

	if c.Points != nil && len(c.Points) < 3 && (c.IsGlobal == nil || !*c.IsGlobal) {
		details["points"] = "At least 3 points are required to form a polygon if not global"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}

	return nil
}

func (c *UpdateAnnotationCommand) GetID() string {
	return c.ID
}

func (c *UpdateAnnotationCommand) GetUpdates() map[string]interface{} {

	updates := make(map[string]interface{})

	if ok := c.Validate(); ok != nil {
		return nil
	}

	if c.CreatorID != nil {
		updates["creator_id"] = *c.CreatorID
	}
	if c.Value != nil {
		updates["tag_value"] = *c.Value
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
			points[i] = vobj.Point{
				X: p.X,
				Y: p.Y,
			}
		}
		updates["polygon"] = points
	}

	return updates
}
