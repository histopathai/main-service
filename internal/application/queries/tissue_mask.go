package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type TissueMaskQuery struct {
	*BaseQuery[*model.TissueMask]
	*HierarchicalQueries[*model.TissueMask]
}

func NewTissueMaskQuery(repo port.TissueMaskRepository) *TissueMaskQuery {
	return &TissueMaskQuery{
		BaseQuery:           &BaseQuery[*model.TissueMask]{repo: repo},
		HierarchicalQueries: &HierarchicalQueries[*model.TissueMask]{repo: repo},
	}
}
