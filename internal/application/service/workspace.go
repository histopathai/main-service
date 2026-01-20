package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type WorkspaceService struct {
	*Service[*model.Workspace]
	usecase *usecase.WorkspaceUseCase
}

func NewWorkspaceService(repo port.Repository[*model.Workspace], uowFactory port.UnitOfWorkFactory) *WorkspaceService {
	return &WorkspaceService{
		Service: &Service[*model.Workspace]{
			repo:       repo,
			uowFactory: uowFactory,
		},

		usecase: usecase.NewWorkspaceUseCase(repo, uowFactory),
	}
}

func (s *WorkspaceService) Create(ctx context.Context, cmd any) (*model.Workspace, error) {

	// Type assertion
	createCmd, ok := cmd.(command.CreateWorkspaceCommand)
	if !ok {
		return nil, errors.NewInternalError("invalid command type for creating workspace", nil)
	}

	entity, err := createCmd.ToEntity()
	if err != nil {
		return nil, err
	}

	return s.usecase.Create(ctx, entity)
}

func (s *WorkspaceService) Update(ctx context.Context, cmd any) error {

	// Type assertion
	updateCmd, ok := cmd.(command.UpdateWorkspaceCommand)
	if !ok {
		return errors.NewInternalError("invalid command type for updating workspace", nil)
	}

	id := updateCmd.GetID()
	updates := updateCmd.GetUpdates()
	if updates == nil {
		return errors.NewValidationError("no updates provided for workspace", map[string]interface{}{
			"id": id,
		})
	}

	return s.usecase.Update(ctx, id, updates)
}
