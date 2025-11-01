package response

import (
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type AnnotationTypeResponse struct {
	ID                    string    `json:"id"`
	CreatorID             string    `json:"creator_id"`
	Name                  string    `json:"name"`
	Description           *string   `json:"description,omitempty"`
	ScoreEnabled          bool      `json:"score_enabled"`
	ClassificationEnabled bool      `json:"classification_enabled"`
	ClassList             []string  `json:"class_list,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func NewAnnotationTypeResponse(at *model.AnnotationType) *AnnotationTypeResponse {
	return &AnnotationTypeResponse{
		ID:                    at.ID,
		CreatorID:             at.CreatorID,
		Name:                  at.Name,
		Description:           at.Description,
		ScoreEnabled:          at.ScoreEnabled,
		ClassificationEnabled: at.ClassificationEnabled,
		ClassList:             at.ClassList,
		CreatedAt:             at.CreatedAt,
		UpdatedAt:             at.UpdatedAt,
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
		Total:   result.Total,
	}

	return &ListResponse[AnnotationTypeResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}
