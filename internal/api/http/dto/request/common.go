package request

import (
	"github.com/histopathai/main-service/internal/shared/query"
)

// QueryPaginationRequest - Query parameter binding with validation
type QueryPaginationRequest struct {
	Limit   int    `form:"limit" binding:"omitempty,gt=0,lte=100" example:"20"`
	Offset  int    `form:"offset" binding:"omitempty,gte=0" example:"0"`
	SortBy  string `form:"sort_by" example:"created_at"`
	SortDir string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// JSONPaginationRequest - JSON body binding with validation
type JSONPaginationRequest struct {
	Limit   int    `json:"limit" binding:"omitempty,gt=0,lte=100" example:"20"`
	Offset  int    `json:"offset" binding:"omitempty,gte=0" example:"0"`
	SortBy  string `json:"sort_by" example:"created_at"`
	SortDir string `json:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// ToPagination - Mapper for QueryPaginationRequest
func (qpr *QueryPaginationRequest) ToPagination() *query.Pagination {
	return &query.Pagination{
		Limit:   qpr.Limit,
		Offset:  qpr.Offset,
		SortBy:  qpr.SortBy,
		SortDir: qpr.SortDir,
	}
}

// ToPagination - Mapper for JSONPaginationRequest
func (jpr *JSONPaginationRequest) ToPagination() *query.Pagination {
	return &query.Pagination{
		Limit:   jpr.Limit,
		Offset:  jpr.Offset,
		SortBy:  jpr.SortBy,
		SortDir: jpr.SortDir,
	}
}

// Filter Request
type JSONFilterRequest struct {
	Field    string      `json:"field" binding:"required" example:"disease"`
	Operator string      `json:"operator" binding:"required" example:"=="`
	Value    interface{} `json:"value" binding:"required" example:"Breast Cancer"`
}

// Batch Operations
type BatchDeleteRequest struct {
	IDs []string `json:"ids" binding:"required,min=1,dive,required" example:"['id1', 'id2']"`
}

type BatchTransferRequest struct {
	IDs    []string `json:"ids" binding:"required,min=1,dive,required" example:"['id1', 'id2']"`
	Target string   `json:"target" binding:"required" example:"workspace-123"`
}
