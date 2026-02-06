package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)

	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	Has(ctx context.Context, key string) (bool, error)

	Delete(ctx context.Context, key string) (bool, error)

	DeletePattern(ctx context.Context, pattern string) (int, error)

	Clear(ctx context.Context) error
	MGet(ctx context.Context, keys []string) (map[string]interface{}, error)

	// MSet stores multiple values at once
	MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error

	// GetStats returns cache statistics (optional, may return nil if not supported)
	GetStats(ctx context.Context) (*Stats, error)
}

type Stats struct {
	Hits      int64 // Number of cache hits
	Misses    int64 // Number of cache misses
	Size      int64 // Current number of items in cache
	Evictions int64 // Number of evicted items (optional)
}

type CacheEntry struct {
	Key       string
	Value     interface{}
	TTL       time.Duration
	ExpiresAt *time.Time // nil means no expiration
}

type TypedCache[T any] struct {
	cache Cache
}

func NewTypedCache[T any](cache Cache) *TypedCache[T] {
	return &TypedCache[T]{cache: cache}
}

func (tc *TypedCache[T]) Get(ctx context.Context, key string) (*T, error) {
	val, err := tc.cache.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}

	typed, ok := val.(T)
	if !ok {
		return nil, ErrTypeMismatch
	}

	return &typed, nil
}

func (tc *TypedCache[T]) Set(ctx context.Context, key string, value T, ttl time.Duration) error {
	return tc.cache.Set(ctx, key, value, ttl)
}

// Common cache errors
var (
	ErrKeyNotFound  = NewCacheError("key not found")
	ErrTypeMismatch = NewCacheError("type mismatch")
	ErrCacheFull    = NewCacheError("cache is full")
	ErrInvalidTTL   = NewCacheError("invalid TTL")
	ErrNotSupported = NewCacheError("operation not supported")
)

type CacheError struct {
	message string
}

func NewCacheError(message string) *CacheError {
	return &CacheError{message: message}
}

func (e *CacheError) Error() string {
	return e.message
}
