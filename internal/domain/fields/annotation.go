package fields

type AnnotationField string

const (
	AnnotationTypeID   AnnotationField = "annotation_type_id"
	AnnotationWsID     AnnotationField = "ws_id"
	AnnotationTagType  AnnotationField = "tag_type"
	AnnotationTagValue AnnotationField = "value"
	AnnotationPolygon  AnnotationField = "polygon"
	AnnotationIsGlobal AnnotationField = "is_global"
	AnnotationColor    AnnotationField = "color"
	AnnotationResource AnnotationField = "resource"
	AnnotationReviewID AnnotationField = "review_id"
)

func (f AnnotationField) APIName() string {
	return string(f)
}

func (f AnnotationField) FirestoreName() string {
	return string(f)
}

func (f AnnotationField) DomainName() string {
	switch f {
	case AnnotationTypeID:
		return "AnnotationTypeID"
	case AnnotationWsID:
		return "WsID"
	case AnnotationTagType:
		return "TagType"
	case AnnotationTagValue:
		return "Value"
	case AnnotationPolygon:
		return "Polygon"
	case AnnotationIsGlobal:
		return "IsGlobal"
	case AnnotationColor:
		return "Color"
	case AnnotationResource:
		return "Resource"
	case AnnotationReviewID:
		return "ReviewID"
	default:
		return ""
	}
}

func (f AnnotationField) IsValid() bool {
	switch f {
	case AnnotationTagType, AnnotationTagValue, AnnotationPolygon, AnnotationIsGlobal, AnnotationColor, AnnotationTypeID, AnnotationWsID, AnnotationResource, AnnotationReviewID:
		return true
	default:
		return false
	}
}

var AnnotationFields = []AnnotationField{
	AnnotationTagType, AnnotationTagValue, AnnotationPolygon, AnnotationIsGlobal, AnnotationColor, AnnotationTypeID, AnnotationWsID, AnnotationResource, AnnotationReviewID,
}
