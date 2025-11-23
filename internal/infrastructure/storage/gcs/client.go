package gcs

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/histopathai/main-service/internal/domain/model"
	domainStorage "github.com/histopathai/main-service/internal/domain/storage"
	"google.golang.org/api/iterator"
)

type GCSClient struct {
	client *storage.Client
	logger *slog.Logger
}

func NewGCSClient(ctx context.Context, logger *slog.Logger) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, mapGCSError(err, "failed to create GCS client")
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
) (*domainStorage.SignedURLPayload, error) {

	opts := &storage.SignedURLOptions{
		Method:  string(method),
		Expires: time.Now().Add(expiration),
	}

	if method == domainStorage.MethodPut && contentType != "" {
		opts.ContentType = contentType
	}

	headers := []string{
		"x-goog-meta-image-id:" + image.ID,
		"x-goog-meta-patient-id:" + image.PatientID,
		"x-goog-meta-creator-id:" + image.CreatorID,
		"x-goog-meta-format:" + image.Format,
		"x-goog-meta-file-name:" + image.Name,
		"x-goog-meta-origin-path:" + image.OriginPath,
		"x-goog-meta-status:" + string(image.Status),
	}
	if image.Width != nil {
		headers = append(headers, fmt.Sprintf("x-goog-meta-width:%d", *image.Width))
	}
	if image.Height != nil {
		headers = append(headers, fmt.Sprintf("x-goog-meta-height:%d", *image.Height))
	}
	if image.Size != nil {
		headers = append(headers, fmt.Sprintf("x-goog-meta-size:%d", *image.Size))
	}

	opts.Headers = headers

	url, err := g.client.Bucket(bucketName).SignedURL(image.OriginPath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %w", mapGCSError(err, "generating signed URL"))
	}

	headersMap := make(map[string]string, len(headers))
	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) == 2 {
			headersMap[parts[0]] = parts[1]
		}
	}

	payload := &domainStorage.SignedURLPayload{
		URL:     url,
		Headers: headersMap,
	}

	return payload, nil
}

func (g *GCSClient) GetObjectMetadata(ctx context.Context,
	bucketName string,
	objectKey string,
) (*domainStorage.ObjectMetadata, error) {
	attrs, err := g.client.Bucket(bucketName).Object(objectKey).Attrs(ctx)

	if err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			return nil, ErrNotFound
		}
		return nil, mapGCSError(err, "retrieving object metadata")
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
	mappedErr := mapGCSError(err, "checking object existence")
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
			return nil, mapGCSError(err, message)
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

func (g *GCSClient) DeleteObject(ctx context.Context, bucketName string, objectKey string) error {
	obj := g.client.Bucket(bucketName).Object(objectKey)

	if err := obj.Delete(ctx); err != nil {
		if errors.Is(err, storage.ErrObjectNotExist) {
			g.logger.Warn("Object not found, already deleted",
				"bucket", bucketName,
				"objectKey", objectKey,
			)
			return nil // Already deleted, not an error
		}
		g.logger.Error("Failed to delete object",
			"error", err,
			"bucket", bucketName,
			"objectKey", objectKey,
		)
		return mapGCSError(err, "deleting object")
	}

	g.logger.Info("Object deleted successfully",
		"bucket", bucketName,
		"objectKey", objectKey,
	)
	return nil
}

func (g *GCSClient) DeleteObjects(ctx context.Context, bucketName string, objectKeys []string) error {
	if len(objectKeys) == 0 {
		return nil
	}

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errors []error
	)

	// Concurrent deletion with worker pool (max 10 concurrent deletions)
	semaphore := make(chan struct{}, 10)

	for _, key := range objectKeys {
		wg.Add(1)
		go func(objectKey string) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			if err := g.DeleteObject(ctx, bucketName, objectKey); err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to delete %s: %w", objectKey, err))
				mu.Unlock()
			}
		}(key)
	}

	wg.Wait()

	if len(errors) > 0 {
		g.logger.Error("Some objects failed to delete",
			"bucket", bucketName,
			"totalObjects", len(objectKeys),
			"failedCount", len(errors),
		)
		return fmt.Errorf("failed to delete %d/%d objects: %v", len(errors), len(objectKeys), errors[0])
	}

	g.logger.Info("All objects deleted successfully",
		"bucket", bucketName,
		"count", len(objectKeys),
	)
	return nil
}

func (g *GCSClient) DeleteByPrefix(ctx context.Context, bucketName string, prefix string) error {
	if prefix == "" {
		return fmt.Errorf("prefix cannot be empty for safety reasons")
	}

	// List all objects with the prefix
	objects, err := g.ListObjects(ctx, bucketName, prefix)
	if err != nil {
		return fmt.Errorf("failed to list objects with prefix %s: %w", prefix, err)
	}

	if len(objects) == 0 {
		g.logger.Info("No objects found with prefix",
			"bucket", bucketName,
			"prefix", prefix,
		)
		return nil
	}

	g.logger.Info("Deleting objects by prefix",
		"bucket", bucketName,
		"prefix", prefix,
		"count", len(objects),
	)

	// Use batch delete
	return g.DeleteObjects(ctx, bucketName, objects)
}

func (g *GCSClient) DeleteImageFiles(ctx context.Context, bucketName string, originPath string) error {
	var deletionErrors []error

	// Delete origin file
	if err := g.DeleteObject(ctx, bucketName, originPath); err != nil {
		deletionErrors = append(deletionErrors, fmt.Errorf("origin file: %w", err))
	}

	// Delete DZI files (assuming DZI folder has same name without extension)
	// Example: "uuid-image.svs" -> "uuid-image/" or "uuid-image_files/"
	baseName := strings.TrimSuffix(originPath, filepath.Ext(originPath))

	// Common DZI patterns
	dziPrefixes := []string{
		baseName + "/",       // folder structure
		baseName + "_files/", // DZI tiles folder
		baseName + ".dzi",    // DZI descriptor file
	}

	for _, prefix := range dziPrefixes {
		if err := g.DeleteByPrefix(ctx, bucketName, prefix); err != nil {
			g.logger.Warn("Failed to delete DZI files",
				"prefix", prefix,
				"error", err,
			)
			// Don't add to errors, DZI might not exist
		}
	}

	if len(deletionErrors) > 0 {
		return fmt.Errorf("deletion errors occurred: %v", deletionErrors)
	}

	return nil
}
