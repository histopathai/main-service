// usecase/validator/annotation_type_validator.go
package validator

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/application/usecase/helper"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeValidator struct {
	repo port.AnnotationTypeRepository
	uow  port.UnitOfWorkFactory
}

func NewAnnotationTypeValidator(
	repo port.AnnotationTypeRepository,
	uow port.UnitOfWorkFactory,
) *AnnotationTypeValidator {
	return &AnnotationTypeValidator{
		repo: repo,
		uow:  uow,
	}
}

func (v *AnnotationTypeValidator) ValidateCreate(ctx context.Context, at *model.AnnotationType) error {
	// 1. Validate parent is None
	if at.Parent.Type != vobj.ParentTypeNone || at.Parent.ID != "" {
		return errors.NewValidationError("annotation type cannot have a parent", map[string]interface{}{
			"parent_type": at.Parent.Type,
			"parent_id":   at.Parent.ID,
		})
	}

	// 2. Check name uniqueness
	isUnique, err := helper.CheckNameUniqueInCollection(ctx, v.repo, at.Name)
	if err != nil {
		return errors.NewInternalError("failed to check name uniqueness", err)
	}
	if !isUnique {
		return errors.NewConflictError("annotation type name already exists", map[string]interface{}{
			"name": at.Name,
		})
	}

	// 3. Validate tag type specific rules
	if err := v.validateTagTypeRules(at); err != nil {
		return err
	}

	return nil
}

func (v *AnnotationTypeValidator) ValidateUpdate(
	ctx context.Context,
	annotationTypeID string,
	updates map[string]interface{},
) error {
	// 1. Check immutable fields
	if err := v.validateImmutableFields(updates); err != nil {
		return err
	}

	// 2. Validate name uniqueness if being updated
	if name, ok := updates[fields.EntityName.DomainName()]; ok {
		isUnique, err := helper.CheckNameUniqueInCollection(ctx, v.repo, name.(string), annotationTypeID)
		if err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		}
		if !isUnique {
			return errors.NewConflictError("annotation type name already exists", map[string]interface{}{
				"name": name,
			})
		}
	}

	// 3. Validate options update if present
	if options, ok := updates[fields.AnnotationTypeOptions.DomainName()]; ok {
		currentAT, err := v.repo.Read(ctx, annotationTypeID)
		if err != nil {
			return errors.NewInternalError("failed to read annotation type", err)
		}

		if err := v.validateOptionsUpdate(ctx, currentAT, options); err != nil {
			return err
		}
	}

	return nil
}

func (v *AnnotationTypeValidator) validateImmutableFields(updates map[string]interface{}) error {
	immutableFields := map[string]string{
		fields.AnnotationTypeTagType.DomainName():    "tag_type",
		fields.AnnotationTypeIsGlobal.DomainName():   "is_global",
		fields.AnnotationTypeIsRequired.DomainName(): "is_required",
	}

	for domainField, apiField := range immutableFields {
		if _, ok := updates[domainField]; ok {
			return errors.NewValidationError("field is immutable and cannot be updated", map[string]interface{}{
				"field":   apiField,
				"message": "this field cannot be changed after creation",
			})
		}
	}

	return nil
}

func (v *AnnotationTypeValidator) validateTagTypeRules(at *model.AnnotationType) error {
	switch at.TagType {
	case vobj.NumberTag:
		return v.validateNumberTag(at)
	case vobj.SelectTag, vobj.MultiSelectTag:
		return v.validateSelectTag(at)
	case vobj.TextTag, vobj.BooleanTag:
		// No specific validation needed
		return nil
	default:
		return errors.NewValidationError("unsupported tag type", map[string]interface{}{
			"tag_type": at.TagType,
		})
	}
}

func (v *AnnotationTypeValidator) validateNumberTag(at *model.AnnotationType) error {
	if at.Min != nil && at.Max != nil {
		if *at.Min >= *at.Max {
			return errors.NewValidationError("min must be less than max", map[string]interface{}{
				"min": *at.Min,
				"max": *at.Max,
			})
		}
	}

	if len(at.Options) > 0 {
		return errors.NewValidationError("number tag should not have options", map[string]interface{}{
			"tag_type": at.TagType,
			"options":  at.Options,
		})
	}

	return nil
}

func (v *AnnotationTypeValidator) validateSelectTag(at *model.AnnotationType) error {
	if len(at.Options) == 0 {
		return errors.NewValidationError("select/multi-select tags must have options", map[string]interface{}{
			"tag_type": at.TagType,
		})
	}

	optionsMap := make(map[string]bool)
	for _, option := range at.Options {
		if optionsMap[option] {
			return errors.NewValidationError("duplicate option found", map[string]interface{}{
				"option": option,
			})
		}
		optionsMap[option] = true

		if option == "" {
			return errors.NewValidationError("options cannot be empty strings", nil)
		}
	}

	if at.Min != nil || at.Max != nil {
		return errors.NewValidationError("select/multi-select tags should not have min/max values", map[string]interface{}{
			"tag_type": at.TagType,
		})
	}

	return nil
}

