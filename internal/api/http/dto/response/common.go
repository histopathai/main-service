package response

import "github.com/histopathai/main-service/internal/domain/vobj"

type ErrorResponse struct {
	ErrorType string                 `json:"error_type,omitempty" example:"VALIDATION_ERROR"`
	Message   string                 `json:"message" example:"İlgili hasta bulunamadı."`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type PaginationResponse struct {
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	HasMore bool `json:"has_more" example:"true"`
}

type ListResponse[T any] struct {
	Data       []T                 `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type DataResponse[T any] struct {
	Data T `json:"data"`
}

type CountResponse struct {
	Count int64 `json:"count" example:"100"`
}

type ParentRefResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
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
