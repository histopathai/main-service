package command

import (
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
	WsID        string
	Format      string
	ContentType string
	// Optional basic fields
	Width  *int
	Height *int
	Size   int64
	// Optional WSI fields
	Magnification *struct {
		Objective         *float64
		NativeLevel       *int
		ScanMagnification *float64
	}
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
	if c.ContentType == "" {
		details["content_type"] = "ContentType is required"
	}
	if c.Size <= 0 {
		details["size"] = "Size is required and must be positive"
	}

	// Optional fields validation
	if c.Width != nil && *c.Width <= 0 {
		details["width"] = "Width must be positive if provided"
	}

	if c.Height != nil && *c.Height <= 0 {
		details["height"] = "Height must be positive if provided"
	}

	if c.Size <= 0 {
		details["size"] = "Size must be positive if provided"
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

	// Logic validation
	if c.ContentType != "" {
		_, err := vobj.NewContentTypeFromString(c.ContentType)
		if err != nil {
			details["content_type"] = "Invalid ContentType"
		}

	} else {
		details["content_type"] = "ContentType is required"
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

func (c *UploadImageCommand) GetSize() int64 {
	return c.Size
}

func (c *UploadImageCommand) GetContent() *model.Content {
	_, ok := c.Validate()
	if !ok {
		return nil
	}

	contentType, _ := vobj.NewContentTypeFromString(c.ContentType)

	if contentType.GetCategory() != "image" {
		return nil
	}
	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil
	}
	baseEntity.SetParent(&vobj.ParentRef{
		ID:   "",
		Type: vobj.ParentTypeImage,
	})

	content := &model.Content{
		Entity:      *baseEntity,
		ContentType: contentType,
		Size:        c.Size,
	}
	return content
}

// =============================================================================
// Upload Content Command
// =============================================================================

type UploadContentCommand struct {
	CreateEntityCommand
	ContentType string
	Size        int64
	Provider    *string
	Path        *string
}

func (c *UploadContentCommand) Validate() (map[string]interface{}, bool) {
	details, ok := c.CreateEntityCommand.Validate()
	if ok {
		details = make(map[string]interface{})
	}

	if c.ContentType == "" {
		details["content_type"] = "ContentType is required"
	} else {
		_, err := vobj.NewContentTypeFromString(c.ContentType)
		if err != nil {
			details["content_type"] = "Invalid ContentType"
		}
	}

	if c.Size <= 0 {
		details["size"] = "Size must be positive"
	}

	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

func (c *UploadContentCommand) ToEntity() (*model.Content, error) {
	if details, ok := c.Validate(); !ok {
		return nil, errors.NewValidationError("validation error", details)
	}

	baseEntity, err := c.CreateEntityCommand.ToEntity()
	if err != nil {
		return nil, err
	}

	contentType, _ := vobj.NewContentTypeFromString(c.ContentType)

	contentEntity := model.Content{
		Entity:      *baseEntity,
		ContentType: contentType,
		Size:        c.Size,
	}

	if c.Provider != nil {
		provider := vobj.ContentProvider(*c.Provider)
		contentEntity.Provider = provider
	}

	if c.Path != nil {
		contentEntity.Path = *c.Path
	}

	return &contentEntity, nil
}