func (v *AnnotationTypeValidator) validateOptionsUpdate(
	ctx context.Context,
	currentAT *model.AnnotationType,
	newOptionsInterface interface{},
) error {
	// 1. Options can only be updated for SELECT and MULTI_SELECT types
	if currentAT.TagType != vobj.SelectTag && currentAT.TagType != vobj.MultiSelectTag {
		return errors.NewValidationError("options can only be updated for SELECT or MULTI_SELECT types", map[string]interface{}{
			"annotation_type_id": currentAT.ID,
			"current_tag_type":   currentAT.TagType,
		})
	}

	newOptions, ok := newOptionsInterface.([]string)
	if !ok {
		return errors.NewValidationError("invalid options format", map[string]interface{}{
			"annotation_type_id": currentAT.ID,
		})
	}

	// 3. Validate new options rules (no duplicates, no empty strings)
	if err := v.validateNewOptions(newOptions); err != nil {
		return err
	}

	// 4. Check if annotation type is in use
	inUse, err := v.isAnnotationTypeInUse(ctx, currentAT.Name)
	if err != nil {
		return errors.NewInternalError("failed to check if annotation type is in use", err)
	}

	if !inUse {
		// No annotations using this type → safe to update
		return nil
	}

	// 5. Annotations exist, check if removed options are in use
	removedOptions := v.getRemovedOptions(currentAT.Options, newOptions)
	if len(removedOptions) == 0 {
		// Only additions → safe to update
		return nil
	}

	// 6. Validate removed options are not in use
	isRemovedOptionInUse, usedOption, err := v.checkRemovedOptionsInUse(ctx, currentAT, removedOptions)
	if err != nil {
		return errors.NewInternalError("failed to check if options are in use", err)
	}

	if isRemovedOptionInUse {
		return errors.NewConflictError("cannot remove option that is in use by annotations", map[string]interface{}{
			"annotation_type_id": currentAT.ID,
			"used_option":        usedOption,
		})
	}

	return nil
}

// validateNewOptions validates the new options list
func (v *AnnotationTypeValidator) validateNewOptions(options []string) error {
	if len(options) == 0 {
		return errors.NewValidationError("options cannot be empty for select/multi-select tags", nil)
	}

	optionsMap := make(map[string]bool)
	for _, option := range options {
		if option == "" {
			return errors.NewValidationError("options cannot contain empty strings", nil)
		}

		if optionsMap[option] {
			return errors.NewValidationError("duplicate option found", map[string]interface{}{
				"option": option,
			})
		}
		optionsMap[option] = true
	}

	return nil
}

// getRemovedOptions returns options that exist in current but not in new
func (v *AnnotationTypeValidator) getRemovedOptions(current, new []string) []string {
	newMap := make(map[string]bool)
	for _, opt := range new {
		newMap[opt] = true
	}

	var removed []string
	for _, currentOpt := range current {
		if !newMap[currentOpt] {
			removed = append(removed, currentOpt)
		}
	}

	return removed
}

// isAnnotationTypeInUse checks if any annotation uses this annotation type
func (v *AnnotationTypeValidator) isAnnotationTypeInUse(ctx context.Context, annotationTypeName string) (bool, error) {
	annotationRepo := v.uow.GetAnnotationRepo()

	builder := query.NewBuilder()
	builder.Where(fields.EntityName.DomainName(), query.OpEqual, annotationTypeName)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	count, err := annotationRepo.Count(ctx, builder.Build())
	if err != nil {
		return false, fmt.Errorf("failed to count annotations: %w", err)
	}

	return count > 0, nil
}

// checkRemovedOptionsInUse checks if any removed options are in use
func (v *AnnotationTypeValidator) checkRemovedOptionsInUse(
	ctx context.Context,
	annotationType *model.AnnotationType,
	removedOptions []string,
) (bool, string, error) {
	annotationRepo := v.uow.GetAnnotationRepo()

	// Create map for O(1) lookup
	removedOptionsMap := make(map[string]bool)
	for _, opt := range removedOptions {
		removedOptionsMap[opt] = true
	}

	// Fetch annotations using this annotation type
	builder := query.NewBuilder()
	builder.Where(fields.EntityName.DomainName(), query.OpEqual, annotationType.Name)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	const limit = 1000
	offset := 0

	for {
		builder.Paginate(limit, offset)
		result, err := annotationRepo.Find(ctx, builder.Build())
		if err != nil {
			return false, "", fmt.Errorf("failed to fetch annotations: %w", err)
		}

		if len(result.Data) == 0 {
			break
		}

		// Check each annotation's value
		for _, annotation := range result.Data {
			if usedOption := v.checkAnnotationValue(annotation.Value, removedOptionsMap); usedOption != "" {
				return true, usedOption, nil
			}
		}

		if !result.HasMore {
			break
		}

		offset += limit
	}

	return false, "", nil
}

// checkAnnotationValue checks if annotation value uses any removed option
func (v *AnnotationTypeValidator) checkAnnotationValue(value interface{}, removedOptionsMap map[string]bool) string {
	// Handle SELECT type (string value)
	if valueStr, ok := value.(string); ok {
		if removedOptionsMap[valueStr] {
			return valueStr
		}
	}

	// Handle MULTI_SELECT type ([]string value)
	if valueSlice, ok := value.([]string); ok {
		for _, v := range valueSlice {
			if removedOptionsMap[v] {
				return v
			}
		}
	}

	// Handle MULTI_SELECT from JSON unmarshalling ([]interface{})
	if valueInterface, ok := value.([]interface{}); ok {
		for _, item := range valueInterface {
			if str, ok := item.(string); ok {
				if removedOptionsMap[str] {
					return str
				}
			}
		}
	}

	return ""
}
