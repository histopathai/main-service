package queries

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

// ======================================
// BaseQuery
// ======================================
type BaseQuery[T port.Entity] struct {
	repo port.Repository[T]
}

func (s *BaseQuery[T]) Get(ctx context.Context, id string) (T, error) {
	return s.repo.Read(ctx, id)
}

func (s *BaseQuery[T]) SoftDelete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}

func (s *BaseQuery[T]) SoftDeleteMany(ctx context.Context, ids []string) error {
	return s.repo.SoftDeleteMany(ctx, ids)
}

func (s *BaseQuery[T]) List(ctx context.Context, spec query.Specification) (*query.Result[T], error) {
	// Add is_deleted filter if not present
	deletedFilterCheck(&spec, false)
	return s.repo.Find(ctx, spec)
}

func (s *BaseQuery[T]) Count(ctx context.Context, spec query.Specification) (int64, error) {
	deletedFilterCheck(&spec, false)
	return s.repo.Count(ctx, spec)
}

// ======================================
// HierarchicalQueries
// ======================================
type HierarchicalQueries[T port.Entity] struct {
	repo port.Repository[T]
}

func (s *HierarchicalQueries[T]) GetByParentID(ctx context.Context, spec query.Specification, parentID string) (*query.Result[T], error) {
	// Add parent ID and deleted filters to the provided spec
	additionalFilters := []query.Filter{
		{
			Field:    fields.EntityParentID.DomainName(),
			Operator: query.OpEqual,
			Value:    parentID,
		},
		{
			Field:    fields.EntityIsDeleted.DomainName(),
			Operator: query.OpEqual,
			Value:    false,
		},
	}

	spec.Filters = append(spec.Filters, additionalFilters...)
	return s.repo.Find(ctx, spec)
}

func (s *HierarchicalQueries[T]) GetByWsID(ctx context.Context, spec query.Specification, wsID string) (*query.Result[T], error) {
	// Add workspace ID and deleted filters to the provided spec
	additionalFilters := []query.Filter{
		{
			Field:    "WsID",
			Operator: query.OpEqual,
			Value:    wsID,
		},
		{
			Field:    fields.EntityIsDeleted.DomainName(),
			Operator: query.OpEqual,
			Value:    false,
		},
	}

	spec.Filters = append(spec.Filters, additionalFilters...)
	return s.repo.Find(ctx, spec)
}

func deletedFilterCheck(spec *query.Specification, includeDeleted bool) {
	if includeDeleted {
		return
	}
	hasDeletedFilter := false
	for _, f := range spec.Filters {
		if f.Field == fields.EntityIsDeleted.DomainName() {
			hasDeletedFilter = true
			break
		}
	}
	if !hasDeletedFilter {
		spec.Filters = append(spec.Filters, query.Filter{
			Field:    fields.EntityIsDeleted.DomainName(),
			Operator: query.OpEqual,
			Value:    false,
		})
	}
}
