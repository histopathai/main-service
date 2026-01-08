package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateWorkspaceCommand struct {
	Name         string
	CreatorID    string
	ParentID     *string
	OrganType    string
	Organization string
	Description  string
	License      string
	ResourceURL  *string
	ReleaseYear  *int
}

func NewCreateWorkspaceCommand(
	name string,
	creatorID string,
	parentID *string,
	organType string,
	organization string,
	description string,
	license string,
	resourceURL *string,
	releaseYear *int,
) (*CreateWorkspaceCommand, error) {
	details := make(map[string]any)
	if name == "" {
		details["name required"] = name
	}

	if creatorID == "" {
		details["creator_id required"] = creatorID
	}
	if organization == "" {
		details["organization required"] = organization
	}
	if license == "" {
		details["license required"] = license
	}

	if len(details) > 0 {
		return nil, errors.NewValidationError("invalid create workspace command", details)
	}

	ot, err := vobj.NewOrganTypeFromString(organType)
	if err != nil {
		return nil, err
	}

	return &CreateWorkspaceCommand{
		Name:         name,
		CreatorID:    creatorID,
		ParentID:     parentID,
		OrganType:    string(ot),
		Organization: organization,
		Description:  description,
		License:      license,
		ResourceURL:  resourceURL,
		ReleaseYear:  releaseYear,
	}, nil
}

func (c *CreateWorkspaceCommand) ToEntity() (*model.Workspace, error) {
	var parentRef *vobj.ParentRef
	var err error

	if c.ParentID != nil {
		parentRef, err = vobj.NewParentRef(*c.ParentID, vobj.ParentTypeWorkspace)
		if err != nil {
			return nil, err
		}
	}

	entity, err := vobj.NewEntity(
		vobj.EntityTypeWorkspace,
		&c.Name,
		c.CreatorID,
		parentRef,
	)
	if err != nil {
		return nil, err
	}

	return &model.Workspace{
		Entity:       entity,
		OrganType:    c.OrganType,
		Organization: c.Organization,
		Description:  c.Description,
		License:      c.License,
		ResourceURL:  c.ResourceURL,
		ReleaseYear:  c.ReleaseYear,
	}, nil
}

type UpdateWorkspaceCommand struct {
	ID           string
	Name         *string
	CreatorID    *string
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

func (c *UpdateWorkspaceCommand) ApplyTo(entity *model.Workspace) (*model.Workspace, error) {
	if c.Name != nil {
		entity.SetName(*c.Name)
	}
	if c.CreatorID != nil {
		entity.SetCreatorID(*c.CreatorID)
	}
	if c.OrganType != nil {
		ot, err := vobj.NewOrganTypeFromString(*c.OrganType)
		if err != nil {
			return nil, err
		}
		entity.OrganType = string(ot)
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
	return entity, nil
}

func (c *UpdateWorkspaceCommand) GetUpdates() (map[string]any, error) {
	updates := make(map[string]any)

	if c.Name != nil {
		updates[constants.NameField] = *c.Name
	}
	if c.CreatorID != nil {
		updates[constants.CreatorIDField] = *c.CreatorID
	}
	if c.OrganType != nil {
		ot, err := vobj.NewOrganTypeFromString(*c.OrganType)
		if err != nil {
			return nil, err
		}
		updates[constants.OrganTypeField] = string(ot)
	}
	if c.Organization != nil {
		updates[constants.OrganizationField] = *c.Organization
	}
	if c.Description != nil {
		updates[constants.DescField] = *c.Description
	}
	if c.License != nil {
		updates[constants.LicenseField] = *c.License
	}
	if c.ResourceURL != nil {
		updates[constants.ResourceURLField] = *c.ResourceURL
	}
	if c.ReleaseYear != nil {
		updates[constants.ReleaseYearField] = *c.ReleaseYear
	}

	return updates, nil
}
