package request

// Annotation DTOs
type PointRequest struct {
	X float64 `json:"x" binding:"required" example:"100.5"`
	Y float64 `json:"y" binding:"required" example:"200.3"`
}

type TagValueRequest struct {
	TagType string      `json:"tag_type" binding:"required,oneof=NUMBER TEXT BOOLEAN SELECT MULTI_SELECT" example:"NUMBER"`
	TagName string      `json:"tag_name" binding:"required" example:"Grade"`
	Value   interface{} `json:"value" binding:"required" example:"3.5"`
	Color   *string     `json:"color,omitempty" example:"#FF0000"`
	Global  bool        `json:"global" example:"false"`
}

type CreateAnnotationRequest struct {
	ParentID string          `json:"parent_id" binding:"required" example:"image-123"`
	Polygon  []PointRequest  `json:"polygon" binding:"required,min=3,dive"`
	Tag      TagValueRequest `json:"tag" binding:"required"`
}

type UpdateAnnotationRequest struct {
	Polygon *[]PointRequest  `json:"polygon,omitempty" binding:"omitempty,min=3,dive"`
	Tag     *TagValueRequest `json:"tag,omitempty"`
}

type ListAnnotationRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}

var ValidAnnotationSortFields = map[string]bool{
	"created_at": true,
	"updated_at": true,
	"name":       true,
}
