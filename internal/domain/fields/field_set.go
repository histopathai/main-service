// internal/domain/fields/fieldset.go
package fields

// ============================================================================
// WorkspaceFieldSet
// ============================================================================

type WorkspaceFieldSet struct{}

func NewWorkspaceFieldSet() *WorkspaceFieldSet {
	return &WorkspaceFieldSet{}
}

func (fs *WorkspaceFieldSet) IsValidField(field string) bool {
	// Check entity fields
	if EntityField(field).IsValid() {
		return true
	}

	// Check workspace-specific fields
	if WorkspaceField(field).IsValid() {
		return true
	}

	return false
}

func (fs *WorkspaceFieldSet) GetAllFields() []string {
	fields := make([]string, 0, len(EntityFields)+len(WorkspaceFields))

	// Add entity fields
	for _, f := range EntityFields {
		fields = append(fields, f.APIName())
	}

	// Add workspace fields
	for _, f := range WorkspaceFields {
		fields = append(fields, f.APIName())
	}

	return fields
}

// ============================================================================
// AnnotationFieldSet
// ============================================================================

type AnnotationFieldSet struct{}

func NewAnnotationFieldSet() *AnnotationFieldSet {
	return &AnnotationFieldSet{}
}

func (fs *AnnotationFieldSet) IsValidField(field string) bool {
	if EntityField(field).IsValid() {
		return true
	}

	if AnnotationField(field).IsValid() {
		return true
	}

	return false
}

func (fs *AnnotationFieldSet) GetAllFields() []string {
	fields := make([]string, 0, len(EntityFields)+len(AnnotationFields))

	for _, f := range EntityFields {
		fields = append(fields, f.APIName())
	}

	for _, f := range AnnotationFields {
		fields = append(fields, f.APIName())
	}

	return fields
}

// ============================================================================
// PatientFieldSet
// ============================================================================

type PatientFieldSet struct{}

func NewPatientFieldSet() *PatientFieldSet {
	return &PatientFieldSet{}
}

func (fs *PatientFieldSet) IsValidField(field string) bool {
	return EntityField(field).IsValid() || PatientField(field).IsValid()
}

func (fs *PatientFieldSet) GetAllFields() []string {
	fields := make([]string, 0, len(EntityFields)+len(PatientFields))

	for _, f := range EntityFields {
		fields = append(fields, f.APIName())
	}

	for _, f := range PatientFields {
		fields = append(fields, f.APIName())
	}

	return fields
}

// ============================================================================
// ImageFieldSet
// ============================================================================

type ImageFieldSet struct{}

func NewImageFieldSet() *ImageFieldSet {
	return &ImageFieldSet{}
}

func (fs *ImageFieldSet) IsValidField(field string) bool {
	return EntityField(field).IsValid() || ImageField(field).IsValid()
}

func (fs *ImageFieldSet) GetAllFields() []string {
	fields := make([]string, 0, len(EntityFields)+len(ImageFields))

	for _, f := range EntityFields {
		fields = append(fields, f.APIName())
	}

	for _, f := range ImageFields {
		fields = append(fields, f.APIName())
	}

	return fields
}

// ============================================================================
// AnnotationTypeFieldSet
// ============================================================================

type AnnotationTypeFieldSet struct{}

func NewAnnotationTypeFieldSet() *AnnotationTypeFieldSet {
	return &AnnotationTypeFieldSet{}
}

func (fs *AnnotationTypeFieldSet) IsValidField(field string) bool {
	return EntityField(field).IsValid() || AnnotationTypeField(field).IsValid()
}

func (fs *AnnotationTypeFieldSet) GetAllFields() []string {
	fields := make([]string, 0, len(EntityFields)+len(AnnotationTypeFields))

	for _, f := range EntityFields {
		fields = append(fields, f.APIName())
	}

	for _, f := range AnnotationTypeFields {
		fields = append(fields, f.APIName())
	}

	return fields
}

// ============================================================================
// ContentFieldSet
// ============================================================================

type ContentFieldSet struct{}

func NewContentFieldSet() *ContentFieldSet {
	return &ContentFieldSet{}
}

func (fs *ContentFieldSet) IsValidField(field string) bool {
	return EntityField(field).IsValid() || ContentField(field).IsValid()
}

func (fs *ContentFieldSet) GetAllFields() []string {
	fields := make([]string, 0, len(EntityFields)+len(ContentFields))

	for _, f := range EntityFields {
		fields = append(fields, f.APIName())
	}

	for _, f := range ContentFields {
		fields = append(fields, f.APIName())
	}

	return fields
}
