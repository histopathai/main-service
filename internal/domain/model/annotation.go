package model

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type Annotation struct {
	vobj.Entity
	Polygon *[]vobj.Point
	vobj.TagValue
}
