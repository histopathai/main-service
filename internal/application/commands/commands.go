package commands

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

// Generic command interfaces artık port.Entity kullanıyor
type CreateCommand[T port.Entity] interface {
	ToEntity() (T, error)
}

type UpdateCommand[T port.Entity] interface {
	GetID() string
	ApplyTo(entity T) (T, error)
	GetUpdates() (map[string]any, error)
}

type ReadCommand[T port.Entity] struct {
	ID string
}

type ListCommand[T port.Entity] struct {
	Pagination *query.Pagination
	Filters    []query.Filter
}

type CountCommand[T port.Entity] struct {
	Filters []query.Filter
}

func NewCountCommand[T port.Entity](filters []query.Filter) *CountCommand[T] {
	return &CountCommand[T]{Filters: filters}
}

func NewListCommand[T port.Entity](Limit, Offset int, SortBy, SortDir string) *ListCommand[T] {
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

type ReadManyCommand[T port.Entity] struct {
	IDs []string
}

type DeleteCommand[T port.Entity] struct {
	ID   string
	Role string
}

type DeleteManyCommand[T port.Entity] struct {
	IDs  []string
	Role string
}

type SoftDeleteCommand[T port.Entity] struct {
	ID string
}

type SoftDeleteManyCommand[T port.Entity] struct {
	IDs []string
}

type TransferableCommand interface {
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

type FindByParentCommand[T port.Entity] struct {
	ParentRef  vobj.ParentRef
	Pagination *query.Pagination
}

type FindByCreatorCommand[T port.Entity] struct {
	CreatorID  string
	Pagination *query.Pagination
}

type FindByNameCommand[T port.Entity] struct {
	Name       string
	Pagination *query.Pagination
}
