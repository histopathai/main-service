package port

import (
	"context"
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

// Entity interface - tüm entity'lerin uyması gereken kontrat
type Entity interface {
	GetID() string
	SetID(string)
	GetCreatorID() string
	SetCreatorID(string)
	GetCreatedAt() time.Time
	SetCreatedAt(time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(time.Time)
}

// Repository - generic repository interface
// T constraint olarak Entity kullanıyoruz ki metodlara erişebilelim
type Repository[T Entity] interface {
	Create(ctx context.Context, entity T) (T, error)
	Read(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Delete(ctx context.Context, id string) error
	Transfer(ctx context.Context, id string, newOwnerID string) error
	FindByFilters(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error)
	FindByName(ctx context.Context, name string) (T, error)
	BatchDelete(ctx context.Context, ids []string) error
	BatchTransfer(ctx context.Context, ids []string, newOwnerID string) error
	BatchUpdate(ctx context.Context, updates map[string]map[string]interface{}) error
	Count(ctx context.Context, filters []query.Filter) (int64, error)
}

// Concrete repository interfaces - pointer tiplerini kullan
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
