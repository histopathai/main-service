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

type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
	logger           *slog.Logger
	validate         *validator.Validate
}

func NewWorkspaceHandler(repo *repository.MainRepository, logger *slog.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: service.NewWorkspaceService(
			repository.NewWorkspaceRepository(repo),
			logger),
		logger:   logger,
		validate: validator.New(),
	}
}

func (h *WorkspaceHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/workspaces", h.CreateWorkspace)
	rg.GET("/workspaces/:id", h.GetWorkspace)
	rg.PATCH("/workspaces/:id", h.UpdateWorkspace)
	rg.GET("/creators/:creator_id/workspaces", h.GetWorkspacesByCreatorID)
	rg.GET("/workspaces", h.GetAllWorkspaces)

}

func (h *WorkspaceHandler) validateAndBind(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(req); err != nil {
		return err
	}
	return h.validate.Struct(req)
}

func (h *WorkspaceHandler) CreateWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest

	if err := h.validateAndBind(c, &req); err != nil {
		h.logger.Warn("CreateWorkspace validation failed", "error", err.Error())
		handleError(c, err)
		return
	}

	resp, err := h.workspaceService.CreateWorkspace(c.Request.Context(), req.ToModel())
	if err != nil {
		h.logger.Error("CreateWorkspace service error", "error", err.Error())
		handleError(c, err)
		return
	}

	h.logger.Info("workspace created successfully", "workspace_id", resp)
	c.JSON(http.StatusCreated, resp)
}

func (h *WorkspaceHandler) GetWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		err := apperrors.NewValidationError("workspace ID is required", nil)
		h.logger.Warn("GetWorkspace missing id parameter")
		handleError(c, err)
		return
	}

	workspace, err := h.workspaceService.GetWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		h.logger.Error("GetWorkspace service error", "workspace_id", workspaceID, "error", err.Error())
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, workspace)
}

func (h *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")
	if workspaceID == "" {
		err := apperrors.NewValidationError("workspace ID is required", nil)
		h.logger.Warn("UpdateWorkspace missing id parameter")
		handleError(c, err)
		return
	}

	var req UpdateWorkspaceRequest

	if err := h.validateAndBind(c, &req); err != nil {
		h.logger.Warn("UpdateWorkspace validation failed", "workspace_id", workspaceID, "error", err.Error())
		handleError(c, err)
		return
	}

	err := h.workspaceService.UpdateWorkspace(c.Request.Context(), workspaceID, req.ToUpdateMap())
	if err != nil {
		h.logger.Error("UpdateWorkspace service error", "workspace_id", workspaceID, "error", err.Error())
		handleError(c, err)
		return
	}

	h.logger.Info("workspace updated successfully", "workspace_id", workspaceID)
	c.Status(http.StatusNoContent)
}

func (h *WorkspaceHandler) GetWorkspacesByCreatorID(c *gin.Context) {
	creatorID := c.Param("creator_id")
	if creatorID == "" {
		err := apperrors.NewValidationError("creator ID is required", nil)
		h.logger.Warn("GetWorkspacesByCreatorID missing creator_id parameter")
		handleError(c, err)
		return
	}

	pagination, err := parsePagination(c)
	if err != nil {
		h.logger.Warn("GetWorkspacesByCreatorID pagination parse failed", "creator_id", creatorID, "error", err.Error())
		handleError(c, err)
		return
	}

	result, err := h.workspaceService.GetWorkspacesByCreatorID(c.Request.Context(), creatorID, pagination)
	if err != nil {
		h.logger.Error("GetWorkspacesByCreatorID service error", "creator_id", creatorID, "error", err.Error())
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *WorkspaceHandler) GetAllWorkspaces(c *gin.Context) {
	pagination, err := parsePagination(c)
	if err != nil {
		h.logger.Warn("GetAllWorkspaces pagination parse failed", "error", err.Error())
		handleError(c, err)
		return
	}

	result, err := h.workspaceService.GetAllWorkspaces(c.Request.Context(), pagination)
	if err != nil {
		h.logger.Error("GetAllWorkspaces service error", "error", err.Error())
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}
