package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateTagCommand struct {
	//BaseEntity fields
	Name      string
	CreatorID string
	//Tag specific fields
	Color       *string
	Description *string
	Type        string
	Required    bool
	Global      bool
	Min         *float64
	Max         *float64
	Options     *[]string
}

func (c *CreateTagCommand) ToEntity() (model.Tag, error) {
	if err := constants.ValidateTagType(c.Type); err != nil {
		return model.Tag{}, err
	}

	if err := constants.ValidateTagPair(c.Type, func() []string {
		if c.Options != nil {
			return *c.Options
		}
		return nil
	}(), c.Min, c.Max); err != nil {
		return model.Tag{}, err
	}

	return model.Tag{
		BaseEntity: model.BaseEntity{
			EntityType: constants.EntityTypeTag,
			Name:       &c.Name,
			CreatorID:  c.CreatorID,
		},
		Color:       c.Color,
		Description: c.Description,
		Type:        model.TagType(c.Type),
		Required:    c.Required,
		Global:      c.Global,
		Min:         c.Min,
		Max:         c.Max,
		Options:     []string{},
	}, nil
}

type UpdateTagCommand struct {
	// BaseEntity fields
	ID        string
	Name      *string
	CreatorID *string
	// Tag specific fields
	Color       *string
	Description *string
	Type        *string
	Required    *bool
	Global      *bool
	Min         *float64
	Max         *float64
	Options     *[]string
}

func (c *UpdateTagCommand) GetID() string {
	return c.ID
}

func (c *UpdateTagCommand) ApplyTo(entity model.Tag) (model.Tag, error) {
	details := make(map[string]any)

	if c.Name != nil {
		entity.Name = c.Name
	}
	if c.CreatorID != nil {
		entity.CreatorID = *c.CreatorID
	}
	if c.Color != nil {
		entity.Color = c.Color
	}
	if c.Description != nil {
		entity.Description = c.Description
	}
	if c.Type != nil {
		if err := constants.ValidateTagType(*c.Type); err != nil {
			return model.Tag{}, err
		}
		entity.Type = model.TagType(*c.Type)
	}
	if c.Required != nil {
		entity.Required = *c.Required
	}
	if c.Global != nil {
		entity.Global = *c.Global
	}
	if c.Min != nil {
		entity.Min = c.Min
	}
	if c.Max != nil {
		entity.Max = c.Max
	}
	if c.Options != nil {
		if err := constants.ValidateTagPair(func() string {
			if c.Type != nil {
				return *c.Type
			}
			return string(entity.Type)
		}(), *c.Options, c.Min, c.Max); err != nil {
			return model.Tag{}, err
		}
		entity.Options = *c.Options
	}

	if len(details) > 0 {
		return model.Tag{}, errors.NewValidationError("invalid fields in UpdateTagCommand", details)
	}

	return entity, nil
}
