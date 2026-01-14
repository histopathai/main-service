package entityspecific

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreatePatientCommand struct {
	Name       string
	Type       string
	CreatorID  string
	ParentID   string
	ParentType string
	Age        *int
	Gender     *string
	Race       *string
	Disease    *string
	Subtype    *string
	Grade      *int
	History    *string
}

func (c *CreatePatientCommand) Validate() error {
	details := make(map[string]interface{})
	if c.Name == "" {
		details["name"] = "Name is required"
	}
	if c.Type == "" {
		details["entity_type"] = "Type is required"
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
	_, err := vobj.NewParentRef(c.ParentID, vobj.ParentType(c.ParentType))
	if err != nil {
		details["parent"] = "Invalid ParentType or ParentID"
	}

	if c.Age == nil || *c.Age < 0 || *c.Age > 120 {
		details["age"] = "Age must be between 0 and 120"
	}

	_, err = vobj.NewEntityTypeFromString(c.Type)
	if err != nil {
		details["entity_type"] = "Invalid EntityType"
	}
	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}

	return nil
}

func (c *CreatePatientCommand) ToEntity() (interface{}, error) {
	if ok := c.Validate(); ok != nil {
		return nil, ok
	}

	entity_type, _ := vobj.NewEntityTypeFromString(c.Type)

	parent, _ := vobj.NewParentRef(c.ParentID, vobj.ParentType(c.ParentType))

	entity, err := vobj.NewEntity(
		entity_type,
		&c.Name,
		c.CreatorID,
		parent)
	if err != nil {
		return nil, err
	}

	return model.Patient{
		Entity:  *entity,
		Age:     c.Age,
		Race:    c.Race,
		Gender:  c.Gender,
		Disease: c.Disease,
		Subtype: c.Subtype,
		Grade:   c.Grade,
		History: c.History,
	}, nil
}

type UpdatePatientCommand struct {
	ID        string
	Name      *string
	CreatorID *string
	Age       *int
	Gender    *string
	Race      *string
	Disease   *string
	Subtype   *string
	Grade     *string
	History   *string
}

func (c *UpdatePatientCommand) Validate() error {
	detail := make(map[string]interface{})
	if c.ID == "" {
		detail["id"] = "ID is required"
	}

	if c.Age != nil {
		if *c.Age < 0 || *c.Age > 120 {
			detail["age"] = "Age must be between 0 and 120"
		}
	}

	if len(detail) > 0 {
		return errors.NewValidationError("validation failed", detail)
	}
	return nil
}

func (c *UpdatePatientCommand) GetID() string {
	return c.ID
}

func (c *UpdatePatientCommand) GetUpdates() map[string]interface{} {
	if ok := c.Validate(); ok != nil {
		return nil
	}

	updates := make(map[string]interface{})

	if c.Name != nil {
		updates["name"] = *c.Name
	}
	if c.CreatorID != nil {
		updates["creator_id"] = *c.CreatorID
	}
	if c.Age != nil {
		updates["age"] = *c.Age
	}
	if c.Gender != nil {
		updates["gender"] = *c.Gender
	}
	if c.Race != nil {
		updates["race"] = *c.Race
	}
	if c.Disease != nil {
		updates["disease"] = *c.Disease
	}
	if c.Subtype != nil {
		updates["subtype"] = *c.Subtype
	}
	if c.Grade != nil {
		updates["grade"] = *c.Grade
	}
	if c.History != nil {
		updates["history"] = *c.History
	}

	return updates
}
