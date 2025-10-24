package model

import "time"

type Workspace struct {
	ID               string
	CreatorID        string
	AnnotationTypeID *string
	Name             string
	OrganType        string
	Organization     string
	Description      string
	License          string
	ResourceURL      *string
	ReleaseYear      *int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
