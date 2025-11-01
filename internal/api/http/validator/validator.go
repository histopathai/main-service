package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type RequestValidator struct {
	validate *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &RequestValidator{validate: v}
}

func (rv *RequestValidator) ValidateStruct(s interface{}) error {
	err := rv.validate.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return errors.NewValidationError("validation failed",
			map[string]interface{}{"error": err.Error()})
	}

	details := make(map[string]interface{})
	for _, e := range validationErrors {
		details[e.Field()] = rv.formatFieldError(e)
	}

	return errors.NewValidationError("validation failed", details)
}

func (rv *RequestValidator) formatFieldError(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "uuid4":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func (rv *RequestValidator) ValidateOrganType(fl string) bool {
	organType := fl

	return constants.IsValidOrganType(organType)
}
