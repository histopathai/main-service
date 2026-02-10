package request

import (
	"fmt"

	"github.com/histopathai/main-service/internal/shared/query"
)

// ============================================================================
// Parent Reference
// ============================================================================

type ParentRefRequest struct {
	ID   string `json:"id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type string `json:"type" binding:"required" example:"workspace"` // workspace, patient, image, annotation_type
}

// ============================================================================
// Query Components
// ============================================================================

type FilterRequest struct {
	Field    string      `json:"field" binding:"required" example:"disease"`
	Operator string      `json:"operator" binding:"required" example:"=="`
	Value    interface{} `json:"value" binding:"required" swaggertype:"string" example:"Cancer"`
}

type SortRequest struct {
	Field     string `json:"field" binding:"required" example:"created_at"`
	Direction string `json:"direction" binding:"required,oneof=asc desc" example:"desc"`
}

type PaginationRequest struct {
	Limit  int `json:"limit" binding:"omitempty,gt=0,lte=100" example:"20"`
	Offset int `json:"offset" binding:"omitempty,gte=0" example:"0"`
}

// ============================================================================
// Generic List Request
// ============================================================================

// ListRequest supports both query parameters and JSON body
// Query param style (simple): /api/v1/workspaces?limit=20&offset=0&sort_by=created_at&sort_dir=desc
// JSON body style (advanced): POST with filters array for complex queries
type ListRequest struct {
	// Simple query param style (most common use case)
	Limit   *int    `form:"limit" json:"limit,omitempty" binding:"omitempty,gt=0,lte=100" example:"20"`
	Offset  *int    `form:"offset" json:"offset,omitempty" binding:"omitempty,gte=0" example:"0"`
	SortBy  *string `form:"sort_by" json:"sort_by,omitempty" example:"created_at"`
	SortDir *string `form:"sort_dir" json:"sort_dir,omitempty" binding:"omitempty,oneof=asc desc" example:"desc"`

	// Advanced filtering (primarily for JSON body, but also supports query params)
	Filters []FilterRequest `json:"filters,omitempty" binding:"omitempty,dive"`
	Sorts   []SortRequest   `json:"sorts,omitempty" binding:"omitempty,dive"`

	// Legacy pagination object support (JSON only)
	Pagination *PaginationRequest `json:"pagination,omitempty"`
}

func (r *ListRequest) ToSpecification() (query.Specification, error) {
	builder := query.NewBuilder()

	for _, f := range r.Filters {
		op := query.Operator(f.Operator)
		if !op.IsValid() {
			return query.Specification{}, fmt.Errorf("invalid operator: %s", f.Operator)
		}
		builder.Where(f.Field, op, f.Value)
	}

	if len(r.Sorts) > 0 {
		for _, s := range r.Sorts {
			dir := query.SortDirection(s.Direction)
			if !dir.IsValid() {
				return query.Specification{}, fmt.Errorf("invalid direction: %s", s.Direction)
			}
			builder.OrderBy(s.Field, dir)
		}
	} else if r.SortBy != nil && r.SortDir != nil {
		dir := query.SortDirection(*r.SortDir)
		if !dir.IsValid() {
			return query.Specification{}, fmt.Errorf("invalid sort direction: %s", *r.SortDir)
		}
		builder.OrderBy(*r.SortBy, dir)
	}

	if r.Pagination != nil {
		builder.Paginate(r.Pagination.Limit, r.Pagination.Offset)
	} else if r.Limit != nil || r.Offset != nil {
		limit := 20 // default
		offset := 0 // default
		if r.Limit != nil {
			limit = *r.Limit
		}
		if r.Offset != nil {
			offset = *r.Offset
		}
		builder.Paginate(limit, offset)
	}

	return builder.Build(), nil
}
