package command

import (
	"fmt"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
)

// =============================================================================
// Upload Image Command
// =============================================================================

type UploadImageCommand struct {
	CreateEntityCommand
	// Required fields
	WsID   string
	Format string

	Contents []struct {
		ContentType string
		Name        string
		Size        int64
	}

	// Optional basic fields
	Width  *int
	Height *int
	// Optional WSI fields
	Magnification *struct {
		Objective         *float64
		NativeLevel       *int
		ScanMagnification *float64
	}
}

func validateContents(contents []struct {
	ContentType string
	Name        string
	Size        int64
}) (map[string]interface{}, bool) {
	details := make(map[string]interface{})
	for i, content := range contents {
		if content.ContentType == "" {
			details[fmt.Sprintf("contents[%d].content_type", i)] = "ContentType is required"
		}
		if content.Name == "" {
			details[fmt.Sprintf("contents[%d].name", i)] = "Name is required"
		}

		if content.Size <= 0 {
			details[fmt.Sprintf("contents[%d].size", i)] = "Size is required and must be positive"
		}
	}
	return details, len(details) == 0
}

func (c *UploadImageCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.CreateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	// Required fields
	if c.WsID == "" {
		details["ws_id"] = "WsID is required"
	}
	if c.Format == "" {
		details["format"] = "Format is required"
	}

	contentsDetails, ok := validateContents(c.Contents)
	if !ok {
		details["contents"] = contentsDetails
	}

	// Optional fields validation
	if c.Width != nil && *c.Width <= 0 {
		details["width"] = "Width must be positive if provided"
	}

	if c.Height != nil && *c.Height <= 0 {
		details["height"] = "Height must be positive if provided"
	}

	if c.Magnification != nil {
		if c.Magnification.Objective != nil && *c.Magnification.Objective <= 0 {
			details["magnification.objective"] = "Objective must be positive if provided"
		}
		if c.Magnification.NativeLevel != nil && *c.Magnification.NativeLevel < 0 {
			details["magnification.native_level"] = "NativeLevel cannot be negative if provided"
		}
		if c.Magnification.ScanMagnification != nil && *c.Magnification.ScanMagnification <= 0 {
			details["magnification.scan_magnification"] = "ScanMagnification must be positive if provided"
		}
	}

	if len(details) > 0 {
		return details, false
	}

	return nil, true
}

func (c *UploadImageCommand) ToEntity() (*model.Image, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	// Build image entity
	image := &model.Image{
		Entity: *baseEntity,
		WsID:   c.WsID,
		Format: c.Format,
		Width:  c.Width,
		Height: c.Height,
	}

	// Optional magnification
	if c.Magnification != nil {
		image.Magnification = &vobj.OpticalMagnification{
			Objective:         c.Magnification.Objective,
			NativeLevel:       c.Magnification.NativeLevel,
			ScanMagnification: c.Magnification.ScanMagnification,
		}

	}

	return image, nil
}
