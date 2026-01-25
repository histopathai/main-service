package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

type MagnificationResponse struct {
	Objective         *float64 `json:"objective,omitempty" example:"40"`
	NativeLevel       *int     `json:"native_level,omitempty" example:"0"`
	ScanMagnification *float64 `json:"scan_magnification,omitempty" example:"40"`
}

func newMagnificationResponse(mag *vobj.OpticalMagnification) *MagnificationResponse {
	if mag == nil {
		return nil
	}
	return &MagnificationResponse{
		Objective:         mag.Objective,
		NativeLevel:       mag.NativeLevel,
		ScanMagnification: mag.ScanMagnification,
	}
}

type ProcessingInfoResponse struct {
	Status          string     `json:"status" example:"processed"`
	Version         string     `json:"version" example:"v1"`
	FailureReason   *string    `json:"failure_reason,omitempty"`
	RetryCount      int        `json:"retry_count" example:"0"`
	LastProcessedAt *time.Time `json:"last_processed_at,omitempty" example:"2024-01-01T12:00:00Z"`
}

func newProcessingInfoResponse(pi *vobj.ProcessingInfo) ProcessingInfoResponse {
	var lastProcessedAt *time.Time
	if !pi.LastProcessedAt.IsZero() {
		lastProcessedAt = &pi.LastProcessedAt
	}

	return ProcessingInfoResponse{
		Status:          pi.Status.String(),
		Version:         pi.Version.String(),
		FailureReason:   pi.FailureReason,
		RetryCount:      pi.RetryCount,
		LastProcessedAt: lastProcessedAt,
	}
}

type ImageResponse struct {
	ID         string             `json:"id" example:"img-123"`
	EntityType string             `json:"entity_type" example:"image"`
	CreatorID  string             `json:"creator_id" example:"user-123"`
	Parent     *ParentRefResponse `json:"parent,omitempty"`
	WsID       string             `json:"ws_id" example:"ws-123"`
	Name       string             `json:"name" example:"slide1.svs"`

	// Basic properties
	Format string `json:"format" example:"svs"`
	Width  *int   `json:"width,omitempty" example:"40000"`
	Height *int   `json:"height,omitempty" example:"30000"`

	// Magnification
	Magnification *MagnificationResponse `json:"magnification,omitempty"`

	// Content references
	OriginContentID    *string `json:"origin_content_id,omitempty" example:"content-123"`
	DziContentID       *string `json:"dzi_content_id,omitempty" example:"content-124"`
	ThumbnailContentID *string `json:"thumbnail_content_id,omitempty" example:"content-125"`
	IndexmapContentID  *string `json:"indexmap_content_id,omitempty" example:"content-126"`
	TilesContentID     *string `json:"tiles_content_id,omitempty" example:"content-127"`
	ZipTilesContentID  *string `json:"ziptiles_content_id,omitempty" example:"content-128"`

	// Processing
	Processing ProcessingInfoResponse `json:"processing"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-02T12:00:00Z"`
}

func NewImageResponse(img *model.Image) *ImageResponse {
	return &ImageResponse{
		ID:                 img.ID,
		EntityType:         img.EntityType.String(),
		CreatorID:          img.CreatorID,
		Parent:             NewParentRefResponse(&img.Parent),
		WsID:               img.WsID,
		Name:               img.Name,
		Format:             img.Format,
		Width:              img.Width,
		Height:             img.Height,
		Magnification:      newMagnificationResponse(img.Magnification),
		OriginContentID:    img.OriginContentID,
		DziContentID:       img.DziContentID,
		ThumbnailContentID: img.ThumbnailContentID,
		IndexmapContentID:  img.IndexmapContentID,
		TilesContentID:     img.TilesContentID,
		ZipTilesContentID:  img.ZipTilesContentID,
		Processing:         newProcessingInfoResponse(img.Processing),
		CreatedAt:          img.CreatedAt,
		UpdatedAt:          img.UpdatedAt,
	}
}

func NewImageListResponse(result *query.Result[*model.Image]) *ListResponse[ImageResponse] {
	data := make([]ImageResponse, len(result.Data))
	for i, img := range result.Data {
		data[i] = *NewImageResponse(img)
	}

	return &ListResponse[ImageResponse]{
		Data: data,
		Pagination: &PaginationResponse{
			Limit:   result.Limit,
			Offset:  result.Offset,
			HasMore: result.HasMore,
		},
	}
}

// Special responses for image upload
type UploadImagePayload struct {
	ImageID   string            `json:"image_id" example:"img-123"`
	UploadURL string            `json:"upload_url" example:"https://storage.googleapis.com/..."`
	Headers   map[string]string `json:"headers,omitempty"`
	Message   string            `json:"message" example:"Upload the image to this URL"`
}

type UploadImageResponse struct {
	Data UploadImagePayload `json:"data"`
}

// Swagger docs
type ImageDataResponse struct {
	Data ImageResponse `json:"data"`
}

type ImageListResponseDoc struct {
	Data       []ImageResponse     `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
