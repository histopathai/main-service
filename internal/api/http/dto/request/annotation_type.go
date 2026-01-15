package request

type CreateAnnotationTypeRequest struct {
	Name       string   `json:"name" binding:"required" example:"Tumor"`
	TagType    string   `json:"tag_type" binding:"required,oneof=NUMBER TEXT BOOLEAN SELECT MULTI_SELECT" example:"NUMBER"`
	Options    []string `json:"options,omitempty" example:"[\"Option1\", \"Option2\"]"`
	IsGlobal   bool     `json:"is_global" example:"false"`
	IsRequired bool     `json:"is_required" example:"true"`
	Min        *float64 `json:"min,omitempty" example:"1.0"`
	Max        *float64 `json:"max,omitempty" example:"5.0"`
	Color      *string  `json:"color,omitempty" example:"#FF0000"`
}

type UpdateAnnotationTypeRequest struct {
	Name       *string  `json:"name,omitempty" example:"Tumor"`
	Options    []string `json:"options,omitempty" example:"[\"Option1\", \"Option2\"]"`
	IsGlobal   *bool    `json:"is_global,omitempty" example:"false"`
	IsRequired *bool    `json:"is_required,omitempty" example:"true"`
	Min        *float64 `json:"min,omitempty" example:"1.0"`
	Max        *float64 `json:"max,omitempty" example:"5.0"`
	Color      *string  `json:"color,omitempty" example:"#FF0000"`
}

type ListAnnotationTypeRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}

var ValidAnnotationTypeSortFields = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"name":       true,
}
