package response

import "github.com/histopathai/main-service/internal/domain/vobj"

// ============================================================================
// Error Response
// ============================================================================

type ErrorResponse struct {
	ErrorType string                 `json:"error_type,omitempty" example:"VALIDATION_ERROR"`
	Message   string                 `json:"message" example:"Invalid request"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// ============================================================================
// Pagination Response
// ============================================================================

type PaginationResponse struct {
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	HasMore bool `json:"has_more" example:"true"`
}

// ============================================================================
// Generic Responses
// ============================================================================

// DataResponse - Single item response
type DataResponse[T any] struct {
	Data T `json:"data"`
}

// ListResponse - List response with pagination
type ListResponse[T any] struct {
	Data       []T                 `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// CountResponse - Count response
type CountResponse struct {
	Count int64 `json:"count" example:"100"`
}

// MessageResponse - Simple message response
type MessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

// ============================================================================
// Parent Reference Response
// ============================================================================

type ParentRefResponse struct {
	ID   string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Type string `json:"type" example:"workspace"`
}

func NewParentRefResponse(ref *vobj.ParentRef) *ParentRefResponse {
	if ref == nil || ref.IsEmpty() {
		return nil
	}
	return &ParentRefResponse{
		ID:   ref.ID,
		Type: ref.Type.String(),
	}
}
