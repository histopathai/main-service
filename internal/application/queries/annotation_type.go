package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type AnnotationTypeQuery struct {
	*BaseQuery[*model.AnnotationType]
	*HierarchicalQueries[*model.AnnotationType]
}

func NewAnnotationTypeQuery(repo port.AnnotationTypeRepository) *AnnotationTypeQuery {
	return &AnnotationTypeQuery{
		BaseQuery: &BaseQuery[*model.AnnotationType]{
			repo: repo,
		},
		HierarchicalQueries: &HierarchicalQueries[*model.AnnotationType]{
			repo: repo,
		},
	}
}
