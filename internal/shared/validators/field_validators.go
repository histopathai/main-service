package validators

import (
	"github.com/histopathai/main-service/internal/shared/constants"
)

// =============================================================================
// Field Validator Interface and Composite Implementation
// =============================================================================

type FieldValidator interface {
	IsValidField(field string) bool
	GetFieldConstant(field string) (string, bool)
}

type CompositeFieldValidator struct {
	validators []FieldValidator
}

func NewCompositeFieldValidator(validators ...FieldValidator) *CompositeFieldValidator {
	return &CompositeFieldValidator{
		validators: validators,
	}
}

func (v *CompositeFieldValidator) IsValidField(field string) bool {
	for _, validator := range v.validators {
		if validator.IsValidField(field) {
			return true
		}
	}
	return false
}

func (v *CompositeFieldValidator) GetFieldConstant(field string) (string, bool) {
	for _, validator := range v.validators {
		if constant, ok := validator.GetFieldConstant(field); ok {
			return constant, true
		}
	}
	return "", false
}

type EntityFieldValidator struct{}

func (v *EntityFieldValidator) IsValidField(field string) bool {
	validFields := map[string]bool{
		"id":          true,
		"name":        true,
		"entity_type": true,
		"creator_id":  true,
		"parent_id":   true,
		"parent_type": true,
		"created_at":  true,
		"updated_at":  true,
	}
	return validFields[field]
}

func (v *EntityFieldValidator) GetFieldConstant(field string) (string, bool) {
	{
		fieldConstants := map[string]string{
			"id":          constants.IDField,
			"name":        constants.NameField,
			"entity_type": constants.EntityTypeField,
			"creator_id":  constants.CreatorIDField,
			"parent_id":   constants.ParentIDField,
			"parent_type": constants.ParentTypeField,
			"created_at":  constants.CreatedAtField,
			"updated_at":  constants.UpdatedAtField,
		}
		constant, ok := fieldConstants[field]
		return constant, ok
	}

}

type WorkspaceFieldValidator struct{}

func (v *WorkspaceFieldValidator) IsValidField(field string) bool {
	validFields := map[string]bool{
		"organ_type":   true,
		"organization": true,
		"license":      true,
		"resource_url": true,
		"release_year": true,
		"description":  true,
	}
	return validFields[field]
}

func (v *WorkspaceFieldValidator) GetFieldConstant(field string) (string, bool) {
	fieldConstants := map[string]string{
		"organ_type":   constants.WorkspaceOrganTypeField,
		"organization": constants.WorkspaceOrganizationField,
		"license":      constants.WorkspaceLicenseField,
		"resource_url": constants.WorkspaceResourceURLField,
		"release_year": constants.WorkspaceReleaseYearField,
		"description":  constants.WorkspaceDescField,
	}
	constant, ok := fieldConstants[field]
	return constant, ok
}

type PatientFieldValidator struct{}

func (v *PatientFieldValidator) IsValidField(field string) bool {
	validFields := map[string]bool{
		"age":     true,
		"gender":  true,
		"race":    true,
		"disease": true,
		"subtype": true,
		"grade":   true,
		"history": true,
	}
	return validFields[field]
}

func (v *PatientFieldValidator) GetFieldConstant(field string) (string, bool) {
	{
		fieldConstants := map[string]string{
			"age":     constants.PatientAgeField,
			"gender":  constants.PatientGenderField,
			"race":    constants.PatientRaceField,
			"disease": constants.PatientDiseaseField,
			"subtype": constants.PatientSubtypeField,
			"grade":   constants.PatientGradeField,
			"history": constants.PatientHistoryField,
		}
		constant, ok := fieldConstants[field]
		return constant, ok
	}
}

type ImageFieldValidator struct{}

func (v *ImageFieldValidator) IsValidField(field string) bool {
	validFields := map[string]bool{
		"content_type":      true,
		"format":            true,
		"width":             true,
		"height":            true,
		"size":              true,
		"status":            true,
		"failure_reason":    true,
		"retry_count":       true,
		"last_processed_at": true,
		"origin_path":       true,
		"processed_path":    true,
	}
	return validFields[field]
}

func (v *ImageFieldValidator) GetFieldConstant(field string) (string, bool) {
	fieldConstants := map[string]string{
		"content_type":      constants.ImageContentTypeField,
		"format":            constants.ImageFormatField,
		"width":             constants.ImageWidthField,
		"height":            constants.ImageHeightField,
		"size":              constants.ImageSizeField,
		"status":            constants.ImageProcessingStatusField,
		"failure_reason":    constants.ImageProcessingFailureReasonField,
		"retry_count":       constants.ImageProcessingRetryCountField,
		"last_processed_at": constants.ImageLastProcessedAtField,
		"origin_path":       constants.ImageOriginPathField,
		"processed_path":    constants.ImageProcessedPathField,
	}
	constant, ok := fieldConstants[field]
	return constant, ok
}

type AnnotationFieldValidator struct{}

func (v *AnnotationFieldValidator) IsValidField(field string) bool {
	validFields := map[string]bool{
		"tag_type":  true,
		"value":     true,
		"polygon":   true,
		"is_global": true,
		"color":     true,
	}
	return validFields[field]
}

func (v *AnnotationFieldValidator) GetFieldConstant(field string) (string, bool) {

	fieldConstants := map[string]string{
		"tag_type":  constants.TagTypeField,
		"value":     constants.TagValueField,
		"polygon":   constants.PolygonField,
		"is_global": constants.TagGlobalField,
		"color":     constants.TagColorField,
	}
	constant, ok := fieldConstants[field]
	return constant, ok

}

type AnnotatitonTypeFieldValidator struct{}

func (v *AnnotatitonTypeFieldValidator) IsValidField(field string) bool {
	validFields := map[string]bool{
		"tag_type":    true,
		"options":     true,
		"is_global":   true,
		"is_required": true,
		"min":         true,
		"max":         true,
		"color":       true,
	}
	return validFields[field]
}

func (v *AnnotatitonTypeFieldValidator) GetFieldConstant(field string) (string, bool) {

	fieldConstants := map[string]string{
		"tag_type":    constants.TagTypeField,
		"options":     constants.TagOptionsField,
		"is_global":   constants.TagGlobalField,
		"is_required": constants.TagRequiredField,
		"min":         constants.TagMinField,
		"max":         constants.TagMaxField,
		"color":       constants.TagColorsField,
	}
	constant, ok := fieldConstants[field]
	return constant, ok
}
