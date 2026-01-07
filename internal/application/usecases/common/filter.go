package common

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

type FilterUseCase[T any] struct {
	repo repository.Repository[T]
}

func NewFilterUseCase[T any](repo repository.Repository[T]) *FilterUseCase[T] {
	return &FilterUseCase[T]{repo: repo}
}

func (uc *FilterUseCase[T]) Execute(ctx context.Context, filters []query.Filter, pagination *query.Pagination) (*query.Result[T], error) {
	return uc.repo.FindByFilters(ctx, filters, pagination)
}

type FilterByParentUseCase[T any] struct {
	repo repository.Repository[T]
}

func NewFilterByParentUseCase[T any](repo repository.Repository[T]) *FilterByParentUseCase[T] {
	return &FilterByParentUseCase[T]{repo: repo}
}

func (uc *FilterByParentUseCase[T]) Execute(ctx context.Context, parentID string, parentType string, pagination *query.Pagination) (*query.Result[T], error) {
	filters := []query.Filter{
		{
			Field:    constants.ParentIDField,
			Operator: query.OpEqual,
			Value:    parentID,
		},
		{
			Field:    constants.ParentTypeField,
			Operator: query.OpEqual,
			Value:    parentType,
		},
	}
	return uc.repo.FindByFilters(ctx, filters, pagination)
}

type FilterByCreatorUseCase[T any] struct {
	repo repository.Repository[T]
}

func NewFilterByCreatorUseCase[T any](repo repository.Repository[T]) *FilterByCreatorUseCase[T] {
	return &FilterByCreatorUseCase[T]{repo: repo}
}

func (uc *FilterByCreatorUseCase[T]) Execute(ctx context.Context, creatorID string, pagination *query.Pagination) (*query.Result[T], error) {
	filters := []query.Filter{
		{
			Field:    constants.CreatorIDField,
			Operator: query.OpEqual,
			Value:    creatorID,
		},
	}
	return uc.repo.FindByFilters(ctx, filters, pagination)
}

type FilterByNameUseCase[T any] struct {
	repo repository.Repository[T]
}

func NewFilterByNameUseCase[T any](repo repository.Repository[T]) *FilterByNameUseCase[T] {
	return &FilterByNameUseCase[T]{repo: repo}
}

func (uc *FilterByNameUseCase[T]) Execute(ctx context.Context, name string, pagination *query.Pagination, entityType *model.EntityType) (*query.Result[T], error) {

	filters := []query.Filter{
		{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    name,
		},
	}

	if entityType != nil {
		filters = append(filters, query.Filter{
			Field:    constants.EntityTypeField,
			Operator: query.OpEqual,
			Value:    *entityType,
		})
	}

	return uc.repo.FindByFilters(ctx, filters, pagination)
}
