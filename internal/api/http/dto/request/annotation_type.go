package request

type TagRequest struct {
	Name     string   `json:"name" binding:"required" example:"Grade"`
	Type     string   `json:"type" binding:"required,oneof=NUMBER TEXT BOOLEAN SELECT MULTI_SELECT" example:"NUMBER"`
	Options  []string `json:"options,omitempty" example:"[\"Option1\", \"Option2\"]"`
	Global   bool     `json:"global" example:"false"`
	Required bool     `json:"required" example:"true"`
	Min      *float64 `json:"min,omitempty" example:"1.0"`
	Max      *float64 `json:"max,omitempty" example:"5.0"`
	Color    *string  `json:"color,omitempty" example:"#FF0000"`
}

type CreateAnnotationTypeRequest struct {
	Name string     `json:"name" binding:"required" example:"Tumor"`
	Tag  TagRequest `json:"tag" binding:"required"`
}

type UpdateAnnotationTypeRequest struct {
	Name *string     `json:"name,omitempty" example:"Tumor"`
	Tag  *TagRequest `json:"tag,omitempty" binding:"omitempty"`
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
