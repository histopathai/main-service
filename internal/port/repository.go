package port

import (
	"context"
	"time"

	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

type Entity interface {
	GetID() string
	SetID(string)
	GetCreatorID() string
	SetCreatorID(string)
	GetCreatedAt() time.Time
	SetCreatedAt(time.Time)
	GetUpdatedAt() time.Time
	SetUpdatedAt(time.Time)
	GetName() string
	SetName(string)
	GetParent() *vobj.ParentRef
	SetParent(*vobj.ParentRef)
	GetEntityType() vobj.EntityType
	SetEntityType(vobj.EntityType)
	HasParent() bool
	IsDeleted() bool
	SetDeleted(bool)
}

type Repository[T Entity] interface {
	Create(ctx context.Context, entity T) (T, error)
	Read(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	SoftDelete(ctx context.Context, id string) error
	Transfer(ctx context.Context, id string, newOwnerID string) error
	FindByFilters(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error)
	SoftDeleteMany(ctx context.Context, ids []string) error
	TransferMany(ctx context.Context, ids []string, newOwnerID string) error
	Count(ctx context.Context, filters []query.Filter) (int64, error)
}

type UnitOfWorkFactory interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repos map[vobj.EntityType]Repository[Entity]) error) error
}
