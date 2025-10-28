package model

import "time"

type Patient struct {
	ID          string
	WorkspaceID string
	Name        string
	Age         *int
	Gender      *string
	Race        *string
	Disease     *string
	Subtype     *string
	Grade       *int
	History     *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p Patient) GetID() string {
	return p.ID
}

func (p *Patient) SetID(id string) {
	p.ID = id
}

func (p *Patient) SetCreatedAt(t time.Time) {
	p.CreatedAt = t
}

func (p *Patient) SetUpdatedAt(t time.Time) {
	p.UpdatedAt = t
}
