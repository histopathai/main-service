package port

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

type Storage interface {
	GenerateSignedURL(ctx context.Context, key string, opts vobj.SignedURLOptions) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
}
