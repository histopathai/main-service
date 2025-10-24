package repository

import (
	"context"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type Repository[T any] interface {
	Create(ctx context.Context, entity *T) (*T, error)
	GetByID(ctx context.Context, id string) error
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	GetByCriteria(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error)
}

type Transaction interface {
	Create(ctx context.Context, collection string, data interface{}) (string, error)
	Get(ctx context.Context, collection string, id string, result interface{}) error
	Update(ctx context.Context, collection string, id string, data interface{}) error
	Delete(ctx context.Context, collection string, id string) error
	FindByFilters(ctx context.Context, collection string, filters []query.Filter, paginationOpts *query.Pagination, result interface{}) (int, error)
}

type TransactionProvider interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Transaction) error) error
}

type WorkSpaceRepository interface {
	Repository[model.Workspace]
	TransactionProvider
	GetByCreatorID(ctx context.Context, creatorID string) (*query.Result[model.Workspace], error)
	GetByeOrganType(ctx context.Context, organType string) (*query.Result[model.Workspace], error)
}

type PatientRepository interface {
	Repository[model.Patient]
	TransactionProvider
	GetByWorkSpaceID(ctx context.Context, workspaceID string, paginationOpts *query.Pagination) (*query.Result[model.Patient], error)
}

type ImageRepository interface {
	Repository[model.Image]
	TransactionProvider
	GetByPatientID(ctx context.Context, patientID string, paginationOpts *query.Pagination) (*query.Result[model.Image], error)
}

type AnnotationRepository interface {
	Repository[model.Annotation]
	TransactionProvider
	GetByImageID(ctx context.Context, imageID string, paginationOpts *query.Pagination) (*query.Result[model.Annotation], error)
}

type AnnotationTypeRepository interface {
	Repository[model.AnnotationType]
	TransactionProvider
}
