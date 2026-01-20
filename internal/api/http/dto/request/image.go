package request

type UploadImageRequest struct {
	ID     *string          `json:"id,omitempty" binding:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Parent ParentRefRequest `json:"parent" binding:"required"`

	Name   string `json:"name" binding:"required" example:"slide1.svs"`
	Format string `json:"format" binding:"required" example:"svs"`

	// Origin content details
	OriginContent OriginContentRequest `json:"origin_content" binding:"required"`

	// Optional basic fields
	Width  *int `json:"width,omitempty" binding:"omitempty,gte=0" example:"40000"`
	Height *int `json:"height,omitempty" binding:"omitempty,gte=0" example:"30000"`

	// Optional WSI magnification info
	Magnification *MagnificationRequest `json:"magnification,omitempty"`
}

type OriginContentRequest struct {
	Provider    string            `json:"provider" binding:"required,oneof=local s3 gcs azure minio http" example:"s3"`
	Path        string            `json:"path" binding:"required" example:"uploads/550e8400/slide1.svs"`
	ContentType string            `json:"content_type" binding:"required" example:"image/x-aperio-svs"`
	Size        int64             `json:"size" binding:"required,gt=0" example:"524288000"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type MagnificationRequest struct {
	Objective         *float64 `json:"objective,omitempty" binding:"omitempty,gt=0" example:"40"`
	NativeLevel       *int     `json:"native_level,omitempty" binding:"omitempty,gte=0" example:"0"`
	ScanMagnification *float64 `json:"scan_magnification,omitempty" binding:"omitempty,gt=0" example:"40"`
}

type ContentDataRequest struct {
	Provider    string            `json:"provider" binding:"required,oneof=local s3 gcs azure minio http"`
	Path        string            `json:"path" binding:"required"`
	ContentType string            `json:"content_type" binding:"required"`
	Size        int64             `json:"size" binding:"required,gt=0"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type ProcessedContentRequest struct {
	DZI       *ContentDataRequest `json:"dzi,omitempty"`
	Tiles     *ContentDataRequest `json:"tiles,omitempty"`
	Thumbnail *ContentDataRequest `json:"thumbnail,omitempty"`
	IndexMap  *ContentDataRequest `json:"index_map,omitempty"`
}

type ProcessingInfoRequest struct {
	Status          *string `json:"status,omitempty" binding:"omitempty,oneof=PENDING PROCESSING PROCESSED FAILED DELETING"`
	Version         *string `json:"version,omitempty" binding:"omitempty,oneof=v1 v2"`
	FailureReason   *string `json:"failure_reason,omitempty"`
	RetryCount      *int    `json:"retry_count,omitempty" binding:"omitempty,gte=0"`
	LastProcessedAt *string `json:"last_processed_at,omitempty"` // ISO 8601 format
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

	// Origin content (rarely updated)
	OriginContent *ContentDataRequest `json:"origin_content,omitempty"`

	// Processed content
	ProcessedContent *ProcessedContentRequest `json:"processed_content,omitempty"`

	// Processing info
	Processing *ProcessingInfoRequest `json:"processing,omitempty"`
}

type ListImagesRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}
