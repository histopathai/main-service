package common

import (
	"context"

	"github.com/histopathai/main-service/internal/port"
)

type ReadUseCase[T any] struct {
	repo port.Repository[T]
}

func NewReadUseCase[T any](repo port.Repository[T]) *ReadUseCase[T] {
	return &ReadUseCase[T]{repo: repo}
}

func (uc *ReadUseCase[T]) Execute(ctx context.Context, id string) (T, error) {
	return uc.repo.Read(ctx, id)
}

func (uc *ReadUseCase[T]) ExecuteMany(ctx context.Context, ids []string, includeDeleted bool) ([]T, error) {
	return uc.repo.ReadMany(ctx, ids, includeDeleted)
}
