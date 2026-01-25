package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type Service[T port.Entity] struct {
	// Add necessary fields here, e.g., repositories, loggers, etc.
	repo       port.Repository[T]
	uowFactory port.UnitOfWorkFactory
}

func (s *Service[T]) Create(ctx context.Context, cmd command.CreateCommand) (T, error) {
	// Implement the logic to create an entity based on the command

	// This is a placeholder implementation
	var entity T

	return entity, errors.NewNotFoundError("not implemented")
}

func (s *Service[T]) Update(ctx context.Context, cmd command.UpdateCommand) error {
	// Implement the logic to update an entity based on the command

	return errors.NewNotFoundError("not implemented")
}

func (s *Service[T]) Get(ctx context.Context, cmd command.ReadCommand) (T, error) {
	return s.repo.Read(ctx, cmd.ID)
}

func (s *Service[T]) Delete(ctx context.Context, cmd command.DeleteCommand) error {
	id := cmd.ID
	return s.repo.SoftDelete(ctx, id)
}

func (s *Service[T]) DeleteMany(ctx context.Context, cmd command.DeleteCommands) error {
	ids := cmd.IDs
	return s.repo.SoftDeleteMany(ctx, ids)
}

func (s *Service[T]) List(ctx context.Context, spec query.Specification) (*query.Result[T], error) {
	// Add is_deleted filter if not present
	hasDeletedFilter := false
	for _, f := range spec.Filters {
		if f.Field == fields.EntityIsDeleted.DomainName() {
			hasDeletedFilter = true
			break
		}
	}

	if !hasDeletedFilter {
		spec.Filters = append(spec.Filters, query.Filter{
			Field:    fields.EntityIsDeleted.DomainName(),
			Operator: query.OpEqual,
			Value:    false,
		})
	}

	return s.repo.Find(ctx, spec)
}

func (s *Service[T]) Count(ctx context.Context, spec query.Specification) (int64, error) {
	hasDeletedFilter := false
	for _, f := range spec.Filters {
		if f.Field == fields.EntityIsDeleted.DomainName() {
			hasDeletedFilter = true
			break
		}
	}

	if !hasDeletedFilter {
		spec.Filters = append(spec.Filters, query.Filter{
			Field:    fields.EntityIsDeleted.DomainName(),
			Operator: query.OpEqual,
			Value:    false,
		})
	}

	return s.repo.Count(ctx, spec)
}

func (s *Service[T]) GetByParentID(ctx context.Context, cmd command.ReadByParentIDCommand) (*query.Result[T], error) {
	id := cmd.ParentID
	parentTypeStr := cmd.ParentType

	if id == "" || parentTypeStr == "" {
		return nil, errors.NewValidationError("parent ID and parent type must be provided", nil)
	}

	parentType, err := vobj.NewParentTypeFromString(parentTypeStr)
	if err != nil {
		return nil, errors.NewValidationError("invalid parent type", map[string]interface{}{
			"parent_type": parentTypeStr,
		})
	}

	if parentType == vobj.ParentTypeNone {
		return nil, errors.NewValidationError("parent type cannot be 'none'", map[string]interface{}{
			"parent_type": parentTypeStr,
		})
	}

	builder := query.NewBuilder()
	builder.Where(fields.EntityParentID.DomainName(), query.OpEqual, id)
	builder.Where(fields.EntityParentType.DomainName(), query.OpEqual, parentType)
	// Also exclude deleted? Default logic usually applies.
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	return s.repo.Find(ctx, builder.Build())
}
