package model

import (
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type Patient struct {
	vobj.Entity
	Age     *int
	Gender  *string
	Race    *string
	Disease *string
	Subtype *string
	Grade   *int
	History *string
}
