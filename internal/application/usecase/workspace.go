package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceUseCase struct {
	repo port.Repository[*model.Workspace]
	uow  port.UnitOfWorkFactory
}

func (uc *WorkspaceUseCase) Create(ctx context.Context, entity *model.Workspace) (*model.Workspace, error) {

	name := entity.GetName()

	filter := query.Filter{
		Field:    constants.NameField,
		Operator: query.OpEqual,
		Value:    name,
	}
	pagination := query.Pagination{
		Limit:  1,
		Offset: 0,
	}

	result, err := uc.repo.FindByFilters(ctx, []query.Filter{filter}, &pagination)
	if err != nil {
		return nil, err
	}
	if len(result.Data) > 0 {
		return nil, errors.NewConflictError("workspace with the same name already exists", nil)
	}

	entity, err = uc.repo.Create(ctx, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (uc *WorkspaceUseCase) Update(ctx context.Context, id string, updates map[string]interface{}) error {

	if name, ok := updates[constants.NameField]; ok {
		filter := query.Filter{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    name,
		}
		pagination := query.Pagination{
			Limit:  1,
			Offset: 0,
		}

		result, err := uc.repo.FindByFilters(ctx, []query.Filter{filter}, &pagination)
		if err != nil {
			return err
		}
		if len(result.Data) > 0 && result.Data[0].ID != id {
			return errors.NewConflictError("workspace with the same name already exists", nil)
		}
	}

	err := uc.repo.Update(ctx, id, updates)
	if err != nil {
		return err
	}

	return nil
}

func (uc *WorkspaceUseCase) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	// Use soft delete for now
	updates := map[string]interface{}{
		constants.DeletedField: true,
	}

	err := uc.repo.Update(ctx, workspaceID, updates)
	if err != nil {
		return err
	}
	return nil
}
