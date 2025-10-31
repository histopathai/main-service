package response

type ErrorResponse struct {
	ErrorType string                 `json:"error_type,omitempty" example:"VALIDATION_ERROR"`
	Message   string                 `json:"message" example:"İlgili hasta bulunamadı."`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type PaginationResponse struct {
	Limit   int  `json:"limit" example:"20"`
	Offset  int  `json:"offset" example:"0"`
	Total   int  `json:"total,omitempty" example:"150"`
	HasMore bool `json:"has_more" example:"true"`
}

type ListResponse[T any] struct {
	Data       []T                 `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

type DataResponse[T any] struct {
	Data T `json:"data"`
}
