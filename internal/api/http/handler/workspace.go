package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/handler/helper"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	validator "github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceHandler struct {
	helper.BaseHandler
	WsQuery     port.WorkspaceQuery
	WsUsecase   port.WorkspaceUseCase
	WsValidator *validator.Validator
}

func NewWorkspaceHandler(wsQuery port.WorkspaceQuery, wsUsecase port.WorkspaceUseCase, logger *slog.Logger) *WorkspaceHandler {
	return &WorkspaceHandler{
		WsQuery:     wsQuery,
		WsUsecase:   wsUsecase,
		WsValidator: validator.NewValidator(fields.NewWorkspaceFieldSet()),
		BaseHandler: helper.NewBaseHandler(logger),
	}
}

// Create godoc
// @Summary Create a new workspace
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param        request body request.CreateWorkspaceRequest true "Workspace data"
// @Success 201 {object} response.WorkspaceDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces [post]
func (wh *WorkspaceHandler) Create(c *gin.Context) {

	creatorID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		wh.HandleError(c, err)
		return
	}

	var req request.CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wh.HandleError(c, err)
		return
	}

	// DTO -> CMD
	cmd := command.CreateWorkspaceCommand{
		CreateEntityCommand: command.CreateEntityCommand{
			Name:       req.Name,
			EntityType: vobj.EntityTypeWorkspace.String(),
			CreatorID:  creatorID,
			ParentID:   "",
			ParentType: vobj.ParentTypeNone.String(),
		},
		OrganType:       req.OrganType,
		AnnotationTypes: req.AnnotationTypes,
		Organization:    req.Organization,
		Description:     req.Description,
		License:         req.License,
		ResourceURL:     req.ResourceURL,
		ReleaseYear:     req.ReleaseYear,
	}

	// Service Output -> DTO

	errDetails, ok :=
		cmd.Validate()
	if !ok {
		wh.HandleError(c, errors.NewValidationError("Invalid request", errDetails))
		return
	}

	workspace, err := wh.WsUsecase.Create(c.Request.Context(), cmd)
	if err != nil {
		wh.HandleError(c, err)
		return
	}

	wh.Response.Created(c, response.NewWorkspaceResponse(workspace))
}

// Get godoc
// @Summary Get workspace by ID
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} response.WorkspaceDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces/{id} [get]
func (wh *WorkspaceHandler) Get(c *gin.Context) {
	workspace, err := wh.WsQuery.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		wh.HandleError(c, err)
		return
	}

	wh.Response.Success(c, http.StatusOK, response.NewWorkspaceResponse(workspace))

}

// List godoc
// @Summary List workspaces
// @Description List workspaces with optional filtering, sorting, and pagination via query parameters
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name, organ_type)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.WorkspaceListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces [get]
func (wh *WorkspaceHandler) List(c *gin.Context) {
	var req request.ListRequest

	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		wh.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		wh.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := wh.WsValidator.ValidateSpec(spec); err != nil {
		wh.HandleError(c, err)
		return
	}

	workspaces, err := wh.WsQuery.List(c.Request.Context(), spec)
	if err != nil {
		wh.HandleError(c, err)
		return
	}

	wh.Response.Success(c, http.StatusOK, response.NewWorkspaceListResponse(workspaces))
}

// Update godoc
// @Summary Update workspace
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body request.UpdateWorkspaceRequest true "Workspace update request"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces/{id} [put]
func (wh *WorkspaceHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request.UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		wh.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// DTO -> CMD
	cmd := command.UpdateWorkspaceCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{
			ID:        id,
			Name:      req.Name,
			CreatorID: req.CreatorID,
		},
		OrganType:       req.OrganType,
		Organization:    req.Organization,
		Description:     req.Description,
		License:         req.License,
		ResourceURL:     req.ResourceURL,
		ReleaseYear:     req.ReleaseYear,
		AnnotationTypes: req.AnnotationTypes,
	}

	errDetails, ok := cmd.Validate()
	if !ok {
		wh.HandleError(c, errors.NewValidationError("Invalid request", errDetails))
		return
	}

	err := wh.WsUsecase.Update(c.Request.Context(), cmd)
	if err != nil {
		wh.HandleError(c, err)
		return
	}
	wh.Response.NoContent(c)
}

// SoftDelete godoc
// @Summary Soft delete workspace
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces/{id}/soft-delete [delete]
func (wh *WorkspaceHandler) SoftDelete(c *gin.Context) {

	err := wh.WsQuery.SoftDelete(c.Request.Context(), c.Param("id"))
	if err != nil {
		wh.HandleError(c, err)
		return
	}
	wh.Response.NoContent(c)
}

// Count godoc
// @Summary Count workspaces
// @Description Count workspaces with optional filters via query parameters
// @Tags Workspaces
// @Accept json
// @Produce json
// @Success 200 {object} response.CountResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces/count [get]
func (wh *WorkspaceHandler) Count(c *gin.Context) {
	var req request.ListRequest
	var spec validator.Specification

	// Try to bind query parameters (for GET requests)
	if err := c.ShouldBindQuery(&req); err != nil {
		// If binding fails, check if it's because of empty query params
		if err != io.EOF {
			wh.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
				"error": err.Error(),
			}))
			return
		}
	}

	// Convert to specification
	spec, err := req.ToSpecification()
	if err != nil {
		wh.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := wh.WsValidator.ValidateSpec(spec); err != nil {
		wh.HandleError(c, err)
		return
	}

	count, err := wh.WsQuery.Count(c.Request.Context(), spec)
	if err != nil {
		wh.HandleError(c, err)
		return
	}
	wh.Response.Success(c, http.StatusOK, response.CountResponse{Count: count})
}

// SoftDeleteMany godoc
// @Summary Batch soft delete workspaces
// @Tags Workspaces
// @Accept json
// @Produce json
// @Param ids query []string true "Workspace IDs"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /workspaces/soft-delete-many [delete]
func (wh *WorkspaceHandler) SoftDeleteMany(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		wh.HandleError(c, errors.NewValidationError("ids parameter is required", nil))
		return
	}

	err := wh.WsQuery.SoftDeleteMany(c.Request.Context(), ids)
	if err != nil {
		wh.HandleError(c, err)
		return
	}
	wh.Response.NoContent(c)
}
