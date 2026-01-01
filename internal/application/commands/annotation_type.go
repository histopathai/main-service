package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
)

type CreateAnnotationTypeCommand struct {
	//BaseEntity fields
	Name      string
	CreatorID string

	TagIDs []string
}

func (c *CreateAnnotationTypeCommand) ToEntity() (model.AnnotationType, error) {
	return model.AnnotationType{
		BaseEntity: model.BaseEntity{
			EntityType: constants.EntityTypeAnnotationType,
			Name:       &c.Name,
			CreatorID:  c.CreatorID,
		},
		TagIDs: c.TagIDs,
	}, nil
}

type UpdateAnnotationTypeCommand struct {
	// BaseEntity fields
	ID        string
	Name      *string
	CreatorID *string

	TagIDs *[]string
}

func (c *UpdateAnnotationTypeCommand) GetID() string {
	return c.ID
}

func (c *UpdateAnnotationTypeCommand) ApplyTo(entity model.AnnotationType) (model.AnnotationType, error) {

	if c.Name != nil {
		entity.Name = c.Name
	}
	if c.CreatorID != nil {
		entity.CreatorID = *c.CreatorID
	}
	if c.TagIDs != nil {
		entity.TagIDs = *c.TagIDs
	}

	return entity, nil
}
