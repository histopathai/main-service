package model

import "time"

type AnnotationType struct {
	ID                    string
	Name                  string
	Description           *string
	ScoreEnabled          bool
	ScoreName             *string
	ScoreRange            *[2]float64
	ClassificationEnabled bool
	ClassList             *[]string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

func (at AnnotationType) GetID() string {
	return at.ID
}

func (at *AnnotationType) SetID(id string) {
	at.ID = id
}

func (at *AnnotationType) SetCreatedAt(t time.Time) {
	at.CreatedAt = t
}

func (at *AnnotationType) SetUpdatedAt(t time.Time) {
	at.UpdatedAt = t
}
