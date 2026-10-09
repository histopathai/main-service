package model

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

type Image struct {
	vobj.Entity

	WsID string

	// Basic image properties
	Format string
	Width  *int
	Height *int

	// WSI-specific optical information
	Magnification *vobj.OpticalMagnification

	// MPP (microns per pixel) and nominal magnification label ("5x", "10x", "20x", "40x", ...),
	// backfilled from ml/eda/mpp_magnification.csv. Read-only: not settable via Upload/Update.
	MPP                *float64
	MagnificationLabel *string

	// Content references (IDs)
	OriginContentID    *string
	DziContentID       *string
	ThumbnailContentID *string
	IndexmapContentID  *string
	TilesContentID     *string
	ZipTilesContentID  *string

	// TissuePreviewContentID is the preview the tissue mask was computed on.
	TissuePreviewContentID *string

	// Processing state
	Processing *vobj.ProcessingInfo

	MarkedAsCompleted bool

	// Unsuitable ("Çalışmaya uygun değil"): the image is out of the study. Its
	// labels are kept; who marked it, when and (optionally) why.
	Unsuitable     bool
	UnsuitableBy   *string
	UnsuitableAt   *time.Time
	UnsuitableNote *string
}
