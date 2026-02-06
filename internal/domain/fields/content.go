package fields

type ContentField string

const (
	ContentProvider      ContentField = "provider"
	ContentPath          ContentField = "path"
	ContentType          ContentField = "content_type"
	ContentSize          ContentField = "size"
	ContentUploadPending ContentField = "upload_pending"
)

func (f ContentField) APIName() string {
	return string(f)
}

func (f ContentField) FirestoreName() string {
	return string(f)
}

func (f ContentField) DomainName() string {
	switch f {
	case ContentProvider:
		return "Provider"
	case ContentPath:
		return "Path"
	case ContentType:
		return "ContentType"
	case ContentSize:
		return "Size"
	case ContentUploadPending:
		return "UploadPending"
	default:
		return ""
	}
}

func (f ContentField) IsValid() bool {
	switch f {
	case ContentProvider, ContentPath, ContentType, ContentSize, ContentUploadPending:
		return true
	default:
		return false
	}
}

var ContentFields = []ContentField{
	ContentProvider, ContentPath, ContentType, ContentSize, ContentUploadPending,
}
