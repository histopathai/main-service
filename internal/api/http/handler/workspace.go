package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/request"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/response"
	"github.com/histopathai/main-service-refactor/internal/service"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type WorkspaceHandler struct {
	workspaceService *service.WorkspaceService
	BaseHandler      // Embed the BaseHandler
}

func NewWorkspaceHandler(workspaceService *service.WorkspaceService, logger *slog.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		workspaceService: workspaceService,
		BaseHandler:      BaseHandler{logger: logger},
	}
}

// CreateNewWorkspace godoc
// @Summary Create a new workspace
// @Description Create a new workspace with the provided details
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param        request body request.CreateWorkspaceRequest true "Workspace creation request"
// @Success 201 {object} response.DataResponse[response.WorkspaceResponse] "Workspace created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 409 {object} response.ErrorResponse "Workspace already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /workspaces [post]

func (wh *WorkspaceHandler) CreateNewWorkspace(c *gin.Context) {

	creator_id, exists := c.Get("user_id")
	if !exists {
		wh.handleError(c, errors.NewUnauthorizedError("unauthenticated"))
		return
	}

	var req request.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wh.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// DTO -> Service Input
	input := service.CreateWorkspaceInput{
		CreatorID:        creator_id.(string),
		Name:             req.Name,
		OrganType:        req.OrganType,
		AnnotationTypeID: req.AnnotationTypeID,
		Organization:     req.Organization,
		Description:      req.Description,
		License:          req.License,
		ResourceURL:      req.ResourceURL,
		ReleaseYear:      req.ReleaseYear,
	}

	workspace, err := wh.workspaceService.CreateNewWorkspace(c.Request.Context(), &input)
	if err != nil {
		wh.handleError(c, err)
		return
	}

	wh.logger.Info("Workspace created successfully",
		slog.String("workspace_id", workspace.ID),
	)

	c.JSON(http.StatusCreated, response.DataResponse[response.WorkspaceResponse]{
		Data: *response.NewWorkspaceResponse(workspace),
	})
}

//UpdateWorkspace godoc
// @Summary Update an existing workspace
// @Description Update the details of an existing workspace by ID
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param        id   path      string                      true  "Workspace ID"
// @Param        request body request.UpdateWorkspaceRequest true "Workspace update request"
// @Success 204 "Workspace updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "Workspace not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /workspaces/{id} [put]

func (wh *WorkspaceHandler) UpdateWorkspace(c *gin.Context) {

	workspaceID := c.Param("id")

	var req request.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wh.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// DTO -> Service Input
	input := service.UpdateWorkspaceInput{
		Name:             req.Name,
		OrganType:        req.OrganType,
		Organization:     req.Organization,
		Description:      req.Description,
		License:          req.License,
		ResourceURL:      req.ResourceURL,
		ReleaseYear:      req.ReleaseYear,
		AnnotationTypeID: req.AnnotationTypeID,
	}

	err := wh.workspaceService.UpdateWorkspace(c.Request.Context(), workspaceID, input)
	if err != nil {
		wh.handleError(c, err)
		return
	}

	wh.logger.Info("Workspace updated successfully",
		slog.String("workspace_id", workspaceID),
	)

	c.Status(http.StatusNoContent)
}

// GetWorkspaceByID godoc
// @Summary Get workspace details by ID
// @Description Retrieve the details of a workspace by its ID
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param        id   path      string  true  "Workspace ID"
// @Success 200 {object} response.DataResponse[response.WorkspaceResponse] "Workspace details retrieved successfully"
// @Failure 404 {object} response.ErrorResponse "Workspace not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /workspaces/{id} [get]

func (wh *WorkspaceHandler) GetWorkspaceByID(c *gin.Context) {

	workspaceID := c.Param("id")

	workspace, err := wh.workspaceService.GetWorkspaceByID(c.Request.Context(), workspaceID)
	if err != nil {
		wh.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.DataResponse[response.WorkspaceResponse]{
		Data: *response.NewWorkspaceResponse(workspace),
	})
}

// DeleteWorkspace godoc
// @Summary Delete a workspace by ID
// @Description Delete an existing workspace by its ID
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param        id   path      string  true  "Workspace ID"
// @Success 204 "Workspace deleted successfully"
// @Failure 404 {object} response.ErrorResponse "Workspace not found"
// @Failure 409 {object} response.ErrorResponse "Workspace associated with existing patients"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /workspaces/{id} [delete]
func (wh *WorkspaceHandler) DeleteWorkspace(c *gin.Context) {
	workspaceID := c.Param("id")

	err := wh.workspaceService.DeleteWorkspace(c.Request.Context(), workspaceID)
	if err != nil {
		wh.handleError(c, err)
		return
	}

	wh.logger.Info("Workspace deleted successfully",
		slog.String("workspace_id", workspaceID),
	)

	c.Status(http.StatusNoContent)
}

// ListWorkspaces godoc
// @Summary      List workspaces
// @Description  Get paginated list of workspaces
// @Tags         Workspaces
// @Accept       json
// @Produce      json
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success      200 {object} response.ListResponse[response.WorkspaceResponse] "List of workspaces"
// @Failure      400 {object} response.ErrorResponse "Invalid query parameters"
// @Failure      401 {object} response.ErrorResponse "Unauthorized"
// @Failure      500 {object} response.ErrorResponse "Internal server error"
// @Security     BearerAuth
// @Router       /workspaces [get]

func (wh *WorkspaceHandler) ListWorkspaces(c *gin.Context) {
	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		wh.handleError(c, errors.NewValidationError("invalid query parameters",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	pagination := &query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := wh.workspaceService.ListWorkspaces(c.Request.Context(), pagination)
	if err != nil {
		wh.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response.NewWorkspaceListResponse(result))
}
