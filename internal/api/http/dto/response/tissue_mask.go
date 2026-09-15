package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type TissuePointResponse struct {
	X float64 `json:"x" example:"1024.5"`
	Y float64 `json:"y" example:"2048.25"`
}

type TissuePolygonResponse struct {
	Exterior []TissuePointResponse   `json:"exterior"`
	Holes    [][]TissuePointResponse `json:"holes"`
}

type TissueParamsResponse struct {
	Method              string  `json:"method" example:"saturation-gray"`
	SaturationThreshold float64 `json:"saturation_threshold" example:"0.05"`
	GrayThreshold       float64 `json:"gray_threshold" example:"0.92"`
	ClosingRadius       int     `json:"closing_radius" example:"3"`
	MinObjectArea       int     `json:"min_object_area" example:"500"`
	MinHoleArea         int     `json:"min_hole_area" example:"500"`
	SimplifyTolerance   float64 `json:"simplify_tolerance" example:"1"`
}

// TissueMaskSummaryResponse is a mask without its polygons, for lists.
type TissueMaskSummaryResponse struct {
	ImageID          string     `json:"image_id" example:"img-123"`
	WsID             string     `json:"ws_id" example:"ws-123"`
	Status           string     `json:"status" example:"auto"`
	AlgorithmVersion string     `json:"algorithm_version" example:"tissue-v1"`
	TissueAreaRatio  float64    `json:"tissue_area_ratio" example:"0.31"`
	PolygonCount     int        `json:"polygon_count" example:"3"`
	PointCount       int        `json:"point_count" example:"2066"`
	EditedBy         *string    `json:"edited_by,omitempty"`
	EditedAt         *time.Time `json:"edited_at,omitempty"`
	ApprovedBy       *string    `json:"approved_by,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type TissueMaskResponse struct {
	TissueMaskSummaryResponse
	Params        TissueParamsResponse    `json:"params"`
	Polygons      []TissuePolygonResponse `json:"polygons"`
	PreviewWidth  int                     `json:"preview_width" example:"2048"`
	PreviewHeight int                     `json:"preview_height" example:"1451"`
	Level0Width   int                     `json:"level0_width" example:"99840"`
	Level0Height  int                     `json:"level0_height" example:"70740"`
	DownsampleX   float64                 `json:"downsample_x" example:"48.75"`
	DownsampleY   float64                 `json:"downsample_y" example:"48.75"`
}

func NewTissueMaskSummaryResponse(m *model.TissueMask) TissueMaskSummaryResponse {
	return TissueMaskSummaryResponse{
		ImageID:          m.ID,
		WsID:             m.WsID,
		Status:           m.Status.String(),
		AlgorithmVersion: m.AlgorithmVersion,
		TissueAreaRatio:  m.TissueAreaRatio,
		PolygonCount:     len(m.Polygons),
		PointCount:       vobj.TissuePointCount(m.Polygons),
		EditedBy:         m.EditedBy,
		EditedAt:         m.EditedAt,
		ApprovedBy:       m.ApprovedBy,
		ApprovedAt:       m.ApprovedAt,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func NewTissueMaskResponse(m *model.TissueMask) *TissueMaskResponse {
	toPoints := func(pts []vobj.Point) []TissuePointResponse {
		out := make([]TissuePointResponse, len(pts))
		for i, p := range pts {
			out[i] = TissuePointResponse{X: p.X, Y: p.Y}
		}
		return out
	}
	polygons := make([]TissuePolygonResponse, len(m.Polygons))
	for i, p := range m.Polygons {
		holes := make([][]TissuePointResponse, len(p.Holes))
		for j, h := range p.Holes {
			holes[j] = toPoints(h)
		}
		polygons[i] = TissuePolygonResponse{Exterior: toPoints(p.Exterior), Holes: holes}
	}
	return &TissueMaskResponse{
		TissueMaskSummaryResponse: NewTissueMaskSummaryResponse(m),
		Params: TissueParamsResponse{
			Method:              m.Params.Method,
			SaturationThreshold: m.Params.SaturationThreshold,
			GrayThreshold:       m.Params.GrayThreshold,
			ClosingRadius:       m.Params.ClosingRadius,
			MinObjectArea:       m.Params.MinObjectArea,
			MinHoleArea:         m.Params.MinHoleArea,
			SimplifyTolerance:   m.Params.SimplifyTolerance,
		},
		Polygons:      polygons,
		PreviewWidth:  m.PreviewWidth,
		PreviewHeight: m.PreviewHeight,
		Level0Width:   m.Level0Width,
		Level0Height:  m.Level0Height,
		DownsampleX:   m.DownsampleX,
		DownsampleY:   m.DownsampleY,
	}
}

// Swagger docs
type TissueMaskDataResponse struct {
	Data TissueMaskResponse `json:"data"`
}

type TissueMaskListResponseDoc struct {
	Data       []TissueMaskSummaryResponse `json:"data"`
	Pagination *PaginationResponse         `json:"pagination"`
}
