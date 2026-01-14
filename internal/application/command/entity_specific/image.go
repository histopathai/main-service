package entityspecific

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateImageCommand struct {
	ID            *string
	Name          string
	Type          string
	ContentType   string
	CreatorID     string
	ParentID      string
	ParentType    string
	Format        string
	OriginPath    string
	Size          *int64
	Width         *int
	Height        *int
	Status        *string
	ProcessedPath *string
}

func (c *CreateImageCommand) Validate() (interface{}, interface{}, error) {
	details := make(map[string]interface{})
	if c.Name == "" {
		details["name"] = "Name is required"
	}
	if c.Type == "" {
		details["entity_type"] = "Type is required"
	}
	entity_type, err := vobj.NewEntityTypeFromString(c.Type)
	if err != nil {
		details["entity_type"] = "Invalid EntityType"
	}

	if c.ContentType == "" {
		details["content_type"] = "ContentType is required"
	}
	if c.CreatorID == "" {
		details["creator_id"] = "CreatorID is required"
	}
	if c.ParentID == "" {
		details["parent_id"] = "ParentID is required"
	}
	if c.ParentType == "" {
		details["parent_type"] = "ParentType is required"
	}
	if c.Format == "" {
		details["format"] = "Format is required"
	}
	if c.OriginPath == "" {
		details["origin_path"] = "OriginPath is required"
	}
	if c.Size != nil && *c.Size < 0 {
		details["size"] = "Size cannot be negative"
	}
	if c.Width != nil && *c.Width < 0 {
		details["width"] = "Width cannot be negative"
	}
	if c.Height != nil && *c.Height < 0 {
		details["height"] = "Height cannot be negative"
	}

	entity, err := vobj.NewEntity(
		entity_type,
		&c.Name,
		c.CreatorID,
		nil)

	if err != nil {
		details["entity"] = "Failed to create entity"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}
	return entity, nil, nil
}

func (c *CreateImageCommand) ToEntity() (interface{}, error) {
	entity, _, err := c.Validate()
	if err != nil {
		return nil, err
	}

	return model.Image{
		Entity:        *(entity.(*vobj.Entity)),
		ContentType:   c.ContentType,
		ParentID:      c.ParentID,
		ParentType:    vobj.ParentType(c.ParentType),
		Format:        c.Format,
		OriginPath:    c.OriginPath,
		Size:          c.Size,
		Width:         c.Width,
		Height:        c.Height,
		Status:        c.Status,
		ProcessedPath: c.ProcessedPath,
	}, nil
}

func (c *CreateImageCommand) GetID() string {
	return c.ID
}

type UpdateImageCommand struct {
	CreatorID     *string
	Status        *string
	Width         *int
	Height        *int
	Size          *int64
	ProcessedPath *string
}

func (c *UpdateImageCommand) Validate() error {
	// Implement validation logic here
	return nil
}

func (c *UpdateImageCommand) GetID() string {
	return ""
}

func (c *UpdateImageCommand) GetUpdates() map[string]interface{} {
	// Implement logic to return updates as a map
	return nil
}

func (c *UpdateImageCommand) GetUpdatebleFields() []string {
	// Implement logic to return a list of updatable fields
	return nil
}
