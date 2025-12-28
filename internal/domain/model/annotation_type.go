package model

type AnnotationType struct {
	BaseEntity
	Description           *string
	ScoreEnabled          bool
	ScoreName             *string
	ScoreRange            *[2]float64
	ClassificationEnabled bool
	ClassList             []string
}
