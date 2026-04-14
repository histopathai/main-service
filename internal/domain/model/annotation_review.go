package model

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type AnnotationReview struct {
	vobj.Entity
	AnnotationID    string
	ReviewerID      string
	Status          fields.ReviewStatusField
	Comments        *string
	ReviewedAt      time.Time
	ModifiedPolygon *[]vobj.Point
	ModifiedValue   any
}
