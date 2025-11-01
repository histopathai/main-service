package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

type GCSProxyHandler struct {
	gcsClient  *storage.Client
	bucketName string
	logger     *slog.Logger
}

func NewGCSProxyHandler(projectID, bucketName string, logger *slog.Logger) (*GCSProxyHandler, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSProxyHandler{
		gcsClient:  client,
		bucketName: bucketName,
		logger:     logger.WithGroup("gcs_proxy"),
	}, nil
}

func (h *GCSProxyHandler) ProxyObject(c *gin.Context) {
	objectPath := strings.TrimPrefix(c.Param("objectPath"), "/")
	if objectPath == "" {
		h.logger.Warn("GCS Proxy: Empty object path requested")
		c.String(http.StatusBadRequest, "object path is required")
		return
	}

	ctx := c.Request.Context() // Use request context
	rc, err := h.gcsClient.Bucket(h.bucketName).Object(objectPath).NewReader(ctx)
	if err != nil {
		h.logger.Error("GCS Proxy: Object not found", "path", objectPath, "error", err)
		c.String(http.StatusNotFound, fmt.Sprintf("object not found: %s", err.Error()))
		return
	}
	defer rc.Close()

	c.Header("Content-Type", rc.ContentType())
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, rc)
}
