package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

type MagnificationResponse struct {
	Objective         *float64 `json:"objective,omitempty"`
	NativeLevel       *int     `json:"native_level,omitempty"`
	ScanMagnification *float64 `json:"scan_magnification,omitempty"`
}

type ProcessingInfoResponse struct {
	Status          string    `json:"status"`
	Version         string    `json:"version,omitempty"`
	FailureReason   *string   `json:"failure_reason,omitempty"`
	RetryCount      int       `json:"retry_count"`
	LastProcessedAt time.Time `json:"last_processed_at,omitempty"`
}

type ImageResponse struct {
	ID         string            `json:"id"`
	EntityType string            `json:"entity_type"`
	Parent     ParentRefResponse `json:"parent"`
	CreatorID  string            `json:"creator_id"`
	Name       string            `json:"name"`
	WsID       string            `json:"ws_id"`

	// Basic image properties
	Format string `json:"format"`
	Width  *int   `json:"width,omitempty"`
	Height *int   `json:"height,omitempty"`

	// WSI magnification
	Magnification *MagnificationResponse `json:"magnification,omitempty"`

	// Content references
	OriginContentID    *string `json:"origin_content_id,omitempty"`
	DziContentID       *string `json:"dzi_content_id,omitempty"`
	ThumbnailContentID *string `json:"thumbnail_content_id,omitempty"`
	IndexmapContentID  *string `json:"indexmap_content_id,omitempty"`
	TilesContentID     *string `json:"tiles_content_id,omitempty"`
	ZipTilesContentID  *string `json:"ziptiles_content_id,omitempty"`

	// Processing info
	Processing ProcessingInfoResponse `json:"processing"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

func newProcessingInfoResponse(pi vobj.ProcessingInfo) ProcessingInfoResponse {
	return ProcessingInfoResponse{
		Status:          pi.Status.String(),
		Version:         pi.Version.String(),
		FailureReason:   pi.FailureReason,
		RetryCount:      pi.RetryCount,
		LastProcessedAt: pi.LastProcessedAt,
	}
}

func NewImageResponse(img *model.Image) *ImageResponse {
	parent := ParentRefResponse{
		ID:   img.Parent.ID,
		Type: img.Parent.Type.String(),
	}

	return &ImageResponse{
		ID:                 img.ID,
		EntityType:         img.EntityType.String(),
		Parent:             parent,
		CreatorID:          img.CreatorID,
		Name:               img.Name,
		WsID:               img.WsID,
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
		Processing:         newProcessingInfoResponse(*img.Processing),
		CreatedAt:          img.CreatedAt,
		UpdatedAt:          img.UpdatedAt,
	}
}

func NewImageListResponse(result *query.Result[*model.Image]) *ListResponse[ImageResponse] {
	data := make([]ImageResponse, len(result.Data))
	for i, img := range result.Data {
		dto := NewImageResponse(img)
		data[i] = *dto
	}

	pagination := PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	return &ListResponse[ImageResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}

type UploadImagePayload struct {
	ImageID   string            `json:"image_id"`
	UploadURL string            `json:"upload_url"`
	Headers   map[string]string `json:"headers"`
	Message   string            `json:"message"`
}

type UploadImageResponse struct {
	Data UploadImagePayload `json:"data"`
}

// Swagger response DTOs
type ImageDataResponse struct {
	Data ImageResponse `json:"data"`
}

type ImageListResponse struct {
	Data       []ImageResponse     `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
