package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type ImageQuery struct {
	*BaseQuery[*model.Image]
}

func NewImageQuery(repo port.ImageRepository) *ImageQuery {
	return &ImageQuery{
		BaseQuery: &BaseQuery[*model.Image]{
			repo: repo,
		},
	}
}
