package request

// Annotation Type DTOs
type CreateAnnotationTypeRequest struct {
	Name                  string    `json:"name" binding:"required" example:"Tumor"`
	Description           *string   `json:"description,omitempty" example:"Annotation type for tumor regions."`
	ScoreEnabled          bool      `json:"score_enabled" example:"true"`
	ScoreName             *string   `json:"score_name,omitempty" example:"Tumor Grade"`
	ScoreMin              *float64  `json:"score_min,omitempty" example:"1.0"`
	ScoreMax              *float64  `json:"score_max,omitempty" example:"5.0"`
	ClassificationEnabled bool      `json:"classification_enabled" example:"true"`
	ClassList             *[]string `json:"class_list,omitempty" example:"[\"Benign\", \"Malignant\"]"`
}

type UpdateAnnotationTypeRequest struct {
	Name        *string `json:"name,omitempty" example:"Tumor"`
	Description *string `json:"description,omitempty" example:"Annotation type for tumor regions."`
}

type ListAnnotationTypeRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}
