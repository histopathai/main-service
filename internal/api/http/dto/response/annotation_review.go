package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
)

type AnnotationReviewResponse struct {
	ID              string          `json:"id"                         example:"review-123"`
	AnnotationID    string          `json:"annotation_id"              example:"anno-123"`
	ReviewerID      string          `json:"reviewer_id"                example:"user-456"`
	Status          string          `json:"status"                     example:"approved"`
	Comments        *string         `json:"comments,omitempty"         example:"Looks correct"`
	ModifiedValue   interface{}     `json:"modified_value,omitempty"   swaggertype:"string" example:"5.0"`
	ModifiedPolygon []PointResponse `json:"modified_polygon,omitempty"`
	ReviewedAt      time.Time       `json:"reviewed_at"                example:"2024-01-01T12:00:00Z"`
	CreatedAt       time.Time       `json:"created_at"                 example:"2024-01-01T12:00:00Z"`
	UpdatedAt       time.Time       `json:"updated_at"                 example:"2024-01-02T12:00:00Z"`
}

func NewAnnotationReviewResponse(r *model.AnnotationReview) *AnnotationReviewResponse {
	var modifiedPolygon []PointResponse
	if r.ModifiedPolygon != nil {
		modifiedPolygon = NewPointResponse(*r.ModifiedPolygon)
	}

	return &AnnotationReviewResponse{
		ID:              r.ID,
		AnnotationID:    r.AnnotationID,
		ReviewerID:      r.ReviewerID,
		Status:          string(r.Status),
		Comments:        r.Comments,
		ModifiedValue:   r.ModifiedValue,
		ModifiedPolygon: modifiedPolygon,
		ReviewedAt:      r.ReviewedAt,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
	}
}

// Swagger docs
type AnnotationReviewDataResponse struct {
	Data AnnotationReviewResponse `json:"data"`
}
