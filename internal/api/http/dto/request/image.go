package request

type UploadImageRequest struct {
	ID     *string          `json:"id,omitempty" binding:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Parent ParentRefRequest `json:"parent" binding:"required"`

	Name        string `json:"name" binding:"required" example:"slide1.tiff"`
	ContentType string `json:"content_type" binding:"required" example:"image/tiff"`

	Format        string  `json:"format" binding:"required" example:"TIFF"`
	Width         *int    `json:"width,omitempty" binding:"omitempty,gte=0" example:"1024"`
	Height        *int    `json:"height,omitempty" binding:"omitempty,gte=0" example:"768"`
	Size          *int64  `json:"size,omitempty" binding:"omitempty,gte=0" example:"2048000"`
	Status        *string `json:"status,omitempty" binding:"omitempty,oneof=UPLOADED PROCESSING PROCESSED FAILED DELETING" example:"UPLOADED"`
	OriginPath    *string `json:"origin_path,omitempty" binding:"omitempty" example:"s3://bucket/path/to/slide1.tiff"`
	ProcessedPath *string `json:"processed_path,omitempty" binding:"omitempty" example:"s3://bucket/path/to/processed/slide1.tiff"`
}

type UpdateImageRequest struct {
	CreatorID     *string `json:"creator_id,omitempty" binding:"omitempty"`
	Status        *string `json:"status,omitempty" binding:"omitempty,oneof=UPLOADED PROCESSING PROCESSED FAILED DELETING" example:"PROCESSED"`
	Width         *int    `json:"width,omitempty" binding:"omitempty,gte=0" example:"1024"`
	Height        *int    `json:"height,omitempty" binding:"omitempty,gte=0" example:"768"`
	Size          *int64  `json:"size,omitempty" binding:"omitempty,gte=0" example:"2048000"`
	ProcessedPath *string `json:"processed_path,omitempty" binding:"omitempty" example:"s3://bucket/path/to/processed/slide1.tiff"`
}

type ListImagesRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}

var ValidImageSortFields = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"name":       true,
	"size":       true,
	"width":      true,
	"height":     true,
}
