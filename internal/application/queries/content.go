package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type ContentQuery struct {
	*BaseQuery[*model.Content]
	*HierarchicalQueries[*model.Content]
}

func NewContentQuery(
	repo port.ContentRepository,
) *ContentQuery {
	return &ContentQuery{
		BaseQuery: &BaseQuery[*model.Content]{
			repo: repo,
		},
		HierarchicalQueries: &HierarchicalQueries[*model.Content]{
			repo: repo,
		},
	}
}
