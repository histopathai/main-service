package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateAnnotationCommand struct {
	Name      string
	CreatorID string
	ParentID  string
	Polygon   *[]vobj.Point
	TagValue  vobj.TagValue
}

func NewCreateAnnotationCommand(
	name string,
	creatorID string,
	parentID string,
	polygon *[]vobj.Point,
	tagValue vobj.TagValue,
) (*CreateAnnotationCommand, error) {
	details := make(map[string]any)

	if name == "" {
		details["name required"] = name

	}

	if creatorID == "" {
		details["creator_id required"] = creatorID
	}

	if parentID == "" {
		details["parent_id required"] = parentID
	}

	if polygon == nil || len(*polygon) < 3 {
		details["polygon must have at least 3 points"] = polygon
	}

	if len(details) > 0 {
		return nil, errors.NewValidationError("invalid create annotation command", details)
	}

	return &CreateAnnotationCommand{
		Name:      name,
		CreatorID: creatorID,
		ParentID:  parentID,
		Polygon:   polygon,
		TagValue:  tagValue,
	}, nil
}

func (c *CreateAnnotationCommand) ToEntity() (model.Annotation, error) {
	parentRef, err := vobj.NewParentRef(c.ParentID, vobj.ParentTypeImage)
	if err != nil {
		return model.Annotation{}, err
	}

	entity, err := vobj.NewEntity(
		vobj.EntityTypeAnnotation,
		&c.Name,
		c.CreatorID,
		parentRef,
	)
	if err != nil {
		return model.Annotation{}, err
	}

	return model.Annotation{
		Entity:  *entity,
		Polygon: *c.Polygon,
		Tag:     c.TagValue,
	}, nil
}

type UpdateAnnotationCommand struct {
	ID        string
	Name      *string
	CreatorID *string
	Polygon   *[]vobj.Point
	TagValue  *vobj.TagValue
}

func (c *UpdateAnnotationCommand) GetID() string {
	return c.ID
}

func (c *UpdateAnnotationCommand) ApplyTo(entity model.Annotation) (model.Annotation, error) {
	if c.Name != nil {
		entity.SetName(*c.Name)
	}
	if c.CreatorID != nil {
		entity.SetCreatorID(*c.CreatorID)
	}
	if c.Polygon != nil {
		if len(*c.Polygon) < 3 {
			details := map[string]any{"polygon_length": len(*c.Polygon)}
			return model.Annotation{}, errors.NewValidationError("polygon must have at least 3 points", details)
		}
		entity.Polygon = *c.Polygon
	}
	if c.TagValue != nil {
		entity.Tag = *c.TagValue
	}

	return entity, nil
}

func (c *UpdateAnnotationCommand) GetUpdates() (map[string]any, error) {
	updates := make(map[string]any)

	if c.Name != nil {
		updates[constants.NameField] = *c.Name
	}
	if c.CreatorID != nil {
		updates[constants.CreatorIDField] = *c.CreatorID
	}
	if c.Polygon != nil {
		if len(*c.Polygon) < 3 {
			details := map[string]any{"polygon_length": len(*c.Polygon)}
			return nil, errors.NewValidationError("polygon must have at least 3 points", details)
		}
		updates[constants.AnnotationPolygonField] = *c.Polygon
	}
	if c.TagValue != nil {
		updates[constants.AnnotationTagValueField] = *c.TagValue
	}

	return updates, nil
}
