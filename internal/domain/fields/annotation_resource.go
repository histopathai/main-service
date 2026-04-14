package fields

type AnnotationResourceField string

const (
	AnnotationResourceManual   AnnotationResourceField = "manual"
	AnnotationResourceModel    AnnotationResourceField = "model"
	AnnotationResourceImported AnnotationResourceField = "imported"
)

func (f AnnotationResourceField) APIName() string {
	return string(f)
}

func (f AnnotationResourceField) FirestoreName() string {
	return string(f)
}

func (f AnnotationResourceField) DomainName() string {
	switch f {
	case AnnotationResourceManual:
		return "Manual"
	case AnnotationResourceModel:
		return "Model"
	case AnnotationResourceImported:
		return "Imported"
	default:
		return ""
	}
}

func (f AnnotationResourceField) IsValid() bool {
	switch f {
	case AnnotationResourceManual, AnnotationResourceModel, AnnotationResourceImported:
		return true
	default:
		return false
	}
}

var AnnotationResourceFields = []AnnotationResourceField{
	AnnotationResourceManual, AnnotationResourceModel, AnnotationResourceImported,
}
