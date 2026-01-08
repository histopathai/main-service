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
	Tag       vobj.Tag
}

func NewCreateAnnotationTypeCommand(
	name string,
	creatorID string,
	tag vobj.Tag,
) (*CreateAnnotationTypeCommand, error) {
	details := make(map[string]any)
	if name == "" {
		details["name required"] = name
	}

	if creatorID == "" {
		details["creator_id required"] = creatorID
	}

	if tag.Name == "" {
		details["tag name required"] = tag.Name
	}

	if len(details) > 0 {
		return nil, errors.NewValidationError("invalid create annotation type command", details)
	}

	return &CreateAnnotationTypeCommand{
		Name:      name,
		CreatorID: creatorID,
		Tag:       tag,
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
		Entity: *entity,
		Tag:    &c.Tag,
	}, nil
}

type UpdateAnnotationTypeCommand struct {
	ID        string
	Name      *string
	CreatorID *string
	Tag       *vobj.Tag
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
	if c.Tag != nil {
		entity.Tag = c.Tag
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
	if c.Tag != nil {
		updates[constants.TagField] = c.Tag
	}

	return updates, nil
}
