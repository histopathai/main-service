package command

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateEntityCommand struct {
	Name       string
	EntityType string
	CreatorID  string
	ParentID   string
	ParentType string
}

func (c *CreateEntityCommand) Validate() error {
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
	}

	if c.EntityType != "" {
		_, err := vobj.NewEntityTypeFromString(c.EntityType)
		if err != nil {
			details["entity_type"] = "Invalid EntityType"
		}
	}

	if c.ParentType != "" {
		_, err := vobj.NewParentTypeFromString(c.ParentType)
		if err != nil {
			details["parent_type"] = "Invalid ParentType"
		}
	}
	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}
	return nil
}

func (c *CreateEntityCommand) ToEntity() (interface{}, error) {
	if ok := c.Validate(); ok != nil {
		return nil, ok
	}

	entity_type, _ := vobj.NewEntityTypeFromString(c.EntityType)

	parent_type, _ := vobj.NewParentTypeFromString(c.ParentType)
	parent, _ := vobj.NewParentRef(c.ParentID, parent_type)

	entity, err := vobj.NewEntity(
		entity_type,
		c.Name,
		c.CreatorID,
		parent,
	)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

type UpdateEntityCommand struct {
	ID        string
	CreatorID *string
	Name      *string
}

func (c *UpdateEntityCommand) Validate() error {
	details := make(map[string]interface{})
	if c.ID == "" {
		details["id"] = "ID is required in Form data"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}
	return nil
}

func (c *UpdateEntityCommand) GetID() string {
	return c.ID
}

func (c *UpdateEntityCommand) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})

	if ok := c.Validate(); ok != nil {
		return nil
	}

	if c.CreatorID != nil {
		updates["creator_id"] = *c.CreatorID
	}
	if c.Name != nil {
		updates["name"] = *c.Name
	}

	return updates
}
