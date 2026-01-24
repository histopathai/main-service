package vobj

type TagType string

func (t TagType) String() string {
	return string(t)
}

func (t TagType) IsValid() bool {
	switch t {
	case NumberTag, TextTag, BooleanTag, SelectTag, MultiSelectTag:
		return true
	default:
		return false
	}
}
