package command

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreatePatientCommand struct {
	CreateEntityCommand

	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *int
	History *string
}

func (c *CreatePatientCommand) Validate() error {
	details := make(map[string]interface{})

	// Base validation if have errors pull them
	if err := c.CreateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	// Patient-specific validations
	if c.ParentType != vobj.ParentTypeWorkspace.String() {
		details["parent_type"] = "ParentType must be WORKSPACE"
	}

	if c.Age == nil {
		details["age"] = "Age is required"
	} else if *c.Age < 0 || *c.Age > 120 {
		details["age"] = "Age must be between 0 and 120"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}

	return nil
}

func (c *CreatePatientCommand) ToEntity() (interface{}, error) {
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

	return &model.Patient{
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
	UpdateEntityCommand

	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *string
	History *string
}

func (c *UpdatePatientCommand) Validate() error {
	details := make(map[string]interface{})

	// Base validation if have errors pull them
	if err := c.UpdateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	// Patient-specific validations
	if c.Age != nil && (*c.Age < 0 || *c.Age > 120) {
		details["age"] = "Age must be between 0 and 120"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}
	return nil
}

func (c *UpdatePatientCommand) GetUpdates() map[string]interface{} {
	if err := c.Validate(); err != nil {
		return nil
	}

	// Base updates
	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	// Patient-specific updates
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
