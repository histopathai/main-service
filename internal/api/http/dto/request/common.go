package request

type QueryPaginationRequest struct {
	Limit   int    `form:"limit,default=20" binding:"omitempty,gt=0" example:"20"`
	Offset  int    `form:"offset,default=0" binding:"omitempty,gte=0" example:"0"`
	SortBy  string `form:"sort_by,default=created_at" example:"created_at"`
	SortDir string `form:"sort_dir,default=desc" binding:"omitempty,oneof=asc desc" example:"desc"`
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
