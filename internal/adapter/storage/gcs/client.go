package gcs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/storage"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
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

func (g *GCSAdapter) Provider() vobj.ContentProvider {
	return vobj.ContentProviderGCS
}

func (g *GCSAdapter) Exists(ctx context.Context, content model.Content) (bool, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(content.Path)

	_, err := obj.Attrs(ctx)
	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return false, nil
		}
		return false, mapGCSError(err, "checking object existence in GCS")
	}

	return true, nil
}

func (g *GCSAdapter) GetAttributes(ctx context.Context, content model.Content) (*port.FileAttributes, error) {
	bucket := g.client.Bucket(g.bucketName)
	obj := bucket.Object(content.Path)

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
	method port.SignedURLMethod,
	content model.Content,
	expiry time.Duration) (*port.PresignedURLPayload, error) {

	opts := &storage.SignedURLOptions{
		Method:  method.String(),
		Expires: time.Now().Add(expiry),
		Scheme:  storage.SigningSchemeV4,
	}

	if method == port.MethodPut && content.ContentType.String() != "" {
		opts.ContentType = content.ContentType.String()
	}

	var headers []string
	headersMap := make(map[string]string)

	if method == port.MethodPut {
		// Populate metadata from content
		metadata := make(map[string]string)

		if content.ID != "" {
			metadata["id"] = content.ID
		}
		if content.Name != "" {
			metadata["name"] = content.Name
		}
		if content.CreatorID != "" {
			metadata["creator-id"] = content.CreatorID
		}
		if content.EntityType != "" {
			metadata["entity-type"] = string(content.EntityType)
		}
		if content.Provider != "" {
			metadata["provider"] = string(content.Provider)
		}
		if content.Path != "" {
			metadata["path"] = content.Path
		}
		// Size might not be known yet during upload request sometimes, but if it is:
		if content.Size > 0 {
			metadata["size"] = fmt.Sprintf("%d", content.Size)
		}
		if content.ContentType != "" {
			metadata["content-type"] = string(content.ContentType)
		}

		if len(metadata) > 0 {
			for k, v := range metadata {
				headerKey := fmt.Sprintf("x-goog-meta-%s", k)
				fullHeader := fmt.Sprintf("%s:%s", headerKey, v)

				headers = append(headers, fullHeader)
				headersMap[headerKey] = v
			}
			opts.Headers = headers
		}
	}

	url, err := g.client.Bucket(g.bucketName).SignedURL(content.Path, opts)
	if err != nil {
		return nil, mapGCSError(err, "generating signed URL for GCS object")
	}

	return &port.PresignedURLPayload{
		URL:       url,
		Method:    method,
		ExpiresAt: opts.Expires,
		Headers:   headersMap,
	}, nil
}
