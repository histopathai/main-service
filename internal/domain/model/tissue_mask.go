package model

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

// TissueMask holds the tissue regions of one image. Its ID is the image ID
// (tissue_masks/{image_id}) and its parent is that image. ML reads this
// collection directly from Firestore, so the stored shape is a contract.
type TissueMask struct {
	vobj.Entity

	WsID             string
	Status           vobj.TissueMaskStatus
	AlgorithmVersion string
	Params           vobj.TissueParams
	Polygons         []vobj.TissuePolygon

	PreviewWidth  int
	PreviewHeight int
	Level0Width   int
	Level0Height  int
	// DownsampleX and DownsampleY map preview pixels to level-0 pixels.
	DownsampleX     float64
	DownsampleY     float64
	TissueAreaRatio float64

	EditedBy   *string
	EditedAt   *time.Time
	ApprovedBy *string
	ApprovedAt *time.Time
}
