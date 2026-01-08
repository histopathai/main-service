package commands

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateImageCommand struct {
	Name          string
	CreatorID     string
	ParentID      string
	Format        string
	OriginPath    string
	ProcessedPath *string
	Width         *int
	Height        *int
	Size          *int64
}

func NewCreateImageCommand(
	name string,
	creatorID string,
	parentID string,
	format string,
	originPath string,
	processedPath *string,
	width *int,
	height *int,
	size *int64,
) (*CreateImageCommand, error) {
	detail := make(map[string]any)
	if name == "" {
		detail["name required"] = name
	}

	if creatorID == "" {
		detail["creator_id required"] = creatorID
	}

	if parentID == "" {
		detail["parent_id required"] = parentID
	}

	if format == "" {
		detail["format required"] = format
	}

	if originPath == "" {
		detail["origin_path required"] = originPath
	}

	n_details := validateNumericImageFields(width, height, size)
	for k, v := range n_details {
		detail[k] = v
	}

	if len(detail) > 0 {
		return nil, errors.NewValidationError("invalid create image command", detail)
	}

	return &CreateImageCommand{
		Name:          name,
		CreatorID:     creatorID,
		ParentID:      parentID,
		Format:        format,
		OriginPath:    originPath,
		ProcessedPath: processedPath,
		Width:         width,
		Height:        height,
		Size:          size,
	}, nil
}

func (c *CreateImageCommand) ToEntity() (model.Image, error) {
	parentRef, err := vobj.NewParentRef(c.ParentID, vobj.ParentTypePatient)
	if err != nil {
		return model.Image{}, err
	}

	entity, err := vobj.NewEntity(
		vobj.EntityTypeImage,
		&c.Name,
		c.CreatorID,
		parentRef,
	)
	if err != nil {
		return model.Image{}, err
	}

	return model.Image{
		Entity:        entity,
		Format:        c.Format,
		OriginPath:    c.OriginPath,
		ProcessedPath: c.ProcessedPath,
		Width:         c.Width,
		Height:        c.Height,
		Size:          c.Size,
	}, nil
}

type UpdateImageCommand struct {
	ID            string
	Name          *string
	CreatorID     *string
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
	validation_details := validateNumericImageFields(c.Width, c.Height, c.Size)
	if len(validation_details) > 0 {
		return model.Image{}, errors.NewValidationError("invalid update image command", validation_details)
	}

	if c.Name != nil {
		entity.Name = c.Name
	}
	if c.CreatorID != nil {
		entity.CreatorID = *c.CreatorID
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
	if c.RetryCount != nil {
		entity.ProcessReport.RetryCount = *c.RetryCount
	}

	return entity, nil
}

func (c *UpdateImageCommand) GetUpdates() (map[string]any, error) {
	validation_details := validateNumericImageFields(c.Width, c.Height, c.Size)
	if len(validation_details) > 0 {
		return nil, errors.NewValidationError("invalid update image command", validation_details)
	}

	updates := make(map[string]any)

	if c.Name != nil {
		updates[constants.NameField] = *c.Name
	}
	if c.CreatorID != nil {
		updates[constants.CreatorIDField] = *c.CreatorID
	}
	if c.Format != nil {
		updates[constants.FormatField] = *c.Format
	}
	if c.OriginPath != nil {
		updates[constants.OriginPathField] = *c.OriginPath
	}
	if c.ProcessedPath != nil {
		updates[constants.ProcessedPathField] = *c.ProcessedPath
	}
	if c.Width != nil {
		updates[constants.WidthField] = *c.Width
	}
	if c.Height != nil {
		updates[constants.HeightField] = *c.Height
	}
	if c.Size != nil {
		updates[constants.SizeField] = *c.Size
	}
	if c.FailureReason != nil {
		updates[constants.FailureReasonField] = *c.FailureReason
	}
	if c.RetryCount != nil {
		updates[constants.RetryCountField] = *c.RetryCount
	}

	return updates, nil
}

func validateNumericImageFields(width, height *int, size *int64) map[string]any {
	details := make(map[string]any)

	if width != nil && *width <= 0 {
		details["width must be positive"] = *width
	}
	if height != nil && *height <= 0 {
		details["height must be positive"] = *height
	}
	if size != nil && *size < 0 {
		details["size cannot be negative"] = *size
	}

	return details
}
