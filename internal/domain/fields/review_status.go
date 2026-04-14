package fields

type ReviewStatusField string

const (
	ReviewStatusPending  ReviewStatusField = "pending"
	ReviewStatusApproved ReviewStatusField = "approved"
	ReviewStatusRejected ReviewStatusField = "rejected"
	ReviewStatusModified ReviewStatusField = "modified"
)

func (f ReviewStatusField) APIName() string {
	return string(f)
}

func (f ReviewStatusField) FirestoreName() string {
	return string(f)
}

func (f ReviewStatusField) DomainName() string {
	switch f {
	case ReviewStatusPending:
		return "Pending"
	case ReviewStatusApproved:
		return "Approved"
	case ReviewStatusRejected:
		return "Rejected"
	case ReviewStatusModified:
		return "Modified"
	default:
		return ""
	}
}

func (f ReviewStatusField) IsValid() bool {
	switch f {
	case ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusModified:
		return true
	default:
		return false
	}
}

var ReviewStatusFields = []ReviewStatusField{
	ReviewStatusPending, ReviewStatusApproved, ReviewStatusRejected, ReviewStatusModified,
}
