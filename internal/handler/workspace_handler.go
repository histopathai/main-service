package handler

import (
	"log/slog"
	"net/http"
	"strconv"

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

func (h *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	// Parse query parameters
	filters := make(map[string]interface{})

	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if organType := c.Query("organ_type"); organType != "" {
		filters["organ_type"] = organType
	}
	if creatorID := c.Query("creator_id"); creatorID != "" {
		filters["creator_id"] = creatorID
	}

	// Parse pagination
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if o := c.Query("offset"); o != "" {
		if parsedOffset, err := strconv.Atoi(o); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	pagination := repository.Pagination{
		Limit:  limit,
		Offset: offset,
	}

	workspaces, err := h.workspaceService.ListWorkspaces(c.Request.Context(), filters, pagination)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"workspaces": workspaces,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(workspaces),
		},
	})
}

func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workspace id is required",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	err := h.workspaceService.UpdateWorkspace(c.Request.Context(), workspaceID, updates)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "workspace updated successfully",
	})
}

func (h *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workspace id is required",
		})
		return
	}

	err := h.workspaceService.DeleteWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "workspace deleted successfully",
	})
}
