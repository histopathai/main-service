package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

type Service[T port.Entity] struct {
	// Add necessary fields here, e.g., repositories, loggers, etc.
	repo       port.Repository[T]
	uowFactory port.UnitOfWorkFactory
}

func (s *Service[T]) Create(ctx context.Context, cmd command.CreateCommand) (T, error) {
	// Implement the logic to create an entity based on the command

	var entity T

	return entity, nil
}

func (s *Service[T]) Update(ctx context.Context, cmd command.UpdateCommand) error {
	// Implement the logic to update an entity based on the command
	return nil
}

func (s *Service[T]) Read(ctx context.Context, cmd command.ReadCommand) (T, error) {
	return s.repo.Read(ctx, cmd.ID)
}

func (s *Service[T]) Delete(ctx context.Context, cmd command.DeleteCommand) error {
	// Implement the logic to delete an entity based on the command
	return nil
}

func (s *Service[T]) DeleteMany(ctx context.Context, cmd command.DeleteCommands) error {
	// Implement the logic to delete multiple entities based on the command
	return nil
}

func (s *Service[T]) List(ctx context.Context, cmd command.ListCommand) ([]query.Result[T], error) {
	// Implement the logic to list entities based on the command
	return nil, nil
}

func (s *Service[T]) Count(ctx context.Context, cmd command.CountCommand) (int64, error) {
	// Implement the logic to count entities based on the command
	return 0, nil
}

func (s *Service[T]) ReadByParentID(ctx context.Context, cmd command.ListCommand) ([]query.Result[T], error) {
	// Implement the logic to read entities by parent ID based on the command
	return nil, nil
}
