package common

import "github.com/histopathai/main-service/internal/domain/port"

type CreateCommand[T port.Entity] interface {
	Validate() error
	ToEntity() (T, error)
}
