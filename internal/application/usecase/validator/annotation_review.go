package validator

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type AnnotationReviewValidator struct {
	repo         port.AnnotationReviewRepository
	uow          port.UnitOfWorkFactory
	annValidator *AnnotationValidator
}

func NewAnnotationReviewValidator(repo port.AnnotationReviewRepository, uow port.UnitOfWorkFactory) *AnnotationReviewValidator {
	return &AnnotationReviewValidator{
		repo:         repo,
		uow:          uow,
		annValidator: NewAnnotationValidator(uow.GetAnnotationRepo(), uow),
	}
}

func (v *AnnotationReviewValidator) ValidateCreate(ctx context.Context, review *model.AnnotationReview) error {
	// Check if annotation exists
	// Annotation is the parent — use Parent.ID
	annotation, err := v.uow.GetAnnotationRepo().Read(ctx, review.Parent.ID)
	if err != nil {
		if appErr, ok := err.(*errors.Err); ok && appErr.Type == errors.ErrorTypeNotFound {
			return errors.NewNotFoundError("annotation not found")
		}
		return errors.NewInternalError("failed to get annotation", err)
	}
	if annotation == nil {
		return errors.NewNotFoundError("annotation not found")
	}

	// Rule: cannot review own annotation if it is a manual annotation
	if annotation.CreatorID == review.ReviewerID && annotation.Resource == fields.AnnotationResourceManual {
		return errors.NewValidationError("cannot review your own manual annotation", nil)
	}

	// If modifying the value, validate the new value against the annotation type rules
	if review.Status == fields.ReviewStatusModified && review.ModifiedValue != nil {
		tempAnnotation := *annotation
		tempAnnotation.Value = review.ModifiedValue
		if err := v.annValidator.CheckAnnotationIsValid(ctx, &tempAnnotation); err != nil {
			return err
		}
	}

	return nil
}

func (v *AnnotationReviewValidator) ValidateUpdate(ctx context.Context, id string, requesterID string, updates map[string]interface{}) error {
	// Fetch existing review
	existing, err := v.repo.Read(ctx, id)
	if err != nil {
		return errors.NewInternalError("failed to get annotation review", err)
	}
	if existing == nil {
		return errors.NewNotFoundError("annotation review not found")
	}

	// Only reviewer can update their review
	if existing.ReviewerID != requesterID {
		return errors.NewForbiddenError("you are not the reviewer of this annotation review; you cannot update it")
	}

	// If modifying the value, validate the new value against the annotation type rules
	if statusVal, ok := updates[fields.AnnotationReviewStatus.DomainName()]; ok {
		status, ok := statusVal.(fields.ReviewStatusField)
		if !ok {
			return errors.NewValidationError("invalid type for status", map[string]interface{}{
				"field": "status",
				"error": "must be a ReviewStatusField",
			})
		}
		if status == fields.ReviewStatusModified {
			modifiedValue, hasModifiedValue := updates[fields.AnnotationReviewModifiedValue.DomainName()]
			if hasModifiedValue {
				// Fetch the parent annotation to validate against its type
				annotation, err := v.uow.GetAnnotationRepo().Read(ctx, existing.Parent.ID)
				if err != nil {
					return errors.NewInternalError("failed to get parent annotation", err)
				}
				if annotation == nil {
					return errors.NewNotFoundError("parent annotation not found")
				}

				tempAnnotation := *annotation
				tempAnnotation.Value = modifiedValue
				if err := v.annValidator.CheckAnnotationIsValid(ctx, &tempAnnotation); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
