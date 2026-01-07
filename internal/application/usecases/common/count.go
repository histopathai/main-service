package common

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/query"
)

type CountUseCase[T any] struct {
	repo repository.Repository[T]
}

func NewCountUseCase[T any](repo repository.Repository[T]) *CountUseCase[T] {
	return &CountUseCase[T]{repo: repo}
}

func (uc *CountUseCase[T]) Execute(ctx context.Context, filters []query.Filter) (int64, error) {
	return uc.repo.Count(ctx, filters)
}
