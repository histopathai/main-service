package commands

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

type CreateCommand[T any] interface {
	ToEntity() (T, error)
}

type UpdateCommand[T any] interface {
	GetID() string
	ApplyTo(entity T) (T, error)
	GetUpdates() (map[string]any, error)
}

type ReadCommand[T any] struct {
	ID string
}

type ListCommand[T any] struct {
	Pagination *query.Pagination
	Filters    []query.Filter
}

type CountCommand[T any] struct {
	Filters []query.Filter
}

func NewCountCommand[T any](filters []query.Filter) *CountCommand[T] {
	return &CountCommand[T]{Filters: filters}
}

func NewListCommand[T any](Limit, Offset int, SortBy, SortDir string) *ListCommand[T] {

	pg := query.Pagination{
		Limit:   Limit,
		Offset:  Offset,
		SortBy:  SortBy,
		SortDir: SortDir,
	}
	pg.ApplyDefaults()

	return &ListCommand[T]{
		Pagination: &pg,
		Filters:    []query.Filter{},
	}
}

type ReadManyCommand[T any] struct {
	IDs []string
}

type DeleteCommand[T any] struct {
	ID   string
	Role string
}

type DeleteManyCommand[T any] struct {
	IDs  []string
	Role string
}

type SoftDeleteCommand[T any] struct {
	ID string
}

type SoftDeleteManyCommand[T any] struct {
	IDs []string
}

type TransfarableCommand interface {
	GetParentID() string
	GetParentType() string
	SetParentID(newParentID string)
}

type TransferCommand struct {
	ID          string
	OldParentID string

	NewParentID string
}

type TransferManyCommand struct {
	IDs         []string
	OldParentID string
	NewParentID string
}

type FindByParentCommand[T any] struct {
	ParentRef  vobj.ParentRef
	Pagination *query.Pagination
}

type FindByCreatorCommand[T any] struct {
	CreatorID  string
	Pagination *query.Pagination
}

type FindByNameCommand[T any] struct {
	Name       string
	Pagination *query.Pagination
}
