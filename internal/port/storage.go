package port

import (
	"context"
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
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

type PresignedURLPayload struct {
	URL       string            `json:"url"`
	Method    SignedURLMethod   `json:"method"`
	ExpiresAt time.Time         `json:"expires_at"`
	Headers   map[string]string `json:"headers,omitempty"`
}

type Storage interface {
	Provider() vobj.ContentProvider

	Exists(ctx context.Context, content model.Content) (bool, error)

	GetAttributes(ctx context.Context, content model.Content) (*FileAttributes, error)

	// metadata parametresi eklendi
	GenerateSignedURL(ctx context.Context, method SignedURLMethod, content model.Content, expiry time.Duration) (*PresignedURLPayload, error)
}

type FileAttributes struct {
	Size        int64
	ContentType string
	UpdatedAt   time.Time
}
