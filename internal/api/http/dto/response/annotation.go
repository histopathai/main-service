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
	jsonPoints := make([]PointResponse, len(points))
	for i, point := range points {
		jsonPoints[i] = PointResponse{
			X: point.X,
			Y: point.Y,
		}
	}
	return jsonPoints
}

// TagValueResponse - swagger compatible version
// Note: Value is kept as interface{} but swagger will see it as string
type TagValueResponse struct {
	TagType string `json:"tag_type" example:"NUMBER"`
	TagName string `json:"tag_name" example:"Grade"`
	// swaggertype: string - This tells swagger to treat it as string in docs
	Value  interface{} `json:"value" swaggertype:"string" example:"3.5"`
	Color  *string     `json:"color,omitempty" example:"#FF0000"`
	Global bool        `json:"global" example:"false"`
}

type AnnotationResponse struct {
	ID          string             `json:"id" example:"anno-123"`
	EntityType  string             `json:"entity_type" example:"annotation"`
	Name        *string            `json:"name,omitempty" example:"Sample Annotation"`
	CreatorID   string             `json:"creator_id" example:"user-123"`
	Parent      *ParentRefResponse `json:"parent,omitempty"`
	Polygon     []PointResponse    `json:"polygon"`
	Tag         TagValueResponse   `json:"tag"`
	HasChildren bool               `json:"has_children" example:"false"`
	ChildCount  *int64             `json:"child_count,omitempty" example:"0"`
	CreatedAt   time.Time          `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   time.Time          `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Deleted     bool               `json:"deleted" example:"false"`
}

func NewAnnotationResponse(a *model.Annotation) *AnnotationResponse {
	var parent *ParentRefResponse
	if a.Parent != nil {
		parent = &ParentRefResponse{
			ID:   a.Parent.ID,
			Type: a.Parent.Type.String(),
		}
	}

	tag := TagValueResponse{
		TagType: a.TagType.String(),
		TagName: a.TagName,
		Value:   a.Value,
		Color:   a.Color,
		Global:  a.Global,
	}

	var polygon []PointResponse
	if a.Polygon != nil {
		polygon = NewPointResponse(*a.Polygon)
	}

	return &AnnotationResponse{
		ID:         a.ID,
		EntityType: a.EntityType.String(),
		Name:       a.Name,
		CreatorID:  a.CreatorID,
		Parent:     parent,
		Polygon:    polygon,
		Tag:        tag,
		CreatedAt:  a.CreatedAt,
		UpdatedAt:  a.UpdatedAt,
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
	}

	return &ListResponse[*AnnotationResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}

// Added DTOs for swagger responses. Swagger requires a concrete type for response schemas.
type AnnotationDataResponse struct {
	Data AnnotationResponse `json:"data"`
}

type AnnotationListResponse struct {
	Data       []AnnotationResponse `json:"data"`
	Pagination *PaginationResponse  `json:"pagination,omitempty"`
}
