package query

import (
	"fmt"

	"github.com/histopathai/main-service/internal/shared/errors"
)

type Pagination struct {
	Limit   int
	Offset  int
	SortBy  string
	SortDir string
}

func (p *Pagination) ApplyDefaults() {
	if p.Limit <= 0 {
		p.Limit = 20
	}
	if p.Limit > 100 {
		p.Limit = 100
	}
	if p.Offset < 0 {
		p.Offset = 0
	}
	if p.SortBy == "" {
		p.SortBy = "created_at"
	}
	if p.SortDir == "" {
		p.SortDir = "desc"
	}
}

func (p *Pagination) ValidateSortFields(validFields map[string]bool) error {
	if p.SortBy == "" {
		return nil
	}

	if _, ok := validFields[p.SortBy]; ok {
		return nil
	}

	keys := make([]string, 0, len(validFields))
	for k := range validFields {
		keys = append(keys, k)
	}

	return errors.NewValidationError("invalid sort field", map[string]interface{}{
		"sort_by": fmt.Sprintf("must be one of: %v", keys),
	})
}
