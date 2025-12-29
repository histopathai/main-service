package repository

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type Repository[T any] interface {
	Create(ctx context.Context, entity T) (T, error)
	Read(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, id string, updates map[string]any) error
	Delete(ctx context.Context, id string) error
	FindByFilters(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error)
	Count(ctx context.Context, filters []query.Filter) (int64, error)

	ReadMany(ctx context.Context, ids []string, includeDeleted bool) ([]T, error)
	UpdateMany(ctx context.Context, updates map[string]any, ids []string) error
	DeleteMany(ctx context.Context, ids []string) error
}

type TransferableRepository[T any] interface {
	Repository[T]
	Transfer(ctx context.Context, id, newOwnerID, transferField string) error
	TransferMany(ctx context.Context, ids []string, newOwnerID, transferField string) error
}

type Repositories struct {
	WorkspaceRepo      Repository[*model.Workspace]
	PatientRepo        TransferableRepository[*model.Patient]
	ImageRepo          TransferableRepository[*model.Image]
	AnnotationRepo     Repository[*model.Annotation]
	AnnotationTypeRepo Repository[*model.AnnotationType]
}

type UnitOfWorkFactory interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repos *Repositories) error) error
}
