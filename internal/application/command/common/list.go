package common

import (
	"github.com/histopathai/main-service/internal/shared/query"
)

type ListCommand struct {
	Pagination *query.Pagination
	Filters    []query.Filter
}
