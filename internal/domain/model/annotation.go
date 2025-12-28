package model

type Point struct {
	X float64
	Y float64
}

type Annotation struct {
	BaseEntity
	ImageID     string
	AnnotatorID string
	Polygon     []Point
	Score       *float64
	Class       *string
	Description *string
}
