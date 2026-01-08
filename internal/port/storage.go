package port

import (
	"context"

	"errors"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

var (
	ErrEmptyKey           = errors.New("empty key provided")
	ErrInvalidExpiration  = errors.New("invalid expiration duration")
	ErrSignedURLFailed    = errors.New("failed to generate signed URL")
	ErrStorageUnavailable = errors.New("storage service unavailable")
	ErrObjectNotFound     = errors.New("object not found")
)

type Storage interface {
	GenerateSignedURL(ctx context.Context, key string, opts vobj.SignedURLOptions) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
}
