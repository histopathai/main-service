package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateImageCommand struct {
	// Base Entity fields
	Name      string
	CreatorID string
	ParentID  string
	// Image specific fields
	Format        string
	OriginPath    string
	ProcessedPath *string
	Width         *int
	Height        *int
	Size          *int64
}

func (c *CreateImageCommand) ToEntity() (model.Image, error) {
	if err := validateNumericImageFields(c.Width, c.Height, c.Size); err != nil {
		return model.Image{}, err
	}

	return model.Image{
		BaseEntity: model.BaseEntity{
			EntityType: constants.EntityTypeImage,
			Name:       &c.Name,
			CreatorID:  c.CreatorID,
			Parent:     &model.ParentRef{ID: c.ParentID, Type: constants.ParentTypePatient},
		},
		Format:        c.Format,
		OriginPath:    c.OriginPath,
		ProcessedPath: c.ProcessedPath,
		Width:         c.Width,
		Height:        c.Height,
		Size:          c.Size,
	}, nil
}

type UpdateImageCommand struct {
	// Base Entity fields
	ID        string
	Name      *string
	CreatorID *string
	ParentID  *string
	// Image specific fields
	Format        *string
	OriginPath    *string
	ProcessedPath *string
	Width         *int
	Height        *int
	Size          *int64
	FailureReason *string
	RetryCount    *int
}

func (c *UpdateImageCommand) GetID() string {
	return c.ID
}

func (c *UpdateImageCommand) ApplyTo(entity model.Image) (model.Image, error) {

	if err := validateNumericImageFields(c.Width, c.Height, c.Size); err != nil {
		return model.Image{}, err
	}
	if c.Name != nil {
		entity.Name = c.Name
	}
	if c.CreatorID != nil {
		entity.CreatorID = *c.CreatorID
	}
	if c.ParentID != nil {
		entity.Parent = &model.ParentRef{ID: *c.ParentID, Type: constants.ParentTypePatient}
	}
	if c.Format != nil {
		entity.Format = *c.Format
	}
	if c.OriginPath != nil {
		entity.OriginPath = *c.OriginPath
	}
	if c.ProcessedPath != nil {
		entity.ProcessedPath = c.ProcessedPath
	}
	if c.Width != nil {
		entity.Width = c.Width
	}
	if c.Height != nil {
		entity.Height = c.Height
	}
	if c.Size != nil {
		entity.Size = c.Size
	}
	if c.FailureReason != nil {
		entity.ProcessReport.FailureReason = c.FailureReason
	}

	return entity, nil
}

func validateNumericImageFields(width, height *int, size *int64) error {
	details := make(map[string]interface{})
	if width != nil && *width <= 0 {
		details["width"] = *width
	}
	if height != nil && *height <= 0 {
		details["height"] = *height
	}
	if size != nil && *size < 0 {
		details["size"] = *size
	}
	if len(details) > 0 {
		return errors.NewValidationError("invalid fields in ImageFields", details)
	}
	return nil
}
