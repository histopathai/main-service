package model

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
)

// ProcessedContent represents the artifacts generated from image processing
type ProcessedContent struct {
	DZI       *vobj.Content // image.dzi file
	Tiles     *vobj.Content // v2: tiles.zip, v1: tiles/ directory
	Thumbnail *vobj.Content // thumbnail.jpg
	IndexMap  *vobj.Content // v2 only: indexmap.json (inside zip or separate)
}

type Image struct {
	vobj.Entity

	WsID string

	// Basic image properties
	Format string
	Width  *int
	Height *int

	// WSI-specific optical information
	Magnification *vobj.OpticalMagnification

	// Content references
	OriginContent    *vobj.Content
	ProcessedContent *ProcessedContent

	// Processing state
	Processing *vobj.ProcessingInfo
}

// Status checks
func (i *Image) IsProcessed() bool {
	return i.Processing.Status == vobj.StatusProcessed && i.ProcessedContent != nil
}

func (i *Image) HasOriginContent() bool {
	return i.OriginContent != nil
}

// Upload source detection
func (i *Image) IsWebUpload() bool {
	return i.OriginContent != nil && i.OriginContent.Provider.IsCloud()
}

func (i *Image) IsAdminUpload() bool {
	return i.OriginContent != nil && i.OriginContent.Provider == vobj.ContentProviderLocal
}

// Processing version checks
func (i *Image) GetProcessingVersion() vobj.ProcessingVersion {
	return i.Processing.Version
}

func (i *Image) IsV1Processing() bool {
	return i.Processing.Version == vobj.ProcessingV1
}

func (i *Image) IsV2Processing() bool {
	return i.Processing.Version == vobj.ProcessingV2
}

// Processing content setters
func (i *Image) SetProcessedContentV1(dziContent, tilesDir, thumbnail *vobj.Content) {
	i.ProcessedContent = &ProcessedContent{
		DZI:       dziContent,
		Tiles:     tilesDir,
		Thumbnail: thumbnail,
	}
	i.Processing.MarkAsProcessed(vobj.ProcessingV1)
}

func (i *Image) SetProcessedContentV2(dziContent, tilesZip, thumbnail, indexMap *vobj.Content) {
	i.ProcessedContent = &ProcessedContent{
		DZI:       dziContent,
		Tiles:     tilesZip,
		Thumbnail: thumbnail,
		IndexMap:  indexMap,
	}
	i.Processing.MarkAsProcessed(vobj.ProcessingV2)
}

// Processing state management
func (i *Image) MarkAsProcessing() {
	i.Processing.Status = vobj.StatusProcessing
}

func (i *Image) MarkAsFailed(reason string) {
	i.Processing.MarkAsFailed(reason)
}

func (i *Image) MarkForRetry() {
	i.Processing.MarkForRetry()
}

func (i *Image) MarkAsDeleting() {
	i.Processing.Status = vobj.StatusDeleting
}

func (i *Image) IsRetryable(maxRetries int) bool {
	return i.Processing.IsRetryable(maxRetries)
}
