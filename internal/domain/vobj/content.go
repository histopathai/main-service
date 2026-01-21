package vobj

import "errors"

type ContentType string

const (
	// Image types (standard MIME types)
	ContentTypeImageSVS   ContentType = "image/x-aperio-svs"
	ContentTypeImageTIFF  ContentType = "image/tiff"
	ContentTypeImageNDPI  ContentType = "image/x-ndpi"
	ContentTypeImageVMS   ContentType = "image/x-vms"
	ContentTypeImageVMU   ContentType = "image/x-vmu"
	ContentTypeImageSCN   ContentType = "image/x-scn"
	ContentTypeImageMIRAX ContentType = "image/x-mirax"
	ContentTypeImageBIF   ContentType = "image/x-bif"
	ContentTypeImageDNG   ContentType = "image/x-adobe-dng"
	ContentTypeImageBMP   ContentType = "image/bmp"
	ContentTypeImageJPEG  ContentType = "image/jpeg"
	ContentTypeImagePNG   ContentType = "image/png"

	// Archive types
	ContentTypeApplicationZip ContentType = "application/zip"

	// Document types
	ContentTypeApplicationJSON ContentType = "application/json"

	// DZI (Deep Zoom Image) - XML based format
	ContentTypeApplicationDZI ContentType = "application/xml"

	// Generic fallback
	ContentTypeApplicationOctetStream ContentType = "application/octet-stream"
)

type ContentProvider string

const (
	ContentProviderLocal ContentProvider = "local"
	ContentProviderS3    ContentProvider = "s3"
	ContentProviderGCS   ContentProvider = "gcs"
	ContentProviderAzure ContentProvider = "azure"
	ContentProviderMinIO ContentProvider = "minio"
	ContentProviderHTTP  ContentProvider = "http"
)

func (ct ContentType) GetCategory() string {
	switch ct {
	case ContentTypeImageSVS, ContentTypeImageTIFF, ContentTypeImageNDPI,
		ContentTypeImageVMS, ContentTypeImageVMU, ContentTypeImageSCN,
		ContentTypeImageMIRAX, ContentTypeImageBIF, ContentTypeImageDNG,
		ContentTypeImageBMP, ContentTypeImageJPEG, ContentTypeImagePNG:
		return "image"
	case ContentTypeApplicationZip:
		return "archive"
	case ContentTypeApplicationJSON, ContentTypeApplicationDZI:
		return "document"
	default:
		return "other"
	}
}

func (ct ContentType) IsValid() bool {
	switch ct {
	case ContentTypeImageSVS, ContentTypeImageTIFF, ContentTypeImageNDPI,
		ContentTypeImageVMS, ContentTypeImageVMU, ContentTypeImageSCN,
		ContentTypeImageMIRAX, ContentTypeImageBIF, ContentTypeImageDNG,
		ContentTypeImageBMP, ContentTypeImageJPEG, ContentTypeImagePNG,
		ContentTypeApplicationZip, ContentTypeApplicationJSON,
		ContentTypeApplicationDZI, ContentTypeApplicationOctetStream:
		return true
	default:
		return false
	}
}

func NewContentTypeFromString(s string) (ContentType, error) {

	if s == "" {
		return "", errors.New("content type string is empty")
	}
	value := ContentType(s)
	if value.IsValid() {
		return value, nil
	} else {
		return "", errors.New("invalid content type: " + s)
	}
}

func (ct ContentType) String() string {
	return string(ct)
}

func (cp ContentProvider) IsValid() bool {
	switch cp {
	case ContentProviderLocal, ContentProviderS3, ContentProviderGCS,
		ContentProviderAzure, ContentProviderMinIO, ContentProviderHTTP:
		return true
	default:
		return false
	}
}

func (cp ContentProvider) String() string {
	return string(cp)
}

func (cp ContentProvider) IsCloud() bool {
	switch cp {
	case ContentProviderS3, ContentProviderGCS, ContentProviderAzure, ContentProviderMinIO:
		return true
	default:
		return false
	}
}

type Content struct {
	Provider    ContentProvider
	Path        string
	ContentType ContentType
	Size        int64
}

func GetContentTypeFromExtension(ext string) ContentType {
	switch ext {
	case ".svs":
		return ContentTypeImageSVS
	case ".tif", ".tiff":
		return ContentTypeImageTIFF
	case ".ndpi":
		return ContentTypeImageNDPI
	case ".vms":
		return ContentTypeImageVMS
	case ".vmu":
		return ContentTypeImageVMU
	case ".scn":
		return ContentTypeImageSCN
	case ".mrz":
		return ContentTypeImageMIRAX
	case ".bif":
		return ContentTypeImageBIF
	case ".dng":
		return ContentTypeImageDNG
	case ".bmp":
		return ContentTypeImageBMP
	case ".jpg", ".jpeg":
		return ContentTypeImageJPEG
	case ".png":
		return ContentTypeImagePNG
	case ".zip":
		return ContentTypeApplicationZip
	case ".json":
		return ContentTypeApplicationJSON
	case ".dzi":
		return ContentTypeApplicationDZI
	default:
		return ContentTypeApplicationOctetStream
	}
}
