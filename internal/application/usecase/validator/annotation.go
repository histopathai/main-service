package validator

import (
	"context"

	"github.com/histopathai/main-service/internal/application/usecase/helper"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type AnnotationValidator struct {
	repo port.AnnotationRepository
	uow  port.UnitOfWorkFactory
}

func NewAnnotationValidator(repo port.AnnotationRepository, uow port.UnitOfWorkFactory) *AnnotationValidator {
	return &AnnotationValidator{
		repo: repo,
		uow:  uow,
	}
}

func (v *AnnotationValidator) ValidateCreate(ctx context.Context, annotation *model.Annotation) error {
	if err := helper.CheckParentExists(ctx, &annotation.Parent, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}
	// Check WSID exists
	if err := helper.CheckParentExists(ctx, &vobj.ParentRef{ID: annotation.WsID, Type: vobj.ParentTypeWorkspace}, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}

	if err := v.checkAnnotationIsValid(ctx, annotation); err != nil {
		return errors.NewInternalError("failed to check annotation is valid", err)
	}

	return nil
}

func (v *AnnotationValidator) ValidateUpdate(ctx context.Context, id string, updates map[string]interface{}) error {
	// Fetch existing annotation
	existing, err := v.repo.Read(ctx, id)
	if err != nil {
		return errors.NewInternalError("failed to get annotation", err)
	}
	if existing == nil {
		return errors.NewNotFoundError("annotation not found")
	}

	// Create a copy to apply updates
	updatedAnnotation := *existing

	// Apply relevant updates
	if val, ok := updates[fields.AnnotationTagValue.DomainName()]; ok {
		updatedAnnotation.Value = val
		if err := v.checkAnnotationIsValid(ctx, &updatedAnnotation); err != nil {
			return errors.NewInternalError("failed to check annotation is valid", err)
		}
	}
	return nil
}

// Helper functions
func (v *AnnotationValidator) checkAnnotationIsValid(ctx context.Context, annotation *model.Annotation) error {
	annotation_type, err := v.uow.GetAnnotationTypeRepo().Read(ctx, annotation.AnnotationTypeID)
	if err != nil {
		return errors.NewInternalError("failed to get annotation type", err)
	}
	if annotation_type == nil {
		return errors.NewInternalError("annotation type not found", nil)
	}

	if annotation_type.IsRequired {
		if annotation.Value == nil {
			return errors.NewInternalError("annotation value is required", nil)
		}
	}

	switch annotation_type.TagType {
	case vobj.SelectTag:
		if annotation.Value != nil {
			val, ok := annotation.Value.(string)
			if !ok {
				return errors.NewValidationError("invalid type for select tag", nil)
			}
			if !helper.Contains(annotation_type.Options, val) {
				return errors.NewValidationError("annotation value is not valid", map[string]interface{}{
					"value":   val,
					"options": annotation_type.Options,
				})
			}
		}
	case vobj.MultiSelectTag:
		if annotation.Value != nil {
			vals, ok := annotation.Value.([]string)
			if !ok {
				return errors.NewValidationError("invalid type for multi select tag", nil)
			}
			if annotation_type.IsRequired && len(vals) == 0 {
				return errors.NewValidationError("annotation value is required", nil)
			}
			for _, v := range vals {
				if !helper.Contains(annotation_type.Options, v) {
					return errors.NewValidationError("annotation value is not valid", map[string]interface{}{
						"value":   v,
						"options": annotation_type.Options,
					})
				}
			}
		}
	case vobj.TextTag:
		if annotation_type.IsRequired && (annotation.Value == nil || annotation.Value == "") {
			return errors.NewValidationError("annotation value is required", nil)
		}
		if annotation.Value != nil {
			if _, ok := annotation.Value.(string); !ok {
				return errors.NewValidationError("invalid type for text tag", nil)
			}
		}
	case vobj.NumberTag:
		if annotation.Value != nil {
			val, ok := annotation.Value.(float64)
			if !ok {
				return errors.NewValidationError("invalid type for number tag", nil)
			}
			if annotation_type.Max != nil && val > *annotation_type.Max {
				return errors.NewValidationError("annotation value is not valid", map[string]interface{}{
					"value": val,
					"max":   *annotation_type.Max,
				})
			}
			if annotation_type.Min != nil && val < *annotation_type.Min {
				return errors.NewValidationError("annotation value is not valid", map[string]interface{}{
					"value": val,
					"min":   *annotation_type.Min,
				})
			}
		}
	case vobj.BooleanTag:
		if annotation.Value != nil {
			if _, ok := annotation.Value.(bool); !ok {
				return errors.NewValidationError("annotation value is not valid", nil)
			}
		}
	}

	return nil
}
