package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/service"
)

type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
}

func NewWorkspaceHandler(repo *repository.MainRepository, logger *slog.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: service.NewWorkspaceService(
			repository.NewWorkspaceRepository(repo),
			logger),
	}
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	var req service.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}
	resp, err := h.workspaceService.CreateWorkspace(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *WorkspaceHandler) GetWorkspaces(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workspace id is required",
		})
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, workspace)
}
