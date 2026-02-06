package query

import (
	"github.com/histopathai/main-service/internal/shared/errors"
)

type FieldSet interface {
	IsValidField(field string) bool
	GetAllFields() []string
}

type Validator struct {
	allowedFields FieldSet
}

func NewValidator(fields FieldSet) *Validator {
	return &Validator{
		allowedFields: fields,
	}
}

func (v *Validator) ValidateSpec(spec Specification) error {
	for _, filter := range spec.Filters {
		if err := v.validateFilter(filter); err != nil {
			return err
		}
	}

	for _, sort := range spec.Sorts {
		if err := v.validateSort(sort); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateFilter(filter Filter) error {
	if !v.allowedFields.IsValidField(filter.Field) {
		return errors.NewValidationError("invalid filter field", map[string]interface{}{
			"field":          filter.Field,
			"allowed_fields": v.allowedFields.GetAllFields(),
		})
	}

	if !filter.Operator.IsValid() {
		return errors.NewValidationError("invalid filter operator", map[string]interface{}{
			"operator": filter.Operator,
		})
	}

	if filter.Value == nil {
		return errors.NewValidationError("filter value cannot be nil", map[string]interface{}{
			"field": filter.Field,
		})
	}

	return nil
}

func (v *Validator) validateSort(sort Sort) error {
	if !v.allowedFields.IsValidField(sort.Field) {
		return errors.NewValidationError("invalid sort field", map[string]interface{}{
			"field":          sort.Field,
			"allowed_fields": v.allowedFields.GetAllFields(),
		})
	}

	if !sort.Direction.IsValid() {
		return errors.NewValidationError("invalid sort direction", map[string]interface{}{
			"direction": sort.Direction,
		})
	}

	return nil
}
