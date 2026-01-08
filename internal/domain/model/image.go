package model

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type Image struct {
	*vobj.Entity
	Format        string
	Width         *int
	Height        *int
	Size          *int64
	OriginPath    string
	ProcessedPath *string
	ProcessReport *vobj.ImageProcessReport
}
