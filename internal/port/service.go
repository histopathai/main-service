package port

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/shared/query"
)

type Service[T Entity] interface {
	Create(ctx context.Context, cmd command.CreateCommand) (T, error)
	Get(ctx context.Context, cmd command.ReadCommand) (T, error)
	Update(ctx context.Context, cmd command.UpdateCommand) error
	Delete(ctx context.Context, cmd command.DeleteCommand) error
	DeleteMany(ctx context.Context, cmd command.DeleteCommands) error
	List(ctx context.Context, spec query.Specification) (*query.Result[T], error)
	Count(ctx context.Context, spec query.Specification) (int64, error)
	GetByParentID(ctx context.Context, cmd command.ReadByParentIDCommand) (*query.Result[T], error)
}
