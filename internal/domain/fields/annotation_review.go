package fields

type AnnotationReviewField string

const (
	AnnotationReviewReviewerID      AnnotationReviewField = "reviewer_id"
	AnnotationReviewStatus          AnnotationReviewField = "status"
	AnnotationReviewComments        AnnotationReviewField = "comments"
	AnnotationReviewModifiedPolygon AnnotationReviewField = "modified_polygon"
	AnnotationReviewModifiedValue   AnnotationReviewField = "modified_value"
	AnnotationReviewReviewedAt      AnnotationReviewField = "reviewed_at"
)

func (f AnnotationReviewField) APIName() string {
	return string(f)
}

func (f AnnotationReviewField) FirestoreName() string {
	return string(f)
}

func (f AnnotationReviewField) DomainName() string {
	switch f {
	case AnnotationReviewReviewerID:
		return "ReviewerID"
	case AnnotationReviewStatus:
		return "Status"
	case AnnotationReviewComments:
		return "Comments"
	case AnnotationReviewModifiedPolygon:
		return "ModifiedPolygon"
	case AnnotationReviewModifiedValue:
		return "ModifiedValue"
	case AnnotationReviewReviewedAt:
		return "ReviewedAt"
	default:
		return ""
	}
}

func (f AnnotationReviewField) IsValid() bool {
	switch f {
	case AnnotationReviewReviewerID, AnnotationReviewStatus, AnnotationReviewComments, AnnotationReviewModifiedPolygon, AnnotationReviewModifiedValue, AnnotationReviewReviewedAt:
		return true
	default:
		return false
	}
}

var AnnotationReviewFields = []AnnotationReviewField{
	AnnotationReviewReviewerID, AnnotationReviewStatus, AnnotationReviewComments, AnnotationReviewModifiedPolygon, AnnotationReviewModifiedValue, AnnotationReviewReviewedAt,
}
