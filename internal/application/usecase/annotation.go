package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type AnnotationUseCase struct {
	repo      port.AnnotationRepository
	uow       port.UnitOfWorkFactory
	validator *validator.AnnotationValidator
}

func NewAnnotationUseCase(repo port.AnnotationRepository, uow port.UnitOfWorkFactory) *AnnotationUseCase {
	return &AnnotationUseCase{
		repo:      repo,
		uow:       uow,
		validator: validator.NewAnnotationValidator(repo, uow),
	}
}

func (uc *AnnotationUseCase) Create(ctx context.Context, cmd command.CreateAnnotationCommand) (*model.Annotation, error) {

	entity, err := cmd.ToEntity()
	if err != nil {
		return nil, err
	}

	var createdAnnotation *model.Annotation
	err = uc.uow.WithTx(ctx, func(txCtx context.Context) error {

		// Validate
		if err := uc.validator.ValidateCreate(txCtx, entity); err != nil {
			return err
		}

		// Create annotation
		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create annotation", err)
		}

		createdAnnotation = created

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdAnnotation, nil
}

func (uc *AnnotationUseCase) Update(ctx context.Context, cmd command.UpdateAnnotationCommand) error {

	updates := cmd.GetUpdates()
	if updates == nil {
		return errors.NewInternalError("no updates provided", nil)
	}

	id := cmd.GetID()

	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {

		// Validate
		if err := uc.validator.ValidateUpdate(txCtx, id, updates); err != nil {
			return err
		}

		// Update annotation
		if err := uc.repo.Update(txCtx, id, updates); err != nil {
			return errors.NewInternalError("failed to update annotation", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
