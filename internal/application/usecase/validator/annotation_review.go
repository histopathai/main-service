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
		return errors.NewInternalError("failed to get annotation", err)
	}
	if annotation == nil {
		return errors.NewNotFoundError("annotation not found")
	}

	// Rule: cannot review own annotation if it is a manual annotation
	if annotation.CreatorID == review.ReviewerID && annotation.Resource == fields.AnnotationResourceManual {
		return errors.NewValidationError("cannot review your own manual annotation", nil)
	}

	// If modifying, test the new values using AnnotationValidator
	if review.Status == fields.ReviewStatusModified {
		tempAnnotation := *annotation

		if review.ModifiedValue != nil {
			tempAnnotation.Value = review.ModifiedValue
		}
		if review.ModifiedPolygon != nil {
			tempAnnotation.Polygon = review.ModifiedPolygon
		}
		
		if err := v.annValidator.CheckAnnotationIsValid(ctx, &tempAnnotation); err != nil {
			return err
		}
	}

	return nil
}
