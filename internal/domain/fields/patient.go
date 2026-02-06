package fields

type PatientField string

const (
	PatientAge     PatientField = "age"
	PatientGender  PatientField = "gender"
	PatientRace    PatientField = "race"
	PatientDisease PatientField = "disease"
	PatientSubtype PatientField = "subtype"
	PatientGrade   PatientField = "grade"
	PatientHistory PatientField = "history"
)

func (f PatientField) APIName() string {
	return string(f)
}

func (f PatientField) FirestoreName() string {
	return string(f)
}

func (f PatientField) DomainName() string {
	switch f {
	case PatientAge:
		return "Age"
	case PatientGender:
		return "Gender"
	case PatientRace:
		return "Race"
	case PatientDisease:
		return "Disease"
	case PatientSubtype:
		return "Subtype"
	case PatientGrade:
		return "Grade"
	case PatientHistory:
		return "History"
	default:
		return ""
	}
}

func (f PatientField) IsValid() bool {
	switch f {
	case PatientAge, PatientGender, PatientRace, PatientDisease, PatientSubtype, PatientGrade, PatientHistory:
		return true
	default:
		return false
	}
}

var PatientFields = []PatientField{
	PatientAge, PatientGender, PatientRace, PatientDisease, PatientSubtype, PatientGrade, PatientHistory,
}
