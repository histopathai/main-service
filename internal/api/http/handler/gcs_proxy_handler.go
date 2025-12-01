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

// ProxyObject handles requests to proxy GCS objects
// @Summary      Proxy GCS Object
// @Description  Proxies a GCS object through the server
// @Tags         GCS
// @Produce      octet-stream
// @Param        objectPath path string true "Path of the GCS object to proxy"
// @Success      200 {file} binary "The requested GCS object"
// @Failure      400 {object} response.ErrorResponse "Invalid object path"
// @Failure      404 {object} response.ErrorResponse "Object not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /proxy/{objectPath} [get]
func (h *GCSProxyHandler) ProxyObject(c *gin.Context) {
	objectPath := strings.TrimPrefix(c.Param("objectPath"), "/")

	// FIX 1: Çift "proxy" prefix'ini temizle
	objectPath = strings.TrimPrefix(objectPath, "proxy/")

	if objectPath == "" {
		h.logger.Warn("GCS Proxy: Empty object path requested")
		c.String(http.StatusBadRequest, "object path is required")
		return
	}

	h.logger.Debug("GCS Proxy: Requesting object",
		"original_path", c.Param("objectPath"),
		"cleaned_path", objectPath)

	ctx := c.Request.Context()
	rc, err := h.gcsClient.Bucket(h.bucketName).Object(objectPath).NewReader(ctx)
	if err != nil {
		h.logger.Error("GCS Proxy: Object not found",
			"path", objectPath,
			"bucket", h.bucketName,
			"error", err)
		c.String(http.StatusNotFound, fmt.Sprintf("object not found: %s", err.Error()))
		return
	}
	defer rc.Close()

	// Set proper content type
	contentType := rc.ContentType()
	if contentType == "" {
		// Fallback content type detection
		if strings.HasSuffix(objectPath, ".jpg") || strings.HasSuffix(objectPath, ".jpeg") {
			contentType = "image/jpeg"
		} else if strings.HasSuffix(objectPath, ".dzi") {
			contentType = "application/xml"
		}
	}

	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000") // Cache for 1 year
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, rc)
	if err != nil {
		h.logger.Error("GCS Proxy: Error copying data", "error", err)
	}
}
