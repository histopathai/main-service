package port

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type Queries[T Entity] interface {
	// Queries (no business logic)
	Get(ctx context.Context, id string) (T, error)
	List(ctx context.Context, spec query.Specification) (*query.Result[T], error)
	Count(ctx context.Context, spec query.Specification) (int64, error)

	// Simple mutations (no business logic)
	SoftDelete(ctx context.Context, id string) error
	SoftDeleteMany(ctx context.Context, ids []string) error
}

type HierarchicalQueries[T Entity] interface {
	GetByParentID(ctx context.Context, spec query.Specification, parentID string) (*query.Result[T], error)
	GetByWsID(ctx context.Context, spec query.Specification, workspaceID string) (*query.Result[T], error)
}

type WorkspaceQeury interface {
	Queries[*model.Workspace]
}

type PatientQuery interface {
	Queries[*model.Patient]
	HierarchicalQueries[*model.Patient]
}

type AnnotationTypeQuery interface {
	Queries[*model.AnnotationType]
	HierarchicalQueries[*model.AnnotationType]
}
type AnnotationQuery interface {
	Queries[*model.Annotation]
	HierarchicalQueries[*model.Annotation]
}
type ImageQuery interface {
	Queries[*model.Image]
	HierarchicalQueries[*model.Image]
}
type ContentQuery interface {
	Queries[*model.Content]
	HierarchicalQueries[*model.Content]
}
