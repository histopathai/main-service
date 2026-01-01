package commands

import (
	"github.com/histopathai/main-service/internal/shared/query"
)

type CreateCommand[T any] interface {
	ToEntity() (T, error)
}

type UpdateCommand[T any] interface {
	GetID() string
	ApplyTo(entity T) (T, error)
}

type ReadCommand[T any] struct {
	EntityID string
}

type ListCommand[T any] struct {
	Pagination query.Pagination
	Filters    []query.Filter
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
		Pagination: pg,
		Filters:    []query.Filter{},
	}
}

type ReadManyCommand[T any] struct {
	EntityIDs      []string
	IncludeDeleted bool
}

type DeleteCommand[T any] struct {
	EntityID string
	Role     string
}

type DeleteManyCommand[T any] struct {
	EntityIDs []string
	Role      string
}

type SoftDeleteCommand[T any] struct {
	EntityID string
}

type SoftDeleteManyCommand[T any] struct {
	EntityIDs []string
}

type TransfarableCommand interface {
	GetParentID() string
	GetParentType() string
	SetParentID(newParentID string)
}

type TransferCommand struct {
	EntityID    string
	OldParentID string

	NewParentID string
}

type TransferManyCommand struct {
	EntityIDs   []string
	OldParentID string
	NewParentID string
}
