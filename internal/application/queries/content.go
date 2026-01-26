package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type ContentQuery struct {
	*BaseQuery[*model.Content]
}

func NewContentQuery(
	repo port.Repository[*model.Content],
) *ContentQuery {
	return &ContentQuery{
		BaseQuery: &BaseQuery[*model.Content]{
			repo: repo,
		},
	}
}
