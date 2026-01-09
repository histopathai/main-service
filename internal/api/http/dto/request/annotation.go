package request

type PointRequest struct {
	X float64 `json:"x" binding:"required" example:"100.5"`
	Y float64 `json:"y" binding:"required" example:"200.3"`
}

// TagValueRequest - swagger compatible version
type TagValueRequest struct {
	TagType string `json:"tag_type" binding:"required,oneof=NUMBER TEXT BOOLEAN SELECT MULTI_SELECT" example:"NUMBER"`
	TagName string `json:"tag_name" binding:"required" example:"Grade"`
	// swaggertype: string - This tells swagger to treat it as string in docs
	Value  interface{} `json:"value" binding:"required" swaggertype:"string" example:"3.5"`
	Color  *string     `json:"color,omitempty" example:"#FF0000"`
	Global bool        `json:"global" example:"false"`
}

type CreateAnnotationRequest struct {
	Parent  ParentRefRequest `json:"parent" binding:"required"`
	Polygon *[]PointRequest  `json:"polygon" binding:"omitempty,min=3,dive"`
	Tag     TagValueRequest  `json:"tag" binding:"required"`
}

type UpdateAnnotationRequest struct {
	Parent  *ParentRefRequest `json:"parent,omitempty" binding:"omitempty"`
	Polygon *[]PointRequest   `json:"polygon,omitempty" binding:"omitempty,min=3,dive"`
	Tag     *TagValueRequest  `json:"tag,omitempty"`
}

type ListAnnotationRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}

var ValidAnnotationSortFields = map[string]bool{
	"created_at": true,
	"updated_at": true,
}
