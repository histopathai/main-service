package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type AnnotationTypeUseCase struct {
	repo      port.AnnotationTypeRepository
	uow       port.UnitOfWorkFactory
	validator *validator.AnnotationTypeValidator
}

func NewAnnotationTypeUseCase(repo port.AnnotationTypeRepository, uow port.UnitOfWorkFactory) *AnnotationTypeUseCase {
	return &AnnotationTypeUseCase{
		repo:      repo,
		uow:       uow,
		validator: validator.NewAnnotationTypeValidator(repo, uow),
	}
}

func (uc *AnnotationTypeUseCase) Create(ctx context.Context, cmd command.CreateAnnotationTypeCommand) (*model.AnnotationType, error) {

	annotationType, err := cmd.ToEntity()
	if err != nil {
		return nil, err
	}

	var createdAnnotationType *model.AnnotationType

	uowerr := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		err := uc.validator.ValidateCreate(txCtx, annotationType)
		if err != nil {
			return err
		}

		createdAnnotationType, err = uc.repo.Create(txCtx, annotationType)
		if err != nil {
			return errors.NewInternalError("failed to create annotation type", err)
		}

		return nil
	})

	if uowerr != nil {
		return nil, uowerr
	}

	return createdAnnotationType, nil
}

func (uc *AnnotationTypeUseCase) Update(ctx context.Context, cmd command.UpdateAnnotationTypeCommand) error {

	updates := cmd.GetUpdates()

	if len(updates) == 0 {
		return errors.NewValidationError("no updates provided", nil)
	}

	if cmd.ID == "" {
		return errors.NewValidationError("annotation type id is required", nil)
	}

	uowerr := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		err := uc.validator.ValidateUpdate(txCtx, cmd.ID, updates)
		if err != nil {
			return err
		}

		err = uc.repo.Update(txCtx, cmd.ID, updates)
		if err != nil {
			return errors.NewInternalError("failed to update annotation type", err)
		}

		return nil
	})

	if uowerr != nil {
		return uowerr
	}

	return nil
}
