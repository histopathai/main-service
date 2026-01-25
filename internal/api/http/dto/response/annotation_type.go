package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeResponse struct {
	ID         string             `json:"id" example:"at-123"`
	EntityType string             `json:"entity_type" example:"annotation_type"`
	CreatorID  string             `json:"creator_id" example:"user-123"`
	Parent     *ParentRefResponse `json:"parent,omitempty"`
	Name       string             `json:"name" example:"Tumor Grade"`
	TagType    string             `json:"tag_type" example:"number"`
	IsGlobal   bool               `json:"is_global" example:"false"`
	IsRequired bool               `json:"is_required" example:"true"`
	Options    []string           `json:"options,omitempty" example:"['Option1','Option2']"`
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
		CreatorID:  at.CreatorID,
		Parent:     NewParentRefResponse(&at.Parent),
		Name:       at.Name,
		TagType:    at.TagType.String(),
		IsGlobal:   at.IsGlobal,
		IsRequired: at.IsRequired,
		Options:    at.Options,
		Min:        at.Min,
		Max:        at.Max,
		Color:      at.Color,
		CreatedAt:  at.CreatedAt,
		UpdatedAt:  at.UpdatedAt,
	}
}

func NewAnnotationTypeListResponse(result *query.Result[*model.AnnotationType]) *ListResponse[AnnotationTypeResponse] {
	data := make([]AnnotationTypeResponse, len(result.Data))
	for i, at := range result.Data {
		data[i] = *NewAnnotationTypeResponse(at)
	}

	return &ListResponse[AnnotationTypeResponse]{
		Data: data,
		Pagination: &PaginationResponse{
			Limit:   result.Limit,
			Offset:  result.Offset,
			HasMore: result.HasMore,
		},
	}
}

// Swagger docs
type AnnotationTypeDataResponse struct {
	Data AnnotationTypeResponse `json:"data"`
}

type AnnotationTypeListResponseDoc struct {
	Data       []AnnotationTypeResponse `json:"data"`
	Pagination *PaginationResponse      `json:"pagination,omitempty"`
}
