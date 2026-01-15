package request

type PointRequest struct {
	X float64 `json:"x" binding:"required" example:"100.5"`
	Y float64 `json:"y" binding:"required" example:"200.3"`
}

type CreateAnnotationRequest struct {
	Parent   ParentRefRequest `json:"parent" binding:"required"`
	Name     string           `json:"name" binding:"required" example:"Annotation 1"`
	TagType  string           `json:"tag_type" binding:"required,oneof=NUMBER TEXT BOOLEAN SELECT MULTI_SELECT" example:"NUMBER"`
	Value    interface{}      `json:"value" binding:"required" swaggertype:"string" example:"3.5"`
	IsGlobal bool             `json:"is_global" binding:"required" example:"false"`
	Color    *string          `json:"color,omitempty" example:"#FF0000"`

	Polygon *[]PointRequest `json:"polygon" binding:"omitempty,min=3,dive"`
}

type UpdateAnnotationRequest struct {
	Value    interface{}     `json:"value" binding:"required" swaggertype:"string" example:"4.2"`
	Color    *string         `json:"color,omitempty" example:"#00FF00"`
	IsGlobal *bool           `json:"is_global,omitempty" example:"true"`
	Polygon  *[]PointRequest `json:"polygon" binding:"omitempty,min=3,dive"`
}

type ListAnnotationRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}

var ValidAnnotationSortFields = map[string]bool{
	"created_at": true,
	"updated_at": true,
}
