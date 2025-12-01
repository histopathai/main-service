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
// @Description  Proxies a GCS object (image tiles, DZI, thumbnails) through the server
// @Tags         GCS
// @Produce      octet-stream
// @Param        imageId path string true "Image UUID"
// @Param        objectPath path string true "Path within the image folder"
// @Success      200 {file} binary "The requested GCS object"
// @Failure      400 {object} response.ErrorResponse "Invalid object path"
// @Failure      404 {object} response.ErrorResponse "Object not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /proxy/{imageId}/{objectPath} [get]
func (h *GCSProxyHandler) ProxyObject(c *gin.Context) {
	imageId := c.Param("imageId")
	objectPath := strings.TrimPrefix(c.Param("objectPath"), "/")

	if imageId == "" || objectPath == "" {
		h.logger.Warn("GCS Proxy: Missing imageId or objectPath",
			"imageId", imageId,
			"objectPath", objectPath)
		c.String(http.StatusBadRequest, "imageId and objectPath are required")
		return
	}

	// Construct full GCS path: {imageId}/{objectPath}
	// Example: 85893f19-ac05-4841-b02b-5cf93c4bf715/image.dzi
	fullPath := fmt.Sprintf("%s/%s", imageId, objectPath)

	h.logger.Debug("GCS Proxy: Requesting object",
		"imageId", imageId,
		"objectPath", objectPath,
		"fullPath", fullPath,
		"bucket", h.bucketName)

	ctx := c.Request.Context()
	rc, err := h.gcsClient.Bucket(h.bucketName).Object(fullPath).NewReader(ctx)
	if err != nil {
		h.logger.Error("GCS Proxy: Object not found",
			"fullPath", fullPath,
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
		} else if strings.HasSuffix(objectPath, ".png") {
			contentType = "image/png"
		}
	}

	// Set cache headers for better performance
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("ETag", fmt.Sprintf(`"%s"`, fullPath))

	c.Status(http.StatusOK)

	written, err := io.Copy(c.Writer, rc)
	if err != nil {
		h.logger.Error("GCS Proxy: Error copying data",
			"error", err,
			"bytesWritten", written)
	} else {
		h.logger.Debug("GCS Proxy: Successfully served object",
			"fullPath", fullPath,
			"bytesWritten", written,
			"contentType", contentType)
	}
}
