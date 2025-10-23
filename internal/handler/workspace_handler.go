package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	apperrors "github.com/histopathai/main-service/internal/errors"
	"github.com/histopathai/main-service/internal/repository"
	"github.com/histopathai/main-service/internal/service"
)

var validate = validator.New()

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
	var req CreateWorkspaceRequest
	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate the request
	err := validate.Struct(&req)
	if err != nil {
		handleError(c, err)
		return
	}

	resp, err := h.workspaceService.CreateWorkspace(c.Request.Context(), req.ToModel())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		err := apperrors.NewValidationError("workspace ID is required", nil)
		handleError(c, err)
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	var req UpdateWorkspaceRequest

	// Bind JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Validate the request
	err := validate.Struct(&req)
	if err != nil {
		handleError(c, err)
		return
	}

	err = h.workspaceService.UpdateWorkspace(c.Request.Context(), c.Param("id"), req.ToUpdateMap())
	if err != nil {
		handleError(c, err)
		return
	}

	//Success
	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) GetWorkspacesByCreatorID(c *gin.Context) {
	creatorID := c.Param("creator_id")
	if creatorID == "" {
		err := apperrors.NewValidationError("creator ID is required", nil)
		handleError(c, err)
		return
	}

	pagination, err := parsePagination(c)
	if err != nil {
		handleError(c, err)
		return
	}

	result, err := h.workspaceService.GetWorkspacesByCreatorID(c.Request.Context(), creatorID, pagination)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *WorkspaceHandler) GetAllWorkspaces(c *gin.Context) {
	pagination, err := parsePagination(c)
	if err != nil {
		handleError(c, err)
		return
	}

	result, err := h.workspaceService.GetAllWorkspaces(c.Request.Context(), pagination)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
