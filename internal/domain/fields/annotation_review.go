package fields

type AnnotationReviewField string

const (
	AnnotationReviewReviewerID AnnotationReviewField = "reviewer_id"
	AnnotationReviewStatus     AnnotationReviewField = "status"
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
	default:
		return ""
	}
}

func (f AnnotationReviewField) IsValid() bool {
	switch f {
	case AnnotationReviewReviewerID, AnnotationReviewStatus:
		return true
	default:
		return false
	}
}

var AnnotationReviewFields = []AnnotationReviewField{
	AnnotationReviewReviewerID, AnnotationReviewStatus,
}
