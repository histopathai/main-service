package model

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type AnnotationType struct {
	vobj.Entity
	Type     vobj.TagType
	Global   bool
	Required bool
	Options  []string
	Min      *float64
	Max      *float64
	Color    *string
}
