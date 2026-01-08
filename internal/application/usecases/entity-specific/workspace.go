package entityspecific

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

// Interface'leri implement ettiğini garanti et
var _ CreateExecutor[model.Workspace] = (*CreateWorkspaceUseCase)(nil)
var _ UpdateExecutor[model.Workspace] = (*UpdateWorkspaceUseCase)(nil)

type CreateWorkspaceUseCase struct {
	repo port.Repository[model.Workspace]
}

func NewCreateWorkspaceUseCase(repo port.Repository[model.Workspace]) *CreateWorkspaceUseCase {
	return &CreateWorkspaceUseCase{repo: repo}
}

func (uc *CreateWorkspaceUseCase) Execute(ctx context.Context, entity *model.Workspace) (*model.Workspace, error) {
	name := entity.Name

	// validate name uniqueness
	filters := []query.Filter{
		{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    name,
		},
	}
	count, err := uc.repo.Count(ctx, filters)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		details := map[string]any{
			"where": "CreateUseCase.Execute",
			"type":  "uniqueness violation",
			"name":  name,
		}

		return nil, errors.NewConflictError("workspace with the given name already exists", details)
	}

	createdEntity, err := uc.repo.Create(ctx, *entity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

func (uc *CreateWorkspaceUseCase) ExecuteMany(ctx context.Context, entities []model.Workspace) ([]model.Workspace, error) {

	created := make([]model.Workspace, 0, len(entities))
	// Bulk operations do not applied here, Later we can improve it if needed
	for _, entity := range entities {
		createdEntity, err := uc.Execute(ctx, &entity)
		if err != nil {
			return nil, err
		}
		created = append(created, *createdEntity)
	}
	return created, nil
}

type UpdateWorkspaceUseCase struct {
	repo port.Repository[model.Workspace]
}

func NewUpdateWorkspaceUseCase(repo port.Repository[model.Workspace]) *UpdateWorkspaceUseCase {
	return &UpdateWorkspaceUseCase{repo: repo}
}

func (uc *UpdateWorkspaceUseCase) Execute(ctx context.Context, id string, updates map[string]any) (*model.Workspace, error) {

	if nameValue, ok := updates[constants.NameField]; ok {
		name, ok := nameValue.(string)
		// Check name uniqueness
		if ok {
			filters := []query.Filter{
				{
					Field:    constants.NameField,
					Operator: query.OpEqual,
					Value:    name,
				},
			}
			count, err := uc.repo.Count(ctx, filters)
			if err != nil {
				return nil, err
			}
			if count > 0 {
				details := map[string]any{
					"where": "UpdateUseCase.Execute",
					"type":  "uniqueness violation",
					"name":  name,
				}

				return nil, errors.NewConflictError("workspace with the given name already exists", details)
			}

		}
	}

	err := uc.repo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}
	updatedEntity, err := uc.repo.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	return &updatedEntity, nil
}

func (uc *UpdateWorkspaceUseCase) ExecuteMany(ctx context.Context, ids []string, updates map[string]any) error {
	// Bulk operations do not applied here, Later we can improve it if needed
	for _, id := range ids {
		if _, err := uc.Execute(ctx, id, updates); err != nil {
			return err
		}
	}
	return nil
}
