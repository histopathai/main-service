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

		// Validate using validator
		if err := uc.validator.ValidateCreate(txCtx, entity); err != nil {
			return err
		}

		// Create review
		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create annotation review", err)
		}

		// Update linked annotation with the review ID
		updates := map[string]interface{}{
			fields.AnnotationReviewID.DomainName(): created.ID,
		}
		if err := uc.uow.GetAnnotationRepo().Update(txCtx, created.AnnotationID, updates); err != nil {
			return errors.NewInternalError("failed to update annotation with review ID", err)
		}

		createdReview = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdReview, nil
}
