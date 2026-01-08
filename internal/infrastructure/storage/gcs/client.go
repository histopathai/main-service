package gcs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/storage"
	"github.com/histopathai/main-service/internal/port"
	"google.golang.org/api/iterator"
)

type GCSStorage struct {
	client *storage.Client
}

func NewGCSStorage(ctx context.Context) (port.Storage, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSStorage{
		client: client,
	}, nil
}

func (g *GCSStorage) GenerateSignedURL(
	ctx context.Context,
	bucketName, key string,
	opts port.SignedURLOptions,
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

	bucket := g.client.Bucket(bucketName)
	url, err := bucket.SignedURL(key, signedOpts)
	if err != nil {
		return "", fmt.Errorf("%w: %v", port.ErrSignedURLFailed, err)
	}

	return url, nil
}

func (g *GCSStorage) Exists(ctx context.Context, bucketName, key string) (bool, error) {
	if key == "" {
		return false, port.ErrEmptyKey
	}

	bucket := g.client.Bucket(bucketName)
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

func (g *GCSStorage) Delete(ctx context.Context, bucketName, key string) error {
	if key == "" {
		return port.ErrEmptyKey
	}

	bucket := g.client.Bucket(bucketName)
	if err := bucket.Object(key).Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil
		}
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	return nil
}

func (g *GCSStorage) DeleteByPrefix(ctx context.Context, bucketName, prefix string) error {
	bucket := g.client.Bucket(bucketName)
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		_ = bucket.Object(attrs.Name).Delete(ctx)
	}
	return nil
}

func (g *GCSStorage) Close() error {
	return g.client.Close()
}
