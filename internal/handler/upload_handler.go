package handler

import (
	"log/slog"
	"net/http"

	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/service"
)

type UploadHandler struct {
	uploadService *service.UploadService
}

func NewUploadHandler(gcs_client *storage.Client, bucketName string, repo *repository.MainRepository, logger *slog.Logger) *UploadHandler {
	return &UploadHandler{
		uploadService: service.NewUploadService(
			gcs_client,
			bucketName,
			repo,
			logger),
	}
}

func (h *UploadHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/process-upload", h.ProcessUpload)
}

func (h *UploadHandler) ProcessUpload(c *gin.Context) {

	var req service.UploadImageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	resp, err := h.uploadService.ProcessUpload(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
