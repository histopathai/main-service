package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type WorkspaceQuery struct {
	*BaseQuery[*model.Workspace]
}

func NewWorkspaceQuery(repo port.Repository[*model.Workspace]) *WorkspaceQuery {
	return &WorkspaceQuery{
		BaseQuery: &BaseQuery[*model.Workspace]{
			repo: repo,
		},
	}
}
