package command

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateWorkspaceCommand struct {
	CreateEntityCommand

	OrganType       string
	Organization    string
	Description     string
	License         string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

func (c *CreateWorkspaceCommand) Validate() error {
	details := make(map[string]interface{})

	// Base validation if have errors pull them
	if err := c.CreateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	// Workspace-specific validations
	if c.OrganType == "" {
		details["organ_type"] = "OrganType is required"
	} else {
		_, err := vobj.NewOrganTypeFromString(c.OrganType)
		if err != nil {
			details["organ_type"] = "Invalid OrganType value"
		}
	}

	if c.Organization == "" {
		details["organization"] = "Organization is required"
	}

	if c.Description == "" {
		details["description"] = "Description is required"
	}

	if c.License == "" {
		details["license"] = "License is required"
	}

	if c.ResourceURL != nil && *c.ResourceURL == "" {
		details["resource_url"] = "ResourceURL cannot be empty if provided"
	}

	if c.ReleaseYear != nil && *c.ReleaseYear < 0 {
		details["release_year"] = "ReleaseYear cannot be negative"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation error", details)
	}
	return nil
}

func (c *CreateWorkspaceCommand) ToEntity() (interface{}, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	// Workspace için ParentType always NONE
	c.CreateEntityCommand.ParentID = ""
	c.CreateEntityCommand.ParentType = vobj.ParentTypeNone.String()

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	entity, ok := baseEntity.(*vobj.Entity)
	if !ok {
		return nil, errors.NewValidationError("failed to cast to Entity", nil)
	}

	organType, _ := vobj.NewOrganTypeFromString(c.OrganType)

	return &model.Workspace{
		Entity:          *entity,
		OrganType:       organType,
		Organization:    c.Organization,
		Description:     c.Description,
		License:         c.License,
		ResourceURL:     c.ResourceURL,
		ReleaseYear:     c.ReleaseYear,
		AnnotationTypes: c.AnnotationTypes,
	}, nil
}

type UpdateWorkspaceCommand struct {
	UpdateEntityCommand

	OrganType       *string
	Organization    *string
	Description     *string
	License         *string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

func (c *UpdateWorkspaceCommand) Validate() error {
	details := make(map[string]interface{})

	// Base validation if have errors pull them
	if err := c.UpdateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	// Workspace-specific validation
	if c.OrganType != nil {
		_, err := vobj.NewOrganTypeFromString(*c.OrganType)
		if err != nil {
			details["organ_type"] = "Invalid OrganType value"
		}
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation error", details)
	}

	return nil
}

func (c *UpdateWorkspaceCommand) GetUpdates() map[string]interface{} {
	if err := c.Validate(); err != nil {
		return nil
	}

	// Base updates
	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	// Workspace-specific updates
	if c.OrganType != nil {
		updates["organ_type"] = *c.OrganType
	}
	if c.Organization != nil {
		updates["organization"] = *c.Organization
	}
	if c.Description != nil {
		updates["description"] = *c.Description
	}
	if c.License != nil {
		updates["license"] = *c.License
	}
	if c.ResourceURL != nil {
		updates["resource_url"] = *c.ResourceURL
	}
	if c.ReleaseYear != nil {
		updates["release_year"] = *c.ReleaseYear
	}
	if c.AnnotationTypes != nil {
		updates["annotation_types"] = c.AnnotationTypes
	}

	return updates
}
