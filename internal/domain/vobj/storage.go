package vobj

import (
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
