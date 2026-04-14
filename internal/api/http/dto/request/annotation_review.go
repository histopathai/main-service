package request

type CreateAnnotationReviewRequest struct {
	AnnotationID    string          `json:"annotation_id"     binding:"required"                      example:"anno-123"`
	Status          string          `json:"status"            binding:"required,oneof=approved rejected modified" example:"approved"`
	Comments        *string         `json:"comments,omitempty"                                         example:"Looks correct"`
	ModifiedValue   interface{}     `json:"modified_value,omitempty"  swaggertype:"string"             example:"5.0"`
	ModifiedPolygon *[]PointRequest `json:"modified_polygon,omitempty" binding:"omitempty,min=3,dive"`
}
