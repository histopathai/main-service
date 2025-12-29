package storage

import (
	"context"
	"time"
)

type SignedURLMethod string

const (
	MethodGet    SignedURLMethod = "GET"    // Download
	MethodPut    SignedURLMethod = "PUT"    // Upload
	MethodPost   SignedURLMethod = "POST"   // Resumable upload
	MethodDelete SignedURLMethod = "DELETE" // Delete
	MethodHead   SignedURLMethod = "HEAD"   // Get metadata
)

type SignedURLOptions struct {
	Method      SignedURLMethod
	ExpiresIn   time.Duration
	ContentType string
	Metadata    map[string]string // Custom metadata for PUT/POST operations
}

type Storage interface {
	GenerateSignedURL(ctx context.Context, key string, opts SignedURLOptions) (string, error)
	Exists(ctx context.Context, key string) (bool, error)
}
