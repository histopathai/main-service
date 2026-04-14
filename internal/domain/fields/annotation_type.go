package fields

type AnnotationTypeField string

const (
	AnnotationTypeTagType    AnnotationTypeField = "tag_type"
	AnnotationTypeIsGlobal   AnnotationTypeField = "is_global"
	AnnotationTypeIsRequired AnnotationTypeField = "is_required"
	AnnotationTypeOptions    AnnotationTypeField = "options"
	AnnotationTypeMin        AnnotationTypeField = "min"
	AnnotationTypeMax        AnnotationTypeField = "max"
	AnnotationTypeColor      AnnotationTypeField = "color"
)

func (f AnnotationTypeField) APIName() string {
	return string(f)
}

func (f AnnotationTypeField) FirestoreName() string {
	return string(f)
}

func (f AnnotationTypeField) DomainName() string {
	switch f {
	case AnnotationTypeTagType:
		return "TagType"
	case AnnotationTypeIsGlobal:
		return "IsGlobal"
	case AnnotationTypeIsRequired:
		return "IsRequired"
	case AnnotationTypeOptions:
		return "Options"
	case AnnotationTypeMin:
		return "Min"
	case AnnotationTypeMax:
		return "Max"
	case AnnotationTypeColor:
		return "Color"
	default:
		return ""
	}
}

func (f AnnotationTypeField) IsValid() bool {
	switch f {
	case AnnotationTypeTagType, AnnotationTypeIsGlobal, AnnotationTypeIsRequired, AnnotationTypeOptions, AnnotationTypeMin, AnnotationTypeMax, AnnotationTypeColor:
		return true
	default:
		return false
	}
}

var AnnotationTypeFields = []AnnotationTypeField{
	AnnotationTypeTagType, AnnotationTypeIsGlobal, AnnotationTypeIsRequired, AnnotationTypeOptions, AnnotationTypeMin, AnnotationTypeMax, AnnotationTypeColor,
}
