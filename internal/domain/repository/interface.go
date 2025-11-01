package repository

import (
	"context"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type Repository[T any] interface {
	Create(ctx context.Context, entity T) (T, error)
	Read(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	Transfer(ctx context.Context, id string, newOwnerID string) error
	FindByFilters(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error)
	FindByName(ctx context.Context, name string) (T, error)
}

type WorkspaceRepository interface {
	Repository[*model.Workspace]
}
type PatientRepository interface {
	Repository[*model.Patient]
}
type ImageRepository interface {
	Repository[*model.Image]
}
type AnnotationRepository interface {
	Repository[*model.Annotation]
}
type AnnotationTypeRepository interface {
	Repository[*model.AnnotationType]
}

type Repositories struct {
	WorkspaceRepo      WorkspaceRepository
	PatientRepo        PatientRepository
	ImageRepo          ImageRepository
	AnnotationRepo     AnnotationRepository
	AnnotationTypeRepo AnnotationTypeRepository
}

type UnitOfWorkFactory interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repos *Repositories) error) error
}
