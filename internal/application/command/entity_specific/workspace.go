package entityspecific

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateWorkspaceCommand struct {
	Name            string
	EntityType      string
	CreatorID       string
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

	if c.Name == "" {
		details["name"] = "Name is required"
	}

	if c.CreatorID == "" {
		details["creator_id"] = "CreatorID is required"
	}

	if c.OrganType == "" {
		details["organ_type"] = "OrganType is required"

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

	_, err := vobj.NewOrganTypeFromString(c.OrganType)
	if err != nil {
		details["organ_type"] = "Invalid OrganType value"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation error", details)
	}
	return nil
}

func (c *CreateWorkspaceCommand) ToEntity() (interface{}, error) {
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

	return model.Workspace{
		Entity:          *entity,
		OrganType:       c.OrganType,
		Organization:    c.Organization,
		Description:     c.Description,
		License:         c.License,
		ResourceURL:     c.ResourceURL,
		ReleaseYear:     c.ReleaseYear,
		AnnotationTypes: c.AnnotationTypes,
	}, nil

}

type UpdateWorkspaceCommand struct {
	ID              string
	CreatorID       *string
	Name            *string
	OrganType       *string
	Organization    *string
	Description     *string
	License         *string
	ResourceURL     *string
	ReleaseYear     *int
	AnnotationTypes []string
}

func (c *UpdateWorkspaceCommand) Validate() error {
	detail := make(map[string]interface{})

	if c.ID == "" {
		detail["id"] = "ID is required in Form data"
	}

	if c.OrganType != nil {
		_, err := vobj.NewOrganTypeFromString(*c.OrganType)
		if err != nil {
			detail["organ_type"] = "Invalid OrganType value"
		}
	}

	if len(detail) > 0 {
		return errors.NewValidationError("validation error", detail)
	}

	return nil
}

func (c *UpdateWorkspaceCommand) GetID() string {
	return c.ID
}

func (c *UpdateWorkspaceCommand) GetUpdates() map[string]interface{} {

	if ok := c.Validate(); ok != nil {
		return nil
	}

	updates := make(map[string]interface{})

	if c.CreatorID != nil {
		updates["creator_id"] = *c.CreatorID
	}
	if c.Name != nil {
		updates["name"] = *c.Name
	}
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
