package model

type Workspace struct {
	BaseEntity
	AnnotationTypeID *string
	OrganType        string
	Organization     string
	Description      string
	License          string
	ResourceURL      *string
	ReleaseYear      *int
}
