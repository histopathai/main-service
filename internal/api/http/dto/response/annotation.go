package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PointResponse struct {
	X float64 `json:"x" example:"100.5"`
	Y float64 `json:"y" example:"200.3"`
}

func NewPointResponse(points []vobj.Point) []PointResponse {
	result := make([]PointResponse, len(points))
	for i, p := range points {
		result[i] = PointResponse{X: p.X, Y: p.Y}
	}
	return result
}

type AnnotationResponse struct {
	ID         string             `json:"id" example:"anno-123"`
	EntityType string             `json:"entity_type" example:"annotation"`
	CreatorID  string             `json:"creator_id" example:"user-123"`
	Parent     *ParentRefResponse `json:"parent,omitempty"`
	WsID       string             `json:"ws_id" example:"ws-123"`
	Name       string             `json:"name" example:"Tumor Region"`
	TagType    string             `json:"tag_type" example:"number"`
	Value      interface{}        `json:"value" swaggertype:"string" example:"3.5"`
	IsGlobal   bool               `json:"is_global" example:"false"`
	Color      *string            `json:"color,omitempty" example:"#FF0000"`
	Polygon    []PointResponse    `json:"polygon,omitempty"`
	CreatedAt  time.Time          `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt  time.Time          `json:"updated_at" example:"2024-01-02T12:00:00Z"`
}

func NewAnnotationResponse(a *model.Annotation) *AnnotationResponse {
	var polygon []PointResponse
	if a.Polygon != nil {
		polygon = NewPointResponse(*a.Polygon)
	}

	return &AnnotationResponse{
		ID:         a.ID,
		EntityType: a.EntityType.String(),
		CreatorID:  a.CreatorID,
		Parent:     NewParentRefResponse(&a.Parent),
		WsID:       a.WsID,
		Name:       a.Name,
		TagType:    a.TagType.String(),
		Value:      a.Value,
		IsGlobal:   a.IsGlobal,
		Color:      a.Color,
		Polygon:    polygon,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
	}
}

func NewAnnotationListResponse(result *query.Result[*model.Annotation]) *ListResponse[AnnotationResponse] {
	data := make([]AnnotationResponse, len(result.Data))
	for i, a := range result.Data {
		data[i] = *NewAnnotationResponse(a)
	}

	return &ListResponse[AnnotationResponse]{
		Data: data,
		Pagination: &PaginationResponse{
			Limit:   result.Limit,
			Offset:  result.Offset,
			HasMore: result.HasMore,
		},
	}
}

// Swagger docs
type AnnotationDataResponse struct {
	Data AnnotationResponse `json:"data"`
}

type AnnotationListResponseDoc struct {
	Data       []AnnotationResponse `json:"data"`
	Pagination *PaginationResponse  `json:"pagination,omitempty"`
}
