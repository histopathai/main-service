package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type TagResponse struct {
	Name     string   `json:"name" example:"Grade"`
	Type     string   `json:"type" example:"NUMBER"`
	Options  []string `json:"options,omitempty"`
	Global   bool     `json:"global" example:"false"`
	Required bool     `json:"required" example:"true"`
	Min      *float64 `json:"min,omitempty" example:"1.0"`
	Max      *float64 `json:"max,omitempty" example:"5.0"`
	Color    *string  `json:"color,omitempty" example:"#FF0000"`
}

type ParentRefResponse struct {
	ID   string `json:"id" example:"workspace-123"`
	Type string `json:"type" example:"workspace"`
}

type AnnotationTypeResponse struct {
	ID          string             `json:"id"`
	EntityType  string             `json:"entity_type" example:"annotation_type"`
	Name        *string            `json:"name,omitempty"`
	CreatorID   string             `json:"creator_id"`
	Parent      *ParentRefResponse `json:"parent,omitempty"`
	Tags        []TagResponse      `json:"tags"`
	HasChildren bool               `json:"has_children"`
	ChildCount  *int64             `json:"child_count,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	Deleted     bool               `json:"deleted"`
}

func NewAnnotationTypeResponse(at *model.AnnotationType) *AnnotationTypeResponse {
	var parent *ParentRefResponse
	if at.Parent != nil {
		parent = &ParentRefResponse{
			ID:   at.Parent.ID,
			Type: at.Parent.Type.String(),
		}
	}

	tags := make([]TagResponse, len(at.Tags))
	for i, tag := range at.Tags {
		tags[i] = TagResponse{
			Name:     tag.Name,
			Type:     tag.Type.String(),
			Options:  tag.Options,
			Global:   tag.Global,
			Required: tag.Required,
			Min:      tag.Min,
			Max:      tag.Max,
			Color:    tag.Color,
		}
	}

	return &AnnotationTypeResponse{
		ID:          at.ID,
		EntityType:  at.EntityType.String(),
		Name:        at.Name,
		CreatorID:   at.CreatorID,
		Parent:      parent,
		Tags:        tags,
		HasChildren: at.HasChildren,
		ChildCount:  at.ChildCount,
		CreatedAt:   at.CreatedAt,
		UpdatedAt:   at.UpdatedAt,
		Deleted:     at.Deleted,
	}
}

func NewAnnotationTypeListResponse(result *query.Result[*model.AnnotationType]) *ListResponse[AnnotationTypeResponse] {
	data := make([]AnnotationTypeResponse, len(result.Data))
	for i, at := range result.Data {
		dto := NewAnnotationTypeResponse(at)
		data[i] = *dto
	}
	pagination := PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	return &ListResponse[AnnotationTypeResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}

// Added DTOs for swagger responses. Swagger requires a concrete type for response schemas.
type AnnotationTypeDataResponse struct {
	Data AnnotationTypeResponse `json:"data"`
}

type AnnotationTypeListResponse struct {
	Data       []AnnotationTypeResponse `json:"data"`
	Pagination *PaginationResponse      `json:"pagination,omitempty"`
}
