package command

import (
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
	"github.com/histopathai/main-service/internal/shared/validators"
)

// =============================================================================
// Filter Commands  Interfaces
// =============================================================================

type FilterCommand interface {
	Validate() error
	ToFilter() query.Filter
}

type BaseFilterCommand struct {
	Field     string
	Operator  string
	Value     interface{}
	Validator validators.FieldValidator
}

func (c *BaseFilterCommand) Validate() error {
	details := make(map[string]interface{})

	if c.Field == "" {
		details["field"] = "Field is required"
	}
	if c.Operator == "" {
		details["operator"] = "Operator is required"
	}
	if c.Value == nil {
		details["value"] = "Value is required"
	}

	if len(details) > 0 {
		return errors.NewValidationError("validation failed", details)
	}

	ops := query.FilterOp(c.Operator)
	if !ops.IsValid() {
		return errors.NewValidationError("validation failed", map[string]interface{}{
			"operator": "Invalid operator",
		})
	}

	if !c.Validator.IsValidField(c.Field) {
		return errors.NewNotFoundError("invalid field")
	}

	return nil
}

func (c *BaseFilterCommand) ToFilter() (query.Filter, error) {
	if err := c.Validate(); err != nil {
		return query.Filter{}, err
	}

	fieldConstant, ok := c.Validator.GetFieldConstant(c.Field)
	if !ok {
		return query.Filter{}, errors.NewNotFoundError("field mapping not found")
	}

	return query.Filter{
		Field:    fieldConstant,
		Operator: query.FilterOp(c.Operator),
		Value:    c.Value,
	}, nil
}

type FilterWorkspaceCommand struct {
	BaseFilterCommand
}

func NewWorkspaceFilterCommand(field, operator string, value interface{}) *FilterWorkspaceCommand {
	return &FilterWorkspaceCommand{
		BaseFilterCommand: BaseFilterCommand{
			Field:     field,
			Operator:  operator,
			Value:     value,
			Validator: validators.NewCompositeFieldValidator(&validators.WorkspaceFieldValidator{}, &validators.EntityFieldValidator{}),
		},
	}
}

type FilterPatientCommand struct {
	BaseFilterCommand
}

func NewPatientFilterCommand(field, operator string, value interface{}) *FilterPatientCommand {
	return &FilterPatientCommand{
		BaseFilterCommand: BaseFilterCommand{
			Field:     field,
			Operator:  operator,
			Value:     value,
			Validator: validators.NewCompositeFieldValidator(&validators.PatientFieldValidator{}, &validators.EntityFieldValidator{}),
		},
	}
}

type FilterImageCommand struct {
	BaseFilterCommand
}

func NewImageFilterCommand(field, operator string, value interface{}) *FilterImageCommand {
	return &FilterImageCommand{
		BaseFilterCommand: BaseFilterCommand{
			Field:     field,
			Operator:  operator,
			Value:     value,
			Validator: validators.NewCompositeFieldValidator(&validators.ImageFieldValidator{}, &validators.EntityFieldValidator{}),
		},
	}
}

type FilterAnnotationCommand struct {
	BaseFilterCommand
}

func NewAnnotationFilterCommand(field, operator string, value interface{}) *FilterAnnotationCommand {
	return &FilterAnnotationCommand{
		BaseFilterCommand: BaseFilterCommand{
			Field:     field,
			Operator:  operator,
			Value:     value,
			Validator: validators.NewCompositeFieldValidator(&validators.AnnotationFieldValidator{}, &validators.EntityFieldValidator{}),
		},
	}
}

type FilterAnnotationTypeCommand struct {
	BaseFilterCommand
}

func NewAnnotationTypeFilterCommand(field, operator string, value interface{}) *FilterAnnotationTypeCommand {
	return &FilterAnnotationTypeCommand{
		BaseFilterCommand: BaseFilterCommand{
			Field:     field,
			Operator:  operator,
			Value:     value,
			Validator: validators.NewCompositeFieldValidator(&validators.AnnotatitonTypeFieldValidator{}, &validators.EntityFieldValidator{}),
		},
	}
}
