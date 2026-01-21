package port

import (
	"context"
	"time"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

type SignedURLMethod string

func (s SignedURLMethod) String() string {
	return string(s)
}

const (
	MethodGet    SignedURLMethod = "GET"
	MethodPut    SignedURLMethod = "PUT"
	MethodDelete SignedURLMethod = "DELETE"
)

type StoragePayload struct {
	URL       string            `json:"url"`
	Method    SignedURLMethod   `json:"method"`
	ExpiresAt time.Time         `json:"expires_at"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type Storage interface {
	Provider() vobj.ContentProvider

	Exists(ctx context.Context, path string) (bool, error)

	GetAttributes(ctx context.Context, path string) (*FileAttributes, error)

	// metadata parametresi eklendi
	GenerateSignedURL(ctx context.Context, path string, method SignedURLMethod, contentType string, metadata map[string]string, expiry time.Duration) (*StoragePayload, error)
}

type FileAttributes struct {
	Size        int64
	ContentType string
	UpdatedAt   time.Time
}
