package command

import (
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/validators"
)

// =============================================================================
// Sort Commands  Interfaces
// =============================================================================
type SortCommand struct {
	Field     string
	Direction string
	Validator validators.FieldValidator
}

func (c *SortCommand) Validate() (map[string]interface{}, bool) {
	details := make(map[string]interface{})

	if c.Field == "" {
		details["field"] = "Field is required"
	}

	if c.Direction != "asc" && c.Direction != "desc" {
		details["direction"] = "Direction must be 'asc' or 'desc'"
	}

	if !c.Validator.IsValidField(c.Field) {
		details["field"] = "Invalid field"
	}

	if len(details) > 0 {
		return details, false
	}

	return nil, true
}

func (c *SortCommand) ToSort() (string, string, error) {
	if _, ok := c.Validate(); !ok {
		return "", "", errors.NewValidationError("invalid sort command", nil)
	}
	fieldConstant, _ := c.Validator.GetFieldConstant(c.Field)

	return fieldConstant, c.Direction, nil
}

// =============================================================================
// Specific Sort Commands for Different Entities
// =============================================================================
type SortWorkspaceCommand struct {
	SortCommand
}

func NewSortWorkspaceCommand(field, direction string) *SortWorkspaceCommand {
	return &SortWorkspaceCommand{
		SortCommand: SortCommand{
			Field:     field,
			Direction: direction,
			Validator: validators.NewCompositeFieldValidator(
				&validators.WorkspaceFieldValidator{},
				&validators.EntityFieldValidator{},
			),
		},
	}
}

type SortImageCommand struct {
	SortCommand
}

func NewSortImageCommand(field, direction string) *SortImageCommand {
	return &SortImageCommand{
		SortCommand: SortCommand{
			Field:     field,
			Direction: direction,
			Validator: validators.NewCompositeFieldValidator(
				&validators.ImageFieldValidator{},
				&validators.EntityFieldValidator{},
			),
		},
	}
}

type SortAnnotationCommand struct {
	SortCommand
}

func NewSortAnnotationCommand(field, direction string) *SortAnnotationCommand {
	return &SortAnnotationCommand{
		SortCommand: SortCommand{
			Field:     field,
			Direction: direction,
			Validator: validators.NewCompositeFieldValidator(
				&validators.AnnotationFieldValidator{},
				&validators.EntityFieldValidator{},
			),
		},
	}
}

type SortAnnotationTypeCommand struct {
	SortCommand
}

func NewSortAnnotationTypeCommand(field, direction string) *SortAnnotationTypeCommand {
	return &SortAnnotationTypeCommand{
		SortCommand: SortCommand{
			Field:     field,
			Direction: direction,
			Validator: validators.NewCompositeFieldValidator(
				&validators.AnnotatitonTypeFieldValidator{},
				&validators.EntityFieldValidator{},
			),
		},
	}
}
