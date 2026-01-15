package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeResponse struct {
	ID         string             `json:"id"`
	EntityType string             `json:"entity_type" example:"annotation_type"`
	Name       string             `json:"name,omitempty"`
	CreatorID  string             `json:"creator_id"`
	Parent     *ParentRefResponse `json:"parent,omitempty"`
	TagType    string             `json:"tag_type" example:"NUMBER"`
	IsGlobal   bool               `json:"is_global" example:"false"`
	IsRequired bool               `json:"is_required" example:"true"`
	Options    []string           `json:"options,omitempty"`
	Min        *float64           `json:"min,omitempty" example:"1.0"`
	Max        *float64           `json:"max,omitempty" example:"5.0"`
	Color      *string            `json:"color,omitempty" example:"#FF0000"`
	CreatedAt  time.Time          `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt  time.Time          `json:"updated_at" example:"2024-01-02T12:00:00Z"`
}

func NewAnnotationTypeResponse(at *model.AnnotationType) *AnnotationTypeResponse {
	return &AnnotationTypeResponse{
		ID:         at.ID,
		EntityType: at.EntityType.String(),
		Name:       *at.Entity.Name,
		CreatorID:  at.Entity.CreatorID,
		Parent:     nil,
		TagType:    at.TagType.String(),
		IsGlobal:   at.IsGlobal,
		IsRequired: at.IsRequired,
		Options:    at.Options,
		Min:        at.Min,
		Max:        at.Max,
		Color:      at.Color,
		CreatedAt:  at.Entity.CreatedAt,
		UpdatedAt:  at.Entity.UpdatedAt,
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
