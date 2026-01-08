package entityspecific

import (
	"context"

	"github.com/histopathai/main-service/internal/port"
)

type CreateExecutor[T port.Entity] interface {
	Execute(ctx context.Context, entity *T) (*T, error)
}

type UpdateExecutor[T port.Entity] interface {
	Execute(ctx context.Context, id string, updates map[string]any) (*T, error)
}
