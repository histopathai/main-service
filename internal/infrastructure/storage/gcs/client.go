package gcs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/storage"
	"github.com/histopathai/main-service-refactor/internal/domain/model"
	domainStorage "github.com/histopathai/main-service-refactor/internal/domain/storage"
	"google.golang.org/api/iterator"
)

type GCSClient struct {
	client *storage.Client
	logger *slog.Logger
}

func NewGCSClient(ctx context.Context, logger *slog.Logger) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", mapGCSError(err))
	}

	return &GCSClient{
		client: client,
		logger: logger,
	}, nil
}

func (g *GCSClient) GenerateSignedURL(
	ctx context.Context,
	bucketName string,
	method domainStorage.SignedURLMethod,
	image *model.Image, contentType string,
	expiration time.Duration,
) (string, error) {

	opts := &storage.SignedURLOptions{
		Method:  string(method),
		Expires: time.Now().Add(expiration),
	}

	if method == domainStorage.MethodPut && contentType != "" {
		opts.ContentType = contentType
	}
	opts.Headers = []string{
		"x-goog-meta-image-id:" + image.ID,
		"x-goog-meta-patient-id:" + image.PatientID,
		"x-goog-meta-creator-id:" + image.CreatorID,
		"x-goog-meta-format:" + image.Format,
		"x-goog-meta-file-name:" + image.FileName,
		"x-goog-meta-origin-path:" + image.OriginPath,
		"x-goog-meta-status:" + string(image.Status),
	}
	if image.Width != nil {
		opts.Headers = append(opts.Headers, fmt.Sprintf("x-goog-meta-width:%d", *image.Width))
	}
	if image.Height != nil {
		opts.Headers = append(opts.Headers, fmt.Sprintf("x-goog-meta-height:%d", *image.Height))
	}
	if image.Size != nil {
		opts.Headers = append(opts.Headers, fmt.Sprintf("x-goog-meta-size:%d", *image.Size))
	}

	metadata := &storage.ObjectAttrs{
		Name:        image.OriginPath,
		ContentType: contentType,
		Metadata: map[string]string{
			"image-id":    image.ID,
			"patient-id":  image.PatientID,
			"creator-id":  image.CreatorID,
			"format":      image.Format,
			"file-name":   image.FileName,
			"origin-path": image.OriginPath,
			"status":      string(image.Status),
		},
	}
	if image.Width != nil {
		metadata.Metadata["width"] = fmt.Sprintf("%d", *image.Width)
	}
	if image.Height != nil {
		metadata.Metadata["height"] = fmt.Sprintf("%d", *image.Height)
	}
	if image.Size != nil {
		metadata.Metadata["size"] = fmt.Sprintf("%d", *image.Size)
	}

	url, err := g.client.Bucket(bucketName).SignedURL(metadata.Name, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", mapGCSError(err))
	}

	return url, nil
}

func (g *GCSClient) GetObjectMetadata(ctx context.Context,
	bucketName string,
	objectKey string,
) (*domainStorage.ObjectMetadata, error) {
	attrs, err := g.client.Bucket(bucketName).Object(objectKey).Attrs(ctx)

	if err != nil {
		if errors.Is(mapGCSError(err), ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, mapGCSError(err)
	}

	metadata := domainStorage.ObjectMetadata{
		Name:        attrs.Name,
		Size:        attrs.Size,
		ContentType: attrs.ContentType,
		CreatedAt:   &attrs.Created,
		UpdatedAt:   &attrs.Updated,
		MetaData:    attrs.Metadata,
	}

	g.logger.Info("Object metadata retrieved successfully",
		"bucket", bucketName,
		"objectKey", objectKey,
		"size", attrs.Size,
	)
	return &metadata, nil
}

func (g *GCSClient) ObjectExists(ctx context.Context,
	bucketName string,
	objectKey string,
) (bool, error) {
	_, err := g.client.Bucket(bucketName).Object(objectKey).Attrs(ctx)
	mappedErr := mapGCSError(err)
	if errors.Is(mappedErr, ErrNotFound) {
		return false, nil
	}
	if mappedErr != nil {
		return false, mappedErr
	}
	return true, nil
}

func (g *GCSClient) ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error) {

	var objects []string
	it := g.client.Bucket(bucketName).Objects(ctx, &storage.Query{Prefix: prefix})

	for {
		attr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			message := "Failed to list objects"
			g.logger.Error(message,
				"error", err,
				"bucket", bucketName,
				"prefix", prefix,
			)
			return nil, mapGCSError(err)
		}
		objects = append(objects, attr.Name)
	}

	g.logger.Info("Objects listed successfully",
		"bucket", bucketName,
		"prefix", prefix,
		"count", len(objects),
	)
	return objects, nil
}

func (g *GCSClient) Close() error {
	return g.client.Close()
}
