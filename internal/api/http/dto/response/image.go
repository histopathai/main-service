package response

import (
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
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

func NewImageResponses(result query.Result[*model.Image]) *ListResponse[ImageResponse] {
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
