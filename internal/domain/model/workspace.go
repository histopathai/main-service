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

func (w Workspace) GetID() string {
	return w.ID
}

func (w *Workspace) SetID(id string) {
	w.ID = id
}

func (w *Workspace) SetCreatedAt(t time.Time) {
	w.CreatedAt = t
}

func (w *Workspace) SetUpdatedAt(t time.Time) {
	w.UpdatedAt = t
}
