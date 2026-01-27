package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type AnnotationQuery struct {
	*BaseQuery[*model.Annotation]
	*HierarchicalQueries[*model.Annotation]
}

func NewAnnotationQuery(repo port.AnnotationRepository) *AnnotationQuery {
	return &AnnotationQuery{
		BaseQuery: &BaseQuery[*model.Annotation]{
			repo: repo,
		},
		HierarchicalQueries: &HierarchicalQueries[*model.Annotation]{
			repo: repo,
		},
	}
}
