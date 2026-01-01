package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
)

type CreateAnnotationCommand struct {
	// BaseEntity fields
	Name      string
	CreatorID string
	Polygon   *[]vobj.Point
	TagValue  model.TagValue
}

func (c *CreateAnnotationCommand) ToEntity() (model.Annotation, error) {

	return model.Annotation{
		BaseEntity: model.BaseEntity{
			EntityType: constants.EntityTypeAnnotation,
			Name:       &c.Name,
			CreatorID:  c.CreatorID,
		},
		Polygon: c.Polygon,
		Tag:     c.TagValue,
	}, nil
}

type UpdateAnnotationCommand struct {
	// BaseEntity fields
	ID        string
	Name      *string
	CreatorID *string

	Polygon  *[]vobj.Point
	TagValue *model.TagValue
}

func (c *UpdateAnnotationCommand) GetID() string {
	return c.ID
}

func (c *UpdateAnnotationCommand) ApplyTo(entity model.Annotation) (model.Annotation, error) {
	if c.Name != nil {
		entity.Name = c.Name
	}
	if c.CreatorID != nil {
		entity.CreatorID = *c.CreatorID
	}
	if c.Polygon != nil {
		entity.Polygon = c.Polygon
	}
	if c.TagValue != nil {
		entity.Tag = *c.TagValue
	}

	return entity, nil
}
