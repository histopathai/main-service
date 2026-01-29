package handler

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/application/proxy"
)

type TileProxyHandler struct {
	tileServer *proxy.TileServer
	logger     *slog.Logger
}

func NewTileProxyHandler(tileServer *proxy.TileServer, logger *slog.Logger) *TileProxyHandler {
	return &TileProxyHandler{
		tileServer: tileServer,
		logger:     logger.WithGroup("tile_proxy"),
	}
}

// ProxyTile handles tile proxy requests
// @Summary      Proxy Tile Request
// @Description  Proxies DZI, thumbnails, and tile requests
// @Tags         Tiles
// @Produce      octet-stream
// @Param        imageId path string true "Image UUID"
// @Param        objectPath path string true "Object path (e.g., image.dzi, 0/0_0.jpeg)"
// @Success      200 {file} binary "The requested object"
// @Failure      400 {object} response.ErrorResponse "Invalid request"
// @Failure      404 {object} response.ErrorResponse "Object not found"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Router       /proxy/{imageId}/{objectPath} [get]
func (h *TileProxyHandler) ProxyTile(c *gin.Context) {
	imageID := c.Param("imageId")
	objectPath := strings.TrimPrefix(c.Param("objectPath"), "/")

	if imageID == "" || objectPath == "" {
		h.logger.Warn("Missing imageId or objectPath",
			"imageId", imageID,
			"objectPath", objectPath)
		c.JSON(http.StatusBadRequest, gin.H{"error": "imageId and objectPath are required"})
		return
	}

	h.logger.Debug("Tile request",
		"imageId", imageID,
		"objectPath", objectPath)

	ctx := c.Request.Context()
	reader, err := h.tileServer.ServeRequest(ctx, imageID, objectPath)
	if err != nil {
		h.logger.Error("Failed to serve request",
			"imageId", imageID,
			"objectPath", objectPath,
			"error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Set content type based on file extension
	contentType := h.getContentType(objectPath)
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")

	c.Status(http.StatusOK)
	written, err := io.Copy(c.Writer, reader)
	if err != nil {
		h.logger.Error("Error copying data", "error", err)
	} else {
		h.logger.Debug("Successfully served",
			"imageId", imageID,
			"objectPath", objectPath,
			"bytes", written)
	}
}

func (h *TileProxyHandler) getContentType(path string) string {
	lowerPath := strings.ToLower(path)

	switch {
	case strings.HasSuffix(lowerPath, ".dzi"):
		return "application/xml"
	case strings.HasSuffix(lowerPath, ".jpg"), strings.HasSuffix(lowerPath, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lowerPath, ".png"):
		return "image/png"
	case strings.HasSuffix(lowerPath, ".json"):
		return "application/json"
	default:
		return "application/octet-stream"
	}
}
