package request

import "github.com/histopathai/main-service/internal/domain/model"

// Annotation DTOs
type CreateAnnotationRequest struct {
	ImageID     string        `json:"image_id" binding:"required"`
	AnnotatorID string        `json:"annotator_id" binding:"omitempty"`
	Polygon     []model.Point `json:"polygon" binding:"required,dive"`
	Score       *float64      `json:"score,omitempty"`
	Class       *string       `json:"class,omitempty"`
	Description *string       `json:"description,omitempty"`
}
