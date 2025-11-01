package response

import (
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type PointResponse struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func NewPointResponse(p []model.Point) []PointResponse {

	jsonPoints := make([]PointResponse, len(p))
	for i, point := range p {
		jsonPoints[i] = PointResponse{
			X: point.X,
			Y: point.Y,
		}
	}
	return jsonPoints

}

type AnnotationResponse struct {
	ID          string          `json:"id"`
	ImageID     string          `json:"image_id"`
	AnnotatorID string          `json:"annotator_id"`
	Polygon     []PointResponse `json:"polygon"`
	Score       *float64        `json:"score,omitempty"`
	Class       *string         `json:"class,omitempty"`
	Description *string         `json:"description,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func NewAnnotationResponse(a *model.Annotation) *AnnotationResponse {
	return &AnnotationResponse{
		ID:          a.ID,
		ImageID:     a.ImageID,
		AnnotatorID: a.AnnotatorID,
		Polygon:     NewPointResponse(a.Polygon),
		Score:       a.Score,
		Class:       a.Class,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func NewAnnotationListResponse(result *query.Result[*model.Annotation]) *ListResponse[*AnnotationResponse] {

	data := make([]*AnnotationResponse, len(result.Data))
	for i, a := range result.Data {
		dto := NewAnnotationResponse(a)
		data[i] = dto
	}

	pagination := PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
		Total:   result.Total,
	}

	return &ListResponse[*AnnotationResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}
