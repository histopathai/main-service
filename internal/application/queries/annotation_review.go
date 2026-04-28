package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type AnnotationReviewQuery struct {
	*BaseQuery[*model.AnnotationReview]
	*HierarchicalQueries[*model.AnnotationReview]
}

func NewAnnotationReviewQuery(repo port.AnnotationReviewRepository) *AnnotationReviewQuery {
	return &AnnotationReviewQuery{
		BaseQuery: &BaseQuery[*model.AnnotationReview]{
			repo: repo,
		},
		HierarchicalQueries: &HierarchicalQueries[*model.AnnotationReview]{
			repo: repo,
		},
	}
}
