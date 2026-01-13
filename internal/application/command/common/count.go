package common

import "github.com/histopathai/main-service/internal/shared/query"

type CountCommand struct {
	Filters []query.Filter
}
