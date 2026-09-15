package handler

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

// maxTissueMaskJSONBytes bounds tissue_mask.json; 40k points take under 2 MiB.
const maxTissueMaskJSONBytes = 16 << 20

type workerPoint struct {
	X float64 `json:"X"`
	Y float64 `json:"Y"`
}

// workerTissueMask mirrors image-processing-service/internal/tissue.Result.
type workerTissueMask struct {
	AlgorithmVersion string `json:"algorithm_version"`
	Params           struct {
		Method              string  `json:"method"`
		SaturationThreshold float64 `json:"saturation_threshold"`
		GrayThreshold       float64 `json:"gray_threshold"`
		ClosingRadius       int     `json:"closing_radius"`
		MinObjectArea       int     `json:"min_object_area"`
		MinHoleArea         int     `json:"min_hole_area"`
		SimplifyTolerance   float64 `json:"simplify_tolerance"`
	} `json:"params"`
	PreviewWidth    int     `json:"preview_width"`
	PreviewHeight   int     `json:"preview_height"`
	Level0Width     int     `json:"level0_width"`
	Level0Height    int     `json:"level0_height"`
	DownsampleX     float64 `json:"downsample_x"`
	DownsampleY     float64 `json:"downsample_y"`
	TissueAreaRatio float64 `json:"tissue_area_ratio"`
	Polygons        []struct {
		Exterior []workerPoint `json:"exterior"`
		Holes    []struct {
			Points []workerPoint `json:"points"`
		} `json:"holes"`
	} `json:"polygons"`
}

func parseWorkerTissueMask(r io.Reader) (command.TissueMaskData, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxTissueMaskJSONBytes+1))
	if err != nil {
		return command.TissueMaskData{}, err
	}
	if len(data) > maxTissueMaskJSONBytes {
		return command.TissueMaskData{}, fmt.Errorf("tissue mask json exceeds %d bytes", maxTissueMaskJSONBytes)
	}
	var w workerTissueMask
	if err := json.Unmarshal(data, &w); err != nil {
		return command.TissueMaskData{}, fmt.Errorf("invalid tissue mask json: %w", err)
	}

	toPoints := func(pts []workerPoint) []vobj.Point {
		out := make([]vobj.Point, len(pts))
		for i, p := range pts {
			out[i] = vobj.Point{X: p.X, Y: p.Y}
		}
		return out
	}
	polygons := make([]vobj.TissuePolygon, len(w.Polygons))
	for i, p := range w.Polygons {
		holes := make([][]vobj.Point, len(p.Holes))
		for j, h := range p.Holes {
			holes[j] = toPoints(h.Points)
		}
		polygons[i] = vobj.TissuePolygon{Exterior: toPoints(p.Exterior), Holes: holes}
	}

	return command.TissueMaskData{
		AlgorithmVersion: w.AlgorithmVersion,
		Params: vobj.TissueParams{
			Method:              w.Params.Method,
			SaturationThreshold: w.Params.SaturationThreshold,
			GrayThreshold:       w.Params.GrayThreshold,
			ClosingRadius:       w.Params.ClosingRadius,
			MinObjectArea:       w.Params.MinObjectArea,
			MinHoleArea:         w.Params.MinHoleArea,
			SimplifyTolerance:   w.Params.SimplifyTolerance,
		},
		Polygons:        polygons,
		PreviewWidth:    w.PreviewWidth,
		PreviewHeight:   w.PreviewHeight,
		Level0Width:     w.Level0Width,
		Level0Height:    w.Level0Height,
		DownsampleX:     w.DownsampleX,
		DownsampleY:     w.DownsampleY,
		TissueAreaRatio: w.TissueAreaRatio,
	}, nil
}
