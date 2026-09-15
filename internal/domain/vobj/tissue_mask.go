package vobj

import (
	"fmt"
	"math"
)

// TissueMaskStatus is the review state of a tissue mask.
//
//	auto ──(save)──► edited ──(approve)──► approved
//	  └───────────(approve)────────────────┘    │
//	edited ◄──────────(save)────────────────────┘
//
// The image-processing worker only writes masks that are missing or auto.
type TissueMaskStatus string

const (
	TissueMaskStatusAuto     TissueMaskStatus = "auto"
	TissueMaskStatusEdited   TissueMaskStatus = "edited"
	TissueMaskStatusApproved TissueMaskStatus = "approved"
)

func (s TissueMaskStatus) IsValid() bool {
	switch s {
	case TissueMaskStatusAuto, TissueMaskStatusEdited, TissueMaskStatusApproved:
		return true
	default:
		return false
	}
}

func (s TissueMaskStatus) String() string {
	return string(s)
}

// Limits shared with image-processing-service/internal/tissue and the
// frontend port (src/core/tissue).
const (
	TissueMaxPoints            = 40000
	TissueMaxClosingRadius     = 64
	TissueMaxSimplifyTolerance = 16.0
)

// TissueParams are the parameters a mask was computed with. Areas and the
// simplify tolerance are in preview pixels.
type TissueParams struct {
	Method              string
	SaturationThreshold float64
	GrayThreshold       float64
	ClosingRadius       int
	MinObjectArea       int
	MinHoleArea         int
	SimplifyTolerance   float64
}

func (p TissueParams) Validate() map[string]interface{} {
	details := map[string]interface{}{}
	switch p.Method {
	case "saturation-gray", "otsu-saturation", "otsu-gray":
	default:
		details["params.method"] = "must be one of saturation-gray, otsu-saturation, otsu-gray"
	}
	if !inRange(p.SaturationThreshold, 0, 1) {
		details["params.saturation_threshold"] = "must be in [0, 1]"
	}
	if !inRange(p.GrayThreshold, 0, 1) {
		details["params.gray_threshold"] = "must be in [0, 1]"
	}
	if p.ClosingRadius < 0 || p.ClosingRadius > TissueMaxClosingRadius {
		details["params.closing_radius"] = fmt.Sprintf("must be in [0, %d]", TissueMaxClosingRadius)
	}
	if p.MinObjectArea < 0 {
		details["params.min_object_area"] = "must be >= 0"
	}
	if p.MinHoleArea < 0 {
		details["params.min_hole_area"] = "must be >= 0"
	}
	if !inRange(p.SimplifyTolerance, 0, TissueMaxSimplifyTolerance) {
		details["params.simplify_tolerance"] = fmt.Sprintf("must be in [0, %v]", TissueMaxSimplifyTolerance)
	}
	return details
}

// TissuePolygon is one tissue region in level-0 pixels. Rings are implicitly
// closed. Firestore stores holes as {points: [...]} maps because it does not
// allow nested arrays.
type TissuePolygon struct {
	Exterior []Point
	Holes    [][]Point
}

// TissuePointCount returns the number of vertices across all rings.
func TissuePointCount(polygons []TissuePolygon) int {
	n := 0
	for _, p := range polygons {
		n += len(p.Exterior)
		for _, h := range p.Holes {
			n += len(h)
		}
	}
	return n
}

// ValidateTissuePolygons checks ring sizes, finiteness, bounds (when the
// level-0 size is known) and the point budget.
func ValidateTissuePolygons(polygons []TissuePolygon, level0Width, level0Height int) map[string]interface{} {
	details := map[string]interface{}{}
	if n := TissuePointCount(polygons); n > TissueMaxPoints {
		details["polygons"] = fmt.Sprintf("%d points exceed the limit of %d; increase simplify_tolerance", n, TissueMaxPoints)
		return details
	}
	checkRing := func(key string, ring []Point) {
		if len(ring) < 3 {
			details[key] = "ring needs at least 3 points"
			return
		}
		for _, pt := range ring {
			if math.IsNaN(pt.X) || math.IsInf(pt.X, 0) || math.IsNaN(pt.Y) || math.IsInf(pt.Y, 0) {
				details[key] = "coordinates must be finite"
				return
			}
			if level0Width > 0 && level0Height > 0 &&
				(pt.X < 0 || pt.Y < 0 || pt.X > float64(level0Width) || pt.Y > float64(level0Height)) {
				details[key] = fmt.Sprintf("coordinates must lie within the %dx%d image", level0Width, level0Height)
				return
			}
		}
	}
	for i, p := range polygons {
		checkRing(fmt.Sprintf("polygons[%d].exterior", i), p.Exterior)
		for j, h := range p.Holes {
			checkRing(fmt.Sprintf("polygons[%d].holes[%d]", i, j), h)
		}
	}
	return details
}

func inRange(v, lo, hi float64) bool {
	return !math.IsNaN(v) && v >= lo && v <= hi
}
