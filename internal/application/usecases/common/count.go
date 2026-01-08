package common

import (
	"context"

	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

type CountUseCase[T any] struct {
	repo port.Repository[T]
}

func NewCountUseCase[T any](repo port.Repository[T]) *CountUseCase[T] {
	return &CountUseCase[T]{repo: repo}
}

func (uc *CountUseCase[T]) Execute(ctx context.Context, filters []query.Filter) (int64, error) {
	return uc.repo.Count(ctx, filters)
}
