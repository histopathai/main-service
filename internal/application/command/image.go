package command

import (
	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateImageCommand struct {
	CreateEntityCommand

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

	// Base validation if has errors pull them
	if err := c.CreateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	// Image-specific validations
	if c.ContentType == "" {
		details["content_type"] = "ContentType is required"
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
		status, err := model.NewImageStatusFromString(*c.Status)
		if err != nil {
			details["status"] = "Invalid ImageStatus"
		} else if status == model.StatusProcessed && (c.ProcessedPath == nil || *c.ProcessedPath == "") {
			details["processed_path"] = "ProcessedPath must be set when status is PROCESSED"
		}
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}
	return nil
}

func (c *CreateImageCommand) ToEntity() (interface{}, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	entity, ok := baseEntity.(*vobj.Entity)
	if !ok {
		return nil, errors.NewValidationError("failed to cast to Entity", nil)
	}

	var status model.ImageStatus
	if c.Status != nil {
		status, _ = model.NewImageStatusFromString(*c.Status)
	}
	// Generate new UID for the image UID + Name
	uuid := uuid.New().String()
	entity.SetID(uuid + "-" + entity.Name)
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

type UpdateImageCommand struct {
	UpdateEntityCommand

	Status        *string
	Width         *int
	Height        *int
	Size          *int64
	ProcessedPath *string
}

func (c *UpdateImageCommand) Validate() error {
	details := make(map[string]interface{})

	// Base validation errors
	if err := c.UpdateEntityCommand.Validate(); err != nil {
		if baseErr, ok := err.(*errors.Err); ok {
			for k, v := range baseErr.Details {
				details[k] = v
			}
		}
	}

	// Image-specific validations
	if c.Status != nil {
		status, err := model.NewImageStatusFromString(*c.Status)
		if err != nil {
			details["status"] = "Invalid ImageStatus"
		} else {
			if status == model.StatusDeleting {
				details["status"] = "Status cannot be set to DELETING"
				details["status_reason"] = "DELETING status is managed by Deletion Request process"
			}
			if status == model.StatusProcessed && (c.ProcessedPath == nil || *c.ProcessedPath == "") {
				details["processed_path"] = "ProcessedPath must be set when status is PROCESSED"
			}
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

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}
	return nil
}

func (c *UpdateImageCommand) GetUpdates() map[string]interface{} {
	if err := c.Validate(); err != nil {
		return nil
	}

	// Base updates
	updates := c.UpdateEntityCommand.GetUpdates()
	if updates == nil {
		updates = make(map[string]interface{})
	}

	// Image-specific updates
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
