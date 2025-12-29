package storage

import "errors"

var (
	// Object-related errors
	ErrObjectNotFound      = errors.New("object not found")
	ErrObjectAlreadyExists = errors.New("object already exists")
	ErrInvalidKey          = errors.New("invalid object key")

	// Operation errors
	ErrSignedURLFailed    = errors.New("failed to generate signed URL")
	ErrInvalidMethod      = errors.New("invalid HTTP method")
	ErrInvalidContentType = errors.New("invalid content type")
	ErrExpiredURL         = errors.New("signed URL has expired")

	// Storage provider errors
	ErrStorageUnavailable = errors.New("storage service unavailable")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrQuotaExceeded      = errors.New("storage quota exceeded")

	// Validation errors
	ErrInvalidExpiration = errors.New("invalid expiration duration")
	ErrEmptyKey          = errors.New("object key cannot be empty")
)
