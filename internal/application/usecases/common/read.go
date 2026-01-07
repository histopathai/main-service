package common

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/repository"
)

type ReadUseCase[T any] struct {
	repo repository.Repository[T]
}

func NewReadUseCase[T any](repo repository.Repository[T]) *ReadUseCase[T] {
	return &ReadUseCase[T]{repo: repo}
}

func (uc *ReadUseCase[T]) Execute(ctx context.Context, id string) (T, error) {
	return uc.repo.Read(ctx, id)
}

func (uc *ReadUseCase[T]) ExecuteMany(ctx context.Context, ids []string, includeDeleted bool) ([]T, error) {
	return uc.repo.ReadMany(ctx, ids, includeDeleted)
}
