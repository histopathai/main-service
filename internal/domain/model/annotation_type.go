package model

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type AnnotationType struct {
	vobj.Entity
	TagType    vobj.TagType
	IsGlobal   bool
	IsRequired bool
	Options    []string
	Min        *float64
	Max        *float64
	Color      *string
}
