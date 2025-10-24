package model

import "time"

type Point struct {
	X float64
	Y float64
}

type Annotation struct {
	ID               string
	ImageID          string
	AnnotatorID      string
	AnnotationTypeID string
	Polygon          []Point
	Score            *float64
	Class            *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
