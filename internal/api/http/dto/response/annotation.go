package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PointResponse struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
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

type TagValueResponse struct {
	TagType string      `json:"tag_type" example:"NUMBER"`
	TagName string      `json:"tag_name" example:"Grade"`
	Value   interface{} `json:"value" example:"3.5"`
	Color   *string     `json:"color,omitempty" example:"#FF0000"`
	Global  bool        `json:"global" example:"false"`
}

type AnnotationResponse struct {
	ID          string             `json:"id"`
	EntityType  string             `json:"entity_type" example:"annotation"`
	Name        *string            `json:"name,omitempty"`
	CreatorID   string             `json:"creator_id"`
	Parent      *ParentRefResponse `json:"parent,omitempty"`
	Polygon     []PointResponse    `json:"polygon"`
	Tag         TagValueResponse   `json:"tag"`
	HasChildren bool               `json:"has_children"`
	ChildCount  *int64             `json:"child_count,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Deleted     bool               `json:"deleted"`
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
		TagType: a.Tag.TagType.String(),
		TagName: a.Tag.TagName,
		Value:   a.Tag.Value,
		Color:   a.Tag.Color,
		Global:  a.Tag.Global,
	}

	return &AnnotationResponse{
		ID:          a.ID,
		EntityType:  a.EntityType.String(),
		Name:        a.Name,
		CreatorID:   a.CreatorID,
		Parent:      parent,
		Polygon:     NewPointResponse(a.Polygon),
		Tag:         tag,
		HasChildren: a.HasChildren,
		ChildCount:  a.ChildCount,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
		Deleted:     a.Deleted,
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
