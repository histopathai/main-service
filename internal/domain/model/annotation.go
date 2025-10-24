package model

type Point struct {
	X float64 `firestore:"x"`
	Y float64 `firestore:"y"`
}

type Annotation struct {
	ID               string
	ImageID          string
	AnnotatorID      string
	AnnotationTypeID string
	Polygon          []Point
	Score            *float64
	Class            *string
	CreatedAt        string
	UpdatedAt        string
}
