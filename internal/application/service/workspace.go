package service

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type WorkspaceService struct {
	*Service[*model.Workspace]
}

func NewWorkspaceService(repo port.Repository[*model.Workspace], uowFactory port.UnitOfWorkFactory) *WorkspaceService {
	return &WorkspaceService{
		Service: &Service[*model.Workspace]{
			repo:       repo,
			uowFactory: uowFactory,
		},
	}
}
