package common

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ListUseCase[T any] struct {
	repo repository.Repository[T]
}

func NewListUseCase[T any](repo repository.Repository[T]) *ListUseCase[T] {
	return &ListUseCase[T]{repo: repo}
}

func (uc *ListUseCase[T]) Execute(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error) {
	return uc.repo.FindByFilters(ctx, filters, paginationOpts)
}
