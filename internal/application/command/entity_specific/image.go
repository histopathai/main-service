package entityspecific

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateImageCommand struct {
	EntityType string
	CreatorID  string

	ID         *string
	Name       string
	ParentID   string
	ParentType string

	ContentType   string
	Format        string
	OriginPath    string
	Size          *int64
	Width         *int
	Height        *int
	Status        *string
	ProcessedPath *string
}

func (c *CreateImageCommand) Validate() error {
	details := make(map[string]interface{})
	if c.Name == "" {
		details["name"] = "Name is required"
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
	if c.Status != nil {
		_, err := model.NewImageStatusFromString(*c.Status)
		if err != nil {
			details["status"] = "Invalid ImageStatus"
		}
		if *c.Status == model.StatusProcessed.String() && (c.ProcessedPath == nil || *c.ProcessedPath == "") {
			details["processed_path"] = "ProcessedPath must be set when status is PROCESSED"
		}

	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}
	return nil
}

func (c *CreateImageCommand) ToEntity() (interface{}, error) {
	err := c.Validate()
	if err != nil {
		return nil, err
	}
	entity_type, _ := vobj.NewEntityTypeFromString(c.EntityType)
	parentType, _ := vobj.NewParentTypeFromString(c.ParentType)
	parent, _ := vobj.NewParentRef(c.ParentID, parentType)
	entity, _ := vobj.NewEntity(
		entity_type,
		&c.Name,
		c.CreatorID,
		parent)

	status, _ := model.NewImageStatusFromString(*c.Status)

	return &model.Image{
		Entity:        *entity,
		ContentType:   c.ContentType,
		Format:        c.Format,
		OriginPath:    c.OriginPath,
		Size:          c.Size,
		Width:         c.Width,
		Height:        c.Height,
		Status:        status,
		ProcessedPath: c.ProcessedPath,
	}, nil

}

func (c *CreateImageCommand) GetID() string {
	if c.ID == nil {
		return ""
	}
	return *c.ID
}

type UpdateImageCommand struct {
	ID            string
	CreatorID     *string
	Status        *string
	Width         *int
	Height        *int
	Size          *int64
	ProcessedPath *string
}

func (c *UpdateImageCommand) Validate() error {
	details := make(map[string]interface{})
	if c.ID == "" {
		details["id"] = "ID is required in Form data"
	}
	if c.Status != nil {
		status, err := model.NewImageStatusFromString(*c.Status)
		if err != nil {
			details["status"] = "Invalid ImageStatus"
		}
		if status == model.StatusDeleting {
			details["status"] = "Status cannot be set to DELETING"
			details["status_reason"] = "DELETING status is managed by Deletion Request process"
		}
		if status == model.StatusProcessed && (c.ProcessedPath == nil || *c.ProcessedPath == "") {
			details["processed_path"] = "ProcessedPath must be set when status is PROCESSED"
		}
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

	return nil
}

func (c *UpdateImageCommand) GetID() string {
	return c.ID

}

func (c *UpdateImageCommand) GetUpdates() map[string]interface{} {
	updates := make(map[string]interface{})
	if c.CreatorID != nil {
		updates["creator_id"] = *c.CreatorID
	}
	if c.Status != nil {
		updates["status"] = *c.Status
	}
	if c.Width != nil {
		updates["width"] = *c.Width
	}
	if c.Height != nil {
		updates["height"] = *c.Height
	}
	if c.Size != nil {
		updates["size"] = *c.Size
	}
	if c.ProcessedPath != nil {
		updates["processed_path"] = *c.ProcessedPath
	}

	return updates
}
