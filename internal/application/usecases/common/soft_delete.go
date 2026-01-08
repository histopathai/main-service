package common

import (
	"context"

	"github.com/histopathai/main-service/internal/port"
)

type SoftDeleteUseCase[T port.Entity] struct {
	repo port.Repository[T]
}

func NewSoftDeleteUseCase[T port.Entity](repo port.Repository[T]) *SoftDeleteUseCase[T] {
	return &SoftDeleteUseCase[T]{repo: repo}
}

func (uc *SoftDeleteUseCase[T]) Execute(ctx context.Context, id string) error {
	updates := map[string]any{
		"deleted": true,
	}
	return uc.repo.Update(ctx, id, updates)
}

func (uc *SoftDeleteUseCase[T]) ExecuteMany(ctx context.Context, ids []string) error {
	updates := map[string]any{
		"deleted": true,
	}
	return uc.repo.UpdateMany(ctx, updates, ids)
}
