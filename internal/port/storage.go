package port

import (
	"context"
	"time"

	"errors"
)

var (
	ErrEmptyKey           = errors.New("empty key provided")
	ErrInvalidExpiration  = errors.New("invalid expiration duration")
	ErrSignedURLFailed    = errors.New("failed to generate signed URL")
	ErrStorageUnavailable = errors.New("storage service unavailable")
	ErrObjectNotFound     = errors.New("object not found")
)

type SignedURLMethod string

const (
	MethodGet  SignedURLMethod = "GET"
	MethodPut  SignedURLMethod = "PUT"
	MethodPost SignedURLMethod = "POST"
	MethodHead SignedURLMethod = "HEAD"
)

type SignedURLOptions struct {
	Method      SignedURLMethod
	ExpiresIn   time.Duration
	ContentType string
	Metadata    map[string]string
}

func NewSignedURLOptions(method SignedURLMethod, expiresIn time.Duration, contentType string, metadata map[string]string) SignedURLOptions {
	return SignedURLOptions{
		Method:      method,
		ExpiresIn:   expiresIn,
		ContentType: contentType,
		Metadata:    metadata,
	}
}

func DefaultSignedURLOptions() SignedURLOptions {
	return SignedURLOptions{
		Method:    MethodGet,
		ExpiresIn: 15 * time.Minute,
	}
}

type Storage interface {
	GenerateSignedURL(ctx context.Context, bucket, key string, opts SignedURLOptions) (string, error)

	Exists(ctx context.Context, bucket, key string) (bool, error)

	Delete(ctx context.Context, bucket, key string) error

	DeleteByPrefix(ctx context.Context, bucket, prefix string) error
}
