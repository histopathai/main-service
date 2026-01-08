package service

import (
	"github.com/histopathai/main-service/internal/application/usecases/common"
	"github.com/histopathai/main-service/internal/application/usecases/composite"
	entityspecific "github.com/histopathai/main-service/internal/application/usecases/entity-specific"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type WorkspaceService struct {
	*BaseService[model.Workspace]
}

func NewWorkspaceService(
	workspaceRepo port.Repository[model.Workspace],
	deleteUc *composite.DeleteUseCase,
) *WorkspaceService {

	createUc := entityspecific.NewCreateWorkspaceUseCase(workspaceRepo)
	updateUc := entityspecific.NewUpdateWorkspaceUseCase(workspaceRepo)

	baseSvc := NewBaseService(
		common.NewReadUseCase(workspaceRepo),
		common.NewListUseCase(workspaceRepo),
		common.NewCountUseCase(workspaceRepo),
		common.NewSoftDeleteUseCase(workspaceRepo),
		common.NewFilterUseCase(workspaceRepo),
		common.NewFilterByParentUseCase(workspaceRepo),
		common.NewFilterByCreatorUseCase(workspaceRepo),
		common.NewFilterByNameUseCase(workspaceRepo),
		deleteUc,
		createUc,
		updateUc,
		vobj.EntityTypeWorkspace,
	)

	return &WorkspaceService{
		BaseService: baseSvc,
	}
}
