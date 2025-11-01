package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ImageResponse struct {
	ID        string    `json:"id"`
	PatientID string    `json:"patient_id"`
	CreatorID string    `json:"creator_id"`
	Name      string    `json:"name"`
	Format    string    `json:"format"`
	Width     *int      `json:"width,omitempty"`
	Height    *int      `json:"height,omitempty"`
	Size      *int64    `json:"size,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewImageResponse(img *model.Image) *ImageResponse {
	return &ImageResponse{
		ID:        img.ID,
		PatientID: img.PatientID,
		CreatorID: img.CreatorID,
		Name:      img.Name,
		Format:    img.Format,
		Width:     img.Width,
		Height:    img.Height,
		Size:      img.Size,
		CreatedAt: img.CreatedAt,
		UpdatedAt: img.UpdatedAt,
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
		Total:   result.Total,
		HasMore: result.HasMore,
	}

	return &ListResponse[ImageResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}

type UploadImagePayload struct {
	UploadURL string `json:"upload_url" example:"https://storage.googleapis.com/..."`
	Message   string `json:"message" example:"Use this URL to upload the image via a PUT request."`
}
type UploadImageResponse struct {
	Data UploadImagePayload `json:"data"`
}

// Added DTOs for swagger responses. Swagger requires a concrete type for response schemas.
type ImageDataResponse struct {
	Data ImageResponse `json:"data"`
}

type ImageListResponse struct {
	Data       []ImageResponse     `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
