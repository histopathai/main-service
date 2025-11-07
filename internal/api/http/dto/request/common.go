package request

import (
	"fmt"

	"github.com/histopathai/main-service/internal/shared/errors"
)

const (
	DefaultLimit   = 20
	MaxLimit       = 100
	DefaultOffset  = 0
	DefaultSortBy  = "created_at"
	DefaultSortDir = "desc"
)

type QueryPaginationRequest struct {
	Limit   int    `form:"limit" binding:"omitempty,gt=0,lte=100" example:"20"`
	Offset  int    `form:"offset" binding:"omitempty,gte=0" example:"0"`
	SortBy  string `form:"sort_by" example:"created_at"`
	SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

func (qpr *QueryPaginationRequest) ApplyDefaults() {
	if qpr.Limit <= 0 {
		qpr.Limit = DefaultLimit
	}
	if qpr.Limit > MaxLimit {
		qpr.Limit = MaxLimit
	}
	if qpr.Offset < 0 {
		qpr.Offset = DefaultOffset
	}
	if qpr.SortBy == "" {
		qpr.SortBy = DefaultSortBy
	}
	if qpr.SortDir == "" {
		qpr.SortDir = DefaultSortDir
	}
}

func (qpr *QueryPaginationRequest) ValidateSortFields(validFields map[string]bool) error {
	if validFields[qpr.SortBy] {
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

type JSONPaginationRequest struct {
	Limit   int    `json:"limit" binding:"omitempty,gt=0" example:"20"`
	Offset  int    `json:"offset" binding:"omitempty,gte=0" example:"0"`
	SortBy  string `json:"sort_by" example:"created_at"`
	SortDir string `json:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

type JSONFilterRequest struct {
	Field    string      `json:"field" binding:"required" example:"disease"`
	Operator string      `json:"operator" binding:"required" example:"=="`
	Value    interface{} `json:"value" binding:"required" example:"Breast Cancer"`
}
