package gcs

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type GCSStorage struct {
	client     *storage.Client
	bucketName string
}

func NewGCSStorage(ctx context.Context, bucketName string) (port.Storage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (g *GCSStorage) GenerateSignedURL(
	ctx context.Context,
	key string,
	opts vobj.SignedURLOptions,
) (string, error) {
	if key == "" {
		return "", port.ErrEmptyKey
	}

	if opts.ExpiresIn <= 0 {
		return "", port.ErrInvalidExpiration
	}

	signedOpts := &storage.SignedURLOptions{
		Method:  string(opts.Method),
		Expires: time.Now().Add(opts.ExpiresIn),
	}

	if opts.ContentType != "" {
		signedOpts.ContentType = opts.ContentType
	}

	if len(opts.Metadata) > 0 {
		headers := make([]string, 0, len(opts.Metadata))
		for k, v := range opts.Metadata {
			headers = append(headers, fmt.Sprintf("x-goog-meta-%s:%s", k, v))
		}
		signedOpts.Headers = headers
	}

	bucket := g.client.Bucket(g.bucketName)
	url, err := bucket.SignedURL(key, signedOpts)
	if err != nil {
		return "", fmt.Errorf("%w: %v", port.ErrSignedURLFailed, err)
	}

	return url, nil
}

func (g *GCSStorage) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, port.ErrEmptyKey
	}

	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(key)

	_, err := obj.Attrs(ctx)
	if err == storage.ErrObjectNotExist {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("%w: %v", port.ErrStorageUnavailable, err)
	}

	return true, nil
}

func (g *GCSStorage) Close() error {
	return g.client.Close()
}
