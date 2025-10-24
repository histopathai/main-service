package model

import "time"

type AnnotationType struct {
	ID                    string
	Name                  string
	Desc                  *string
	ScoreEnabled          bool
	ScoreName             *string
	ScoreDesc             *string
	ScoreRange            *[2]float64
	ClassificationEnabled bool
	ClassList             *[]string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}
