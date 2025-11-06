package request

type UploadImageRequest struct {
	PatientID   string `json:"patient_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatorID   string `json:"creator_id" binding:"omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	ContentType string `json:"content_type" binding:"required" example:"image/tiff"`
	Name        string `json:"name" binding:"required" example:"slide1.tiff"`
	Format      string `json:"format" binding:"required" example:"TIFF"`
	Width       *int   `json:"width,omitempty" binding:"omitempty,gte=0" example:"1024"`
	Height      *int   `json:"height,omitempty" binding:"omitempty,gte=0" example:"768"`
	Size        *int64 `json:"size,omitempty" binding:"omitempty,gte=0" example:"2048000"`
}

type ListImageByPatientIDRequest struct {
	PatientID string `form:"patient_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	JSONPaginationRequest
}
