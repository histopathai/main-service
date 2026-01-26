package request

type CreateAnnotationTypeRequest struct {
	Name       string   `json:"name" binding:"required" example:"Tumor Grade"`
	TagType    string   `json:"tag_type" binding:"required,oneof=number text boolean select multi_select" example:"number"`
	Options    []string `json:"options,omitempty" example:"['Option1','Option2']"`
	IsGlobal   bool     `json:"is_global" example:"false"`
	IsRequired bool     `json:"is_required" example:"true"`
	Min        *float64 `json:"min,omitempty" example:"1.0"`
	Max        *float64 `json:"max,omitempty" example:"5.0"`
	Color      *string  `json:"color,omitempty" example:"#FF0000"`
}

type UpdateAnnotationTypeRequest struct {
	CreatorID  *string  `json:"creator_id" binding:"required" example:"1"`
	Name       *string  `json:"name,omitempty" example:"Tumor Grade"`
	Options    []string `json:"options,omitempty" example:"['Option1','Option2']"`
	IsGlobal   *bool    `json:"is_global,omitempty" example:"false"`
	IsRequired *bool    `json:"is_required,omitempty" example:"true"`
	Min        *float64 `json:"min,omitempty" example:"1.0"`
	Max        *float64 `json:"max,omitempty" example:"5.0"`
	Color      *string  `json:"color,omitempty" example:"#FF0000"`
}
