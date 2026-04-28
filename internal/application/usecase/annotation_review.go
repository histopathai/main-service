package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type AnnotationReviewUseCase struct {
	repo      port.AnnotationReviewRepository
	uow       port.UnitOfWorkFactory
	validator *validator.AnnotationReviewValidator
}

func NewAnnotationReviewUseCase(repo port.AnnotationReviewRepository, uow port.UnitOfWorkFactory) *AnnotationReviewUseCase {
	return &AnnotationReviewUseCase{
		repo:      repo,
		uow:       uow,
		validator: validator.NewAnnotationReviewValidator(repo, uow),
	}
}

func (uc *AnnotationReviewUseCase) Create(ctx context.Context, cmd command.CreateAnnotationReviewCommand) (*model.AnnotationReview, error) {
	entity, err := cmd.ToEntity()
	if err != nil {
		return nil, err
	}

	var createdReview *model.AnnotationReview
	err = uc.uow.WithTx(ctx, func(txCtx context.Context) error {

		// Validate using validator (reads annotation + annotation type)
		if err := uc.validator.ValidateCreate(txCtx, entity); err != nil {
			return err
		}

		// Read annotation BEFORE any write — Firestore requires all reads precede writes in a transaction
		annotationID := entity.Parent.ID
		existingAnnotation, err := uc.uow.GetAnnotationRepo().Read(txCtx, annotationID)
		if err != nil {
			return errors.NewInternalError("failed to read annotation for review ID update", err)
		}

		// Create review (first write)
		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create annotation review", err)
		}

		// Append the new review ID to the existing list (second write)
		updatedReviewIDs := append(existingAnnotation.ReviewIDs, created.ID)
		updates := map[string]interface{}{
			fields.AnnotationReviewIDs.DomainName(): updatedReviewIDs,
		}
		if err := uc.uow.GetAnnotationRepo().Update(txCtx, annotationID, updates); err != nil {
			return errors.NewInternalError("failed to update annotation with review IDs", err)
		}

		createdReview = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdReview, nil
}

// Delete soft-deletes a review. Only the reviewer who created it can delete it.
func (uc *AnnotationReviewUseCase) Delete(ctx context.Context, reviewID string, requesterID string) error {
	return uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		// 1. Fetch the review
		review, err := uc.repo.Read(txCtx, reviewID)
		if err != nil {
			return errors.NewInternalError("failed to read review", err)
		}
		if review == nil {
			return errors.NewNotFoundError("annotation review not found")
		}

		// 2. Ownership check — only the reviewer can delete their own review
		if review.ReviewerID != requesterID {
			return errors.NewForbiddenError("you can only delete your own reviews")
		}

		// 3. Soft-delete the review
		if err := uc.repo.SoftDelete(txCtx, reviewID); err != nil {
			return errors.NewInternalError("failed to delete annotation review", err)
		}

		// 4. Remove this review ID from the annotation's ReviewIDs list
		annotationID := review.Parent.ID
		annotation, err := uc.uow.GetAnnotationRepo().Read(txCtx, annotationID)
		if err != nil {
			return errors.NewInternalError("failed to read annotation for review ID cleanup", err)
		}
		if annotation != nil {
			updatedIDs := make([]string, 0, len(annotation.ReviewIDs))
			for _, id := range annotation.ReviewIDs {
				if id != reviewID {
					updatedIDs = append(updatedIDs, id)
				}
			}
			updates := map[string]interface{}{
				fields.AnnotationReviewIDs.DomainName(): updatedIDs,
			}
			if err := uc.uow.GetAnnotationRepo().Update(txCtx, annotationID, updates); err != nil {
				return errors.NewInternalError("failed to update annotation review IDs after deletion", err)
			}
		}

		return nil
	})
}
