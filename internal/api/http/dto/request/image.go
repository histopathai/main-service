package request

type CreateImageRequest struct {
	Name          string                `json:"name" binding:"required" example:"slide1.svs"`
	Parent        ParentRefRequest      `json:"parent" binding:"required"`
	WsID          string                `json:"ws_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Format        string                `json:"format" binding:"required" example:"svs"`
	ContentType   string                `json:"content_type" binding:"required" example:"image/svs"`
	Width         *int                  `json:"width,omitempty" binding:"omitempty,gte=0" example:"40000"`
	Height        *int                  `json:"height,omitempty" binding:"omitempty,gte=0" example:"30000"`
	Size          *int64                `json:"size,omitempty" binding:"omitempty,gte=0" example:"524288000"`
	Magnification *MagnificationRequest `json:"magnification,omitempty"`
}

type MagnificationRequest struct {
	Objective         *float64 `json:"objective,omitempty" binding:"omitempty,gt=0" example:"40"`
	NativeLevel       *int     `json:"native_level,omitempty" binding:"omitempty,gte=0" example:"0"`
	ScanMagnification *float64 `json:"scan_magnification,omitempty" binding:"omitempty,gt=0" example:"40"`
}

type UpdateImageRequest struct {
	CreatorID *string `json:"creator_id,omitempty" binding:"omitempty"`
	Name      *string `json:"name,omitempty" binding:"omitempty"`

	// Basic fields
	Width  *int    `json:"width,omitempty" binding:"omitempty,gte=0"`
	Height *int    `json:"height,omitempty" binding:"omitempty,gte=0"`
	Size   *int64  `json:"size,omitempty" binding:"omitempty,gte=0"`
	Format *string `json:"format,omitempty"`

	// Magnification
	Magnification *MagnificationRequest `json:"magnification,omitempty"`
}

type ListImagesRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}
