package common

import (
	"context"

	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ListUseCase[T any] struct {
	repo port.Repository[T]
}

func NewListUseCase[T any](repo port.Repository[T]) *ListUseCase[T] {
	return &ListUseCase[T]{repo: repo}
}

func (uc *ListUseCase[T]) Execute(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error) {
	return uc.repo.FindByFilters(ctx, filters, paginationOpts)
}
