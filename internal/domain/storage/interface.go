package storage

import (
	"context"
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
)

type SignedURLMethod string

const (
	MethodGet    SignedURLMethod = "GET"
	MethodPut    SignedURLMethod = "PUT"
	MethodDelete SignedURLMethod = "DELETE"
)

type SignedURLPayload struct {
	URL     string
	Headers map[string]string
}

type ObjectStorage interface {
	GenerateSignedURL(ctx context.Context, bucketName string, method SignedURLMethod,
		image *model.Image, contentType string, expiration time.Duration) (*SignedURLPayload, error)
	GetObjectMetadata(ctx context.Context, bucketName string, objectKey string) (*ObjectMetadata, error)
	ObjectExists(ctx context.Context, bucketName string, objectKey string) (bool, error)
	//DeleteObject(ctx context.Context, bucketName string, objectKey string) error // (Uncomment if needed in future)
	ListObjects(ctx context.Context, bucketName string, prefix string) ([]string, error)
}

type ObjectMetadata struct {
	Name        string
	Size        int64
	ContentType string
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	MetaData    map[string]string
}
