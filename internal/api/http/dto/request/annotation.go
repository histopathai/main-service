package request

type PointRequest struct {
	X float64 `json:"x" binding:"required" example:"100.5"`
	Y float64 `json:"y" binding:"required" example:"200.3"`
}

type CreateAnnotationRequest struct {
	Parent   ParentRefRequest `json:"parent" binding:"required"`
	WsID     string           `json:"ws_id" binding:"required" example:"ws-123"`
	Name     string           `json:"name" binding:"required" example:"Tumor Region"`
	TagType  string           `json:"tag_type" binding:"required,oneof=number text boolean select multi_select" example:"number"`
	Value    interface{}      `json:"value" binding:"required" swaggertype:"string" example:"3.5"`
	IsGlobal bool             `json:"is_global" example:"false"`
	Color    *string          `json:"color,omitempty" example:"#FF0000"`
	Polygon  *[]PointRequest  `json:"polygon,omitempty" binding:"omitempty,min=3,dive"`
}

type UpdateAnnotationRequest struct {
	CreatorID *string         `json:"creator_id" binding:"required" example:"1"`
	Value     interface{}     `json:"value,omitempty" swaggertype:"string" example:"4.2"`
	Color     *string         `json:"color,omitempty" example:"#00FF00"`
	IsGlobal  *bool           `json:"is_global,omitempty" example:"true"`
	Polygon   *[]PointRequest `json:"polygon,omitempty" binding:"omitempty,min=3,dive"`
}
