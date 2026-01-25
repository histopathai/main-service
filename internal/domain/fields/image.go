// internal/domain/fields/image.go
package fields

type ImageField string

const (
	ImageFormat ImageField = "format"
	ImageWidth  ImageField = "width"
	ImageHeight ImageField = "height"
	ImageWsID   ImageField = "ws_id"

	// Processing fields
	ImageProcessingStatus          ImageField = "processing.status"
	ImageProcessingVersion         ImageField = "processing.version"
	ImageProcessingFailureReason   ImageField = "processing.failure_reason"
	ImageProcessingRetryCount      ImageField = "processing.retry_count"
	ImageProcessingLastProcessedAt ImageField = "processing.last_processed_at"

	// Content IDs
	ImageOriginContentID    ImageField = "origin_content_id"
	ImageThumbnailContentID ImageField = "thumbnail_content_id"
	ImageDziContentID       ImageField = "dzi_content_id"
	ImageIndexmapContentID  ImageField = "indexmap_content_id"
	ImageTilesContentID     ImageField = "tiles_content_id"
	ImageZipTilesContentID  ImageField = "ziptiles_content_id"

	ImageSize          ImageField = "size"
	ImageMagnification ImageField = "magnification"
)

func (f ImageField) APIName() string {
	// API'de nested fields flat
	switch f {
	case ImageProcessingStatus:
		return "status"
	case ImageProcessingVersion:
		return "version"
	case ImageProcessingFailureReason:
		return "failure_reason"
	case ImageProcessingRetryCount:
		return "retry_count"
	case ImageProcessingLastProcessedAt:
		return "last_processed_at"
	default:
		return string(f)
	}
}

func (f ImageField) FirestoreName() string {
	return string(f) // Firestore nested notation kullanıyor
}

func (f ImageField) DomainName() string {
	switch f {
	case ImageFormat:
		return "Format"
	case ImageWidth:
		return "Width"
	case ImageHeight:
		return "Height"
	case ImageWsID:
		return "WsID"
	case ImageProcessingStatus:
		return "ProcessingStatus"
	case ImageProcessingVersion:
		return "ProcessingVersion"
	case ImageProcessingFailureReason:
		return "ProcessingFailureReason"
	case ImageProcessingRetryCount:
		return "ProcessingRetryCount"
	case ImageProcessingLastProcessedAt:
		return "ProcessingLastProcessedAt"
	case ImageOriginContentID:
		return "OriginContentID"
	case ImageThumbnailContentID:
		return "ThumbnailContentID"
	case ImageDziContentID:
		return "DziContentID"
	case ImageIndexmapContentID:
		return "IndexmapContentID"
	case ImageTilesContentID:
		return "TilesContentID"
	case ImageZipTilesContentID:
		return "ZipTilesContentID"
	case ImageSize:
		return "Size"
	case ImageMagnification:
		return "Magnification"
	default:
		return ""
	}
}

func (f ImageField) IsValid() bool {
	switch f {
	case ImageFormat, ImageWidth, ImageHeight, ImageWsID,
		ImageProcessingStatus, ImageProcessingVersion, ImageProcessingFailureReason,
		ImageProcessingRetryCount, ImageProcessingLastProcessedAt,
		ImageOriginContentID, ImageThumbnailContentID, ImageDziContentID,
		ImageIndexmapContentID, ImageTilesContentID, ImageZipTilesContentID,
		ImageSize, ImageMagnification:
		return true
	default:
		return false
	}
}

var ImageFields = []ImageField{
	ImageFormat, ImageWidth, ImageHeight, ImageWsID,
	ImageProcessingStatus, ImageProcessingVersion, ImageProcessingFailureReason,
	ImageProcessingRetryCount, ImageProcessingLastProcessedAt,
	ImageOriginContentID, ImageThumbnailContentID, ImageDziContentID,
	ImageIndexmapContentID, ImageTilesContentID, ImageZipTilesContentID,
	ImageSize, ImageMagnification,
}
