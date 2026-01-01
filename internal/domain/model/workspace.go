package model

type Workspace struct {
	BaseEntity
	OrganType    string
	Organization string
	Description  string
	License      string
	ResourceURL  *string
	ReleaseYear  *int
}
