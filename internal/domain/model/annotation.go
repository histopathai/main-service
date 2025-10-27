package model

import "time"

type Point struct {
	X float64
	Y float64
}

type Annotation struct {
	ID          string
	ImageID     string
	AnnotatorID string
	Polygon     []Point
	Score       *float64
	Class       *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (a Annotation) GetID() string {
	return a.ID
}

func (a *Annotation) SetID(id string) {
	a.ID = id
}

func (a *Annotation) SetCreatedAt(t time.Time) {
	a.CreatedAt = t
}

func (a *Annotation) SetUpdatedAt(t time.Time) {
	a.UpdatedAt = t
}
