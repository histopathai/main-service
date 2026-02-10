package request

type MagnificationRequest struct {
	Objective         *float64 `json:"objective,omitempty" binding:"omitempty,gt=0" example:"40"`
	NativeLevel       *int     `json:"native_level,omitempty" binding:"omitempty,gte=0" example:"0"`
	ScanMagnification *float64 `json:"scan_magnification,omitempty" binding:"omitempty,gt=0" example:"40"`
}

type ContentRefRequest struct {
	ContentType string `json:"content_type" binding:"required" example:"image/svs"`
	Name        string `json:"name" binding:"required" example:"slide1.svs"`
	Size        int64  `json:"size" binding:"required" example:"1024"`
}

type UploadImageRequest struct {
	Parent   ParentRefRequest    `json:"parent" binding:"required"`
	Name     string              `json:"name" binding:"required" example:"slide1.svs"`
	WsID     string              `json:"ws_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Format   string              `json:"format" binding:"required" example:"svs"`
	Contents []ContentRefRequest `json:"contents" binding:"required"`
	Width    *int                `json:"width,omitempty" binding:"omitempty,gte=0" example:"40000"`
	Height   *int                `json:"height,omitempty" binding:"omitempty,gte=0" example:"30000"`

	Magnification *MagnificationRequest `json:"magnification,omitempty"`
}

type UpdateImageRequest struct {
	CreatorID     *string               `json:"creator_id" binding:"required" example:"1"`
	Name          *string               `json:"name,omitempty" example:"slide1_updated.svs"`
	Width         *int                  `json:"width,omitempty" binding:"omitempty,gte=0"`
	Height        *int                  `json:"height,omitempty" binding:"omitempty,gte=0"`
	Magnification *MagnificationRequest `json:"magnification,omitempty"`
}
