package model

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type AnnotationType struct {
	vobj.Entity
	Tag vobj.Tag
}
