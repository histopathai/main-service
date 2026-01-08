// internal/domain/port/repository.go
package port

import (
	"context"
	"time"

	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

// Entity is the base interface that all domain entities must implement
type Entity interface {
	// Identity methods
	GetID() string
	SetID(id string)

	// Entity type methods
	GetEntityType() vobj.EntityType
	SetEntityType(entityType vobj.EntityType)

	// Name methods
	GetName() string
	SetName(name string)

	// Creator methods
	GetCreatorID() string
	SetCreatorID(creatorID string)

	// Parent methods
	GetParent() *vobj.ParentRef

	// Timestamp methods
	GetCreatedAt() time.Time
	SetCreatedAt(t time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(t time.Time)

	// Deletion methods
	IsDeleted() bool
	SetDeleted(deleted bool)

	// Children methods
	HasChild() bool
	GetChildCount() int64
}

// Repository is the generic repository interface
type Repository[T Entity] interface {
	Create(ctx context.Context, entity T) (T, error)
	Read(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, id string, updates map[string]any) error
	Delete(ctx context.Context, id string) error
	FindByFilters(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error)
	Count(ctx context.Context, filters []query.Filter) (int64, error)

	ReadMany(ctx context.Context, ids []string, includeDeleted bool) ([]T, error)
	UpdateMany(ctx context.Context, updates map[string]any, ids []string) error
	DeleteMany(ctx context.Context, ids []string) error

	GetChildren(ctx context.Context, parentID string, includeDeleted bool) ([]T, error)
	GetChildrenPaginated(ctx context.Context, parentID string, includeDeleted bool, pagination *query.Pagination) (*query.Result[T], error)
}

// TransferableRepository extends Repository with transfer capabilities
type TransferableRepository[T Entity] interface {
	Repository[T]
	Transfer(ctx context.Context, id, newOwnerID, transferField string) error
	TransferMany(ctx context.Context, ids []string, newOwnerID, transferField string) error
}
