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

type ListRequest struct {
	Filters    []FilterRequest    `json:"filters,omitempty" binding:"omitempty,dive"`
	Sorts      []SortRequest      `json:"sorts,omitempty" binding:"omitempty,dive"`
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

	for _, s := range r.Sorts {
		dir := query.SortDirection(s.Direction)
		if !dir.IsValid() {
			return query.Specification{}, fmt.Errorf("invalid direction: %s", s.Direction)
		}
		builder.OrderBy(s.Field, dir)
	}

	if r.Pagination != nil {
		builder.Paginate(r.Pagination.Limit, r.Pagination.Offset)
	}

	return builder.Build(), nil
}
