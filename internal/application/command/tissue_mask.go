package command

import (
	"math"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

// TissueMaskData is a complete tissue mask as computed on a preview, either by
// the image-processing worker or by the frontend.
type TissueMaskData struct {
	AlgorithmVersion string
	Params           vobj.TissueParams
	Polygons         []vobj.TissuePolygon
	PreviewWidth     int
	PreviewHeight    int
	Level0Width      int
	Level0Height     int
	DownsampleX      float64
	DownsampleY      float64
	TissueAreaRatio  float64
}

func (d *TissueMaskData) Validate() (map[string]interface{}, bool) {
	details := d.Params.Validate()
	if d.AlgorithmVersion == "" {
		details["algorithm_version"] = "algorithm_version is required"
	}
	if d.PreviewWidth <= 0 || d.PreviewHeight <= 0 {
		details["preview_size"] = "preview_width and preview_height must be positive"
	}
	if d.Level0Width <= 0 || d.Level0Height <= 0 {
		details["level0_size"] = "level0_width and level0_height must be positive"
	}
	if !(d.DownsampleX > 0) || !(d.DownsampleY > 0) || math.IsInf(d.DownsampleX, 0) || math.IsInf(d.DownsampleY, 0) {
		details["downsample"] = "downsample_x and downsample_y must be positive"
	}
	if !(d.TissueAreaRatio >= 0 && d.TissueAreaRatio <= 1) {
		details["tissue_area_ratio"] = "tissue_area_ratio must be in [0, 1]"
	}
	for k, v := range vobj.ValidateTissuePolygons(d.Polygons, d.Level0Width, d.Level0Height) {
		details[k] = v
	}
	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

// SaveTissueMaskCommand stores a mask edited in the tissue tab.
type SaveTissueMaskCommand struct {
	TissueMaskData
	ImageID string
	UserID  string
}

func (c *SaveTissueMaskCommand) Validate() (map[string]interface{}, bool) {
	details, _ := c.TissueMaskData.Validate()
	if details == nil {
		details = map[string]interface{}{}
	}
	if c.ImageID == "" {
		details["image_id"] = "image_id is required"
	}
	if c.UserID == "" {
		details["user_id"] = "user_id is required"
	}
	if len(details) > 0 {
		return details, false
	}
	return nil, true
}

// ApplyWorkerTissueMaskCommand stores the default mask computed by the
// image-processing worker (tissue_mask.json).
type ApplyWorkerTissueMaskCommand struct {
	TissueMaskData
	ImageID string
}
