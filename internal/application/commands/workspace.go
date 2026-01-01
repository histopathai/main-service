package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateWorkspaceCommand struct {
	Name         string
	CreatorID    string
	OrganType    string
	Organization string
	Description  string
	License      string
	ResourceURL  *string
	ReleaseYear  *int
}

func (c *CreateWorkspaceCommand) ToEntity() (model.Workspace, error) {
	if constants.ValidateEntityType(constants.EntityTypeWorkspace) == false {
		details := map[string]any{"entity_type": constants.EntityTypeWorkspace}
		return model.Workspace{}, errors.NewValidationError("invalid entity type for workspace", details)
	}

	return model.Workspace{
		BaseEntity: model.BaseEntity{
			Name:       &c.Name,
			CreatorID:  c.CreatorID,
			EntityType: constants.EntityTypeWorkspace,
		},
		OrganType:    c.OrganType,
		Organization: c.Organization,
		Description:  c.Description,
		License:      c.License,
		ResourceURL:  c.ResourceURL,
		ReleaseYear:  c.ReleaseYear,
	}, nil
}

type UpdateWorkspaceCommand struct {
	ID        string
	Name      *string
	CreatorID *string

	OrganType    *string
	Organization *string
	Description  *string
	License      *string
	ResourceURL  *string
	ReleaseYear  *int
}

func (c *UpdateWorkspaceCommand) GetID() string {
	return c.ID
}

func (c *UpdateWorkspaceCommand) ApplyTo(entity *model.Workspace) *model.Workspace {
	if c.Name != nil {
		entity.Name = c.Name
	}
	if c.CreatorID != nil {
		entity.CreatorID = *c.CreatorID
	}
	if c.OrganType != nil {
		entity.OrganType = *c.OrganType
	}
	if c.Organization != nil {
		entity.Organization = *c.Organization
	}
	if c.Description != nil {
		entity.Description = *c.Description
	}
	if c.License != nil {
		entity.License = *c.License
	}
	if c.ResourceURL != nil {
		entity.ResourceURL = c.ResourceURL
	}
	if c.ReleaseYear != nil {
		entity.ReleaseYear = c.ReleaseYear
	}
	return entity
}
