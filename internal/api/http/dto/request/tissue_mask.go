package request

type TissuePointRequest struct {
	X float64 `json:"x" example:"1024.5"`
	Y float64 `json:"y" example:"2048.25"`
}

type TissuePolygonRequest struct {
	Exterior []TissuePointRequest   `json:"exterior" binding:"required"`
	Holes    [][]TissuePointRequest `json:"holes"`
}

type TissueParamsRequest struct {
	Method              string  `json:"method" binding:"required" example:"saturation-gray"`
	SaturationThreshold float64 `json:"saturation_threshold" example:"0.05"`
	GrayThreshold       float64 `json:"gray_threshold" example:"0.92"`
	ClosingRadius       int     `json:"closing_radius" example:"3"`
	MinObjectArea       int     `json:"min_object_area" example:"500"`
	MinHoleArea         int     `json:"min_hole_area" example:"500"`
	SimplifyTolerance   float64 `json:"simplify_tolerance" example:"1"`
}

// SaveTissueMaskRequest carries a complete mask in level-0 pixels. Values the
// frontend computed on the preview (sizes, downsample, area ratio) are stored
// as sent.
type SaveTissueMaskRequest struct {
	AlgorithmVersion string                 `json:"algorithm_version" binding:"required" example:"tissue-v1"`
	Params           TissueParamsRequest    `json:"params" binding:"required"`
	Polygons         []TissuePolygonRequest `json:"polygons"`
	PreviewWidth     int                    `json:"preview_width" example:"2048"`
	PreviewHeight    int                    `json:"preview_height" example:"1451"`
	Level0Width      int                    `json:"level0_width" example:"99840"`
	Level0Height     int                    `json:"level0_height" example:"70740"`
	DownsampleX      float64                `json:"downsample_x" example:"48.75"`
	DownsampleY      float64                `json:"downsample_y" example:"48.75"`
	TissueAreaRatio  float64                `json:"tissue_area_ratio" example:"0.31"`
}
