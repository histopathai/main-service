package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateAnnotationTypeCommand struct {
	Name      string
	CreatorID string
	Tags      []vobj.Tag
}

func NewCreateAnnotationTypeCommand(
	name string,
	creatorID string,
	tags []vobj.Tag,
) (*CreateAnnotationTypeCommand, error) {
	details := make(map[string]any)
	if name == "" {
		details["name required"] = name
	}

	if creatorID == "" {
		details["creator_id required"] = creatorID
	}

	if len(tags) == 0 {
		details["tags required"] = tags
	}

	if len(details) > 0 {
		return nil, errors.NewValidationError("invalid create annotation type command", details)
	}

	return &CreateAnnotationTypeCommand{
		Name:      name,
		CreatorID: creatorID,
		Tags:      tags,
	}, nil
}

func (c *CreateAnnotationTypeCommand) ToEntity() (model.AnnotationType, error) {
	entity, err := vobj.NewEntity(
		vobj.EntityTypeAnnotationType,
		&c.Name,
		c.CreatorID,
		nil,
	)
	if err != nil {
		return model.AnnotationType{}, err
	}

	return model.AnnotationType{
		Entity: entity,
		Tags:   c.Tags,
	}, nil
}

type UpdateAnnotationTypeCommand struct {
	ID        string
	Name      *string
	CreatorID *string
	Tags      *[]vobj.Tag
}

func (c *UpdateAnnotationTypeCommand) GetID() string {
	return c.ID
}

func (c *UpdateAnnotationTypeCommand) ApplyTo(entity model.AnnotationType) (model.AnnotationType, error) {
	if c.Name != nil {
		entity.SetName(*c.Name)
	}
	if c.CreatorID != nil {
		entity.SetCreatorID(*c.CreatorID)
	}
	if c.Tags != nil {
		entity.Tags = *c.Tags
	}

	return entity, nil
}

func (c *UpdateAnnotationTypeCommand) GetUpdates() (map[string]any, error) {
	updates := make(map[string]any)

	if c.Name != nil {
		updates[constants.NameField] = *c.Name
	}
	if c.CreatorID != nil {
		updates[constants.CreatorIDField] = *c.CreatorID
	}
	if c.Tags != nil {
		updates[constants.TagsField] = *c.Tags
	}

	return updates, nil
}
