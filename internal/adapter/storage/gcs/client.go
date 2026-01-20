package gcs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/storage"
	"github.com/histopathai/main-service/internal/port"
)

type GCSAdapter struct {
	client     *storage.Client
	bucketName string
	logger     *slog.Logger
}

var _ port.Storage = (*GCSAdapter)(nil)

func NewGCSAdapter(client *storage.Client, bucketName string, logger *slog.Logger) *GCSAdapter {
	return &GCSAdapter{
		client:     client,
		bucketName: bucketName,
		logger:     logger,
	}
}

func (g *GCSAdapter) Provider() port.StorageProvider {
	return port.ProviderGCS
}

func (g *GCSAdapter) Exists(ctx context.Context, path string) (bool, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(path)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}
		return false, mapGCSError(err, "checking object existence in GCS")
	}

	return true, nil
}

func (g *GCSAdapter) GetAttributes(ctx context.Context, path string) (*port.FileAttributes, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(path)

	attrs, err := obj.Attrs(ctx)
	if err != nil {
		return nil, mapGCSError(err, "getting object attributes from GCS")
	}

	return &port.FileAttributes{
		Size:        attrs.Size,
		ContentType: attrs.ContentType,
		UpdatedAt:   attrs.Updated,
	}, nil
}

func (g *GCSAdapter) GenerateSignedURL(ctx context.Context,
	path string,
	method port.SignedURLMethod,
	contentType string,
	metadata map[string]string,
	expiry time.Duration) (*port.StoragePayload, error) {

	opts := &storage.SignedURLOptions{
		Method:  method.String(),
		Expires: time.Now().Add(expiry),
		Scheme:  storage.SigningSchemeV4,
	}

	if method == port.MethodPut && contentType != "" {
		opts.ContentType = contentType
	}

	var headers []string
	headersMap := make(map[string]string)

	if method == port.MethodPut && len(metadata) > 0 {
		for k, v := range metadata {
			headerKey := fmt.Sprintf("x-goog-meta-%s", k)
			fullHeader := fmt.Sprintf("%s:%s", headerKey, v)

			headers = append(headers, fullHeader)

			headersMap[headerKey] = v
		}
		opts.Headers = headers
	}

	url, err := g.client.Bucket(g.bucketName).SignedURL(path, opts)
	if err != nil {
		return nil, mapGCSError(err, "generating signed URL for GCS object")
	}

	return &port.StoragePayload{
		URL:       url,
		Method:    method,
		ExpiresAt: opts.Expires,
		Headers:   headersMap,
	}, nil
}
