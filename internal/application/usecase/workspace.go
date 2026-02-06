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

type WorkspaceUseCase struct {
	repo      port.WorkspaceRepository
	uow       port.UnitOfWorkFactory
	validator *validator.WorkspaceValidator
}

func NewWorkspaceUseCase(repo port.WorkspaceRepository, uow port.UnitOfWorkFactory) *WorkspaceUseCase {
	return &WorkspaceUseCase{
		repo:      repo,
		uow:       uow,
		validator: validator.NewWorkspaceValidator(repo, uow),
	}
}

func (uc *WorkspaceUseCase) Create(ctx context.Context, cmd command.CreateWorkspaceCommand) (*model.Workspace, error) {

	entity, err := cmd.ToEntity()
	if err != nil {
		return nil, errors.NewInternalError("failed to convert command to entity", err)
	}

	if err := uc.validator.ValidateCreate(ctx, entity); err != nil {
		return nil, err
	}

	createdWorkspace, err := uc.repo.Create(ctx, entity)
	if err != nil {
		return nil, errors.NewInternalError("failed to create workspace", err)
	}

	return createdWorkspace, nil
}

func (uc *WorkspaceUseCase) Update(ctx context.Context, cmd command.UpdateWorkspaceCommand) error {
	updates := cmd.GetUpdates()
	if updates == nil {
		return errors.NewNotFoundError("no updates provided")
	}
	id := cmd.GetID()

	if id == "" {
		return errors.NewNotFoundError("workspace id not provided")
	}

	if err := uc.validator.ValidateUpdate(ctx, id, updates); err != nil {
		return err
	}

	if updates[fields.WorkspaceAnnotationTypes.DomainName()] != nil {
		newAnnotationTypeIDs := updates[fields.WorkspaceAnnotationTypes.DomainName()].([]string)
		currentEntity, err := uc.repo.Read(ctx, id)
		if err != nil {
			return errors.NewInternalError("failed to read workspace", err)
		}
		if currentEntity == nil {
			return errors.NewNotFoundError("workspace not found")
		}

		if currentEntity.AnnotationTypes == nil {
			currentEntity.AnnotationTypes = []string{}
		}

		if err := uc.validator.ValidateAnnotationTypeRemoval(ctx, id, currentEntity.AnnotationTypes, newAnnotationTypeIDs); err != nil {
			return err
		}

	}

	if err := uc.repo.Update(ctx, id, updates); err != nil {
		return errors.NewInternalError("failed to update workspace", err)
	}

	return nil
}
