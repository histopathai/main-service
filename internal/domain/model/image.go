package model

import (
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

	// Content references (IDs)
	OriginContentID    *string
	DziContentID       *string
	ThumbnailContentID *string
	IndexmapContentID  *string
	TilesContentID     *string
	ZipTilesContentID  *string

	// Processing state
	Processing *vobj.ProcessingInfo
}
