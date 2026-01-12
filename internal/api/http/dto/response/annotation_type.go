package response

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeResponse struct {
	ID         string             `json:"id"`
	EntityType string             `json:"entity_type" example:"annotation_type"`
	Name       *string            `json:"name,omitempty"`
	CreatorID  string             `json:"creator_id"`
	Parent     *ParentRefResponse `json:"parent,omitempty"`
	Type       string             `json:"type" example:"NUMBER"`
	Global     bool               `json:"global" example:"false"`
	Required   bool               `json:"required" example:"true"`
	Options    []string           `json:"options,omitempty"`
	Min        *float64           `json:"min,omitempty" example:"1.0"`
	Max        *float64           `json:"max,omitempty" example:"5.0"`
	Color      *string            `json:"color,omitempty" example:"#FF0000"`
}

func NewAnnotationTypeResponse(at *model.AnnotationType) *AnnotationTypeResponse {
	return &AnnotationTypeResponse{
		ID:         at.ID,
		EntityType: string(at.EntityType),
		Name:       at.Entity.Name,
		CreatorID:  at.Entity.CreatorID,
		Parent:     nil,
		Type:       at.Type.String(),
		Global:     at.Global,
		Required:   at.Required,
		Options:    at.Options,
		Min:        at.Min,
		Max:        at.Max,
		Color:      at.Color,
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
