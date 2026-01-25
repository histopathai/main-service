// internal/domain/fields/mapper.go
package fields

// MapToFirestore - API field name → Firestore field name
func MapToFirestore(apiFieldName string) string {
	// Try entity field
	if ef := EntityField(apiFieldName); ef.IsValid() {
		return ef.FirestoreName()
	}

	// Try workspace field
	if wf := WorkspaceField(apiFieldName); wf.IsValid() {
		return wf.FirestoreName()
	}

	// Try patient field
	if pf := PatientField(apiFieldName); pf.IsValid() {
		return pf.FirestoreName()
	}

	// Try image field
	if imf := ImageField(apiFieldName); imf.IsValid() {
		return imf.FirestoreName()
	}

	// Try annotation field
	if af := AnnotationField(apiFieldName); af.IsValid() {
		return af.FirestoreName()
	}

	// Try annotation type field
	if atf := AnnotationTypeField(apiFieldName); atf.IsValid() {
		return atf.FirestoreName()
	}

	// Try content field
	if cf := ContentField(apiFieldName); cf.IsValid() {
		return cf.FirestoreName()
	}

	// Fallback: return as-is
	return apiFieldName
}

// MapToDomain - API field name → Domain constant name
func MapToDomain(apiFieldName string) string {
	// Try entity field
	if ef := EntityField(apiFieldName); ef.IsValid() {
		return ef.DomainName()
	}

	// Try workspace field
	if wf := WorkspaceField(apiFieldName); wf.IsValid() {
		return wf.DomainName()
	}

	// Try patient field
	if pf := PatientField(apiFieldName); pf.IsValid() {
		return pf.DomainName()
	}

	// Try image field
	if imf := ImageField(apiFieldName); imf.IsValid() {
		return imf.DomainName()
	}

	// Try annotation field
	if af := AnnotationField(apiFieldName); af.IsValid() {
		return af.DomainName()
	}

	// Try annotation type field
	if atf := AnnotationTypeField(apiFieldName); atf.IsValid() {
		return atf.DomainName()
	}

	// Try content field
	if cf := ContentField(apiFieldName); cf.IsValid() {
		return cf.DomainName()
	}

	// Fallback: return as-is
	return apiFieldName
}
