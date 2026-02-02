package handler

import (
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

type AnnotationHandler struct {
	helper.BaseHandler
	AQuery     port.AnnotationQuery
	AUseCase   port.AnnotationUseCase
	AValidator *validator.Validator
}

func NewAnnotationHandler(query port.AnnotationQuery, useCase port.AnnotationUseCase, logger *slog.Logger) *AnnotationHandler {
	return &AnnotationHandler{
		AQuery:      query,
		AUseCase:    useCase,
		AValidator:  validator.NewValidator(fields.NewAnnotationFieldSet()),
		BaseHandler: helper.NewBaseHandler(logger),
	}
}

// Create godoc
// @Summary Create a new annotation
// @Tags Annotations
// @Accept json
// @Produce json
// @Param request body request.CreateAnnotationRequest true "Annotation creation request"
// @Success 201 {object} response.AnnotationDataResponse "Annotation created successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations [post]
func (ah *AnnotationHandler) Create(c *gin.Context) {
	annotatorID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	var req request.CreateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ah.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	var polygon []command.CommandPoint
	if req.Polygon != nil && len(*req.Polygon) > 0 {
		points := make([]command.CommandPoint, len(*req.Polygon))
		for i, pr := range *req.Polygon {
			points[i] = command.CommandPoint{X: pr.X, Y: pr.Y}
		}
		polygon = points
	}

	cmd := command.CreateAnnotationCommand{
		CreateEntityCommand: command.CreateEntityCommand{
			Name:       req.Name,
			EntityType: vobj.EntityTypeAnnotation.String(),
			CreatorID:  annotatorID,
			ParentID:   req.Parent.ID,
			ParentType: req.Parent.Type,
		},
		TagType:  req.TagType,
		Value:    req.Value,
		Points:   polygon,
		Color:    req.Color,
		IsGlobal: req.IsGlobal,
	}

	errDetails, ok := cmd.Validate()
	if !ok {
		ah.HandleError(c, errors.NewValidationError("invalid command payload", errDetails))
		return
	}

	createdAnnotation, err := ah.AUseCase.Create(c.Request.Context(), cmd)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	annotationResp := response.NewAnnotationResponse(createdAnnotation)
	ah.Response.Created(c, annotationResp)
}

// Get godoc
// @Summary Get an annotation by ID
// @Tags Annotations
// @Accept json
// @Produce json
// @Param id path string true "Annotation ID"
// @Success 200 {object} response.AnnotationDataResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations/{id} [get]
func (ah *AnnotationHandler) Get(c *gin.Context) {
	annotationID := c.Param("id")
	if annotationID == "" {
		ah.HandleError(c, errors.NewValidationError("invalid annotation ID", map[string]interface{}{
			"id": "Annotation ID cannot be empty",
		}))
		return
	}

	annotation, err := ah.AQuery.Get(c.Request.Context(), annotationID)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	annotationResp := response.NewAnnotationResponse(annotation)
	ah.Response.Success(c, http.StatusOK, annotationResp)
}

// GetByParentID godoc
// @Summary Get annotations by Image ID
// @Description Get annotations belonging to a specific image with optional filtering, sorting, and pagination
// @Tags Annotations
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.AnnotationListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations/image/{image_id} [get]
func (ah *AnnotationHandler) GetByParentID(c *gin.Context) {
	imageID := c.Param("image_id")
	if imageID == "" {
		ah.HandleError(c, errors.NewValidationError("invalid image ID", map[string]interface{}{
			"image_id": "Image ID cannot be empty",
		}))
		return
	}

	var req request.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ah.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ah.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := ah.AValidator.ValidateSpec(spec); err != nil {
		ah.HandleError(c, err)
		return
	}

	result, err := ah.AQuery.GetByParentID(c.Request.Context(), spec, imageID)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	annotationsResp := make([]response.AnnotationResponse, len(result.Data))
	for i, at := range result.Data {
		annotationsResp[i] = *response.NewAnnotationResponse(at)
	}

	ah.Response.SuccessList(c, annotationsResp, paginationResp)
}

// GetByWsID godoc
// @Summary Get annotations by Workspace ID
// @Description Get annotations belonging to a specific workspace with optional filtering, sorting, and pagination
// @Tags Annotations
// @Accept json
// @Produce json
// @Param workspace_id path string true "Workspace ID"
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.AnnotationListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations/workspace/{workspace_id} [get]
func (ah *AnnotationHandler) GetByWsID(c *gin.Context) {
	workspaceID := c.Param("workspace_id")
	if workspaceID == "" {
		ah.HandleError(c, errors.NewValidationError("invalid workspace ID", map[string]interface{}{
			"workspace_id": "Workspace ID cannot be empty",
		}))
		return
	}

	var req request.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ah.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ah.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := ah.AValidator.ValidateSpec(spec); err != nil {
		ah.HandleError(c, err)
		return
	}

	result, err := ah.AQuery.GetByWsID(c.Request.Context(), spec, workspaceID)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	annotationsResp := make([]response.AnnotationResponse, len(result.Data))
	for i, at := range result.Data {
		annotationsResp[i] = *response.NewAnnotationResponse(at)
	}

	ah.Response.SuccessList(c, annotationsResp, paginationResp)
}

// Count godoc
// @Summary Count annotations
// @Description Count annotations with optional filters via query parameters
// @Tags Annotations
// @Accept json
// @Produce json
// @Success 200 {object} response.CountResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations/count [get]
func (ah *AnnotationHandler) Count(c *gin.Context) {
	var req request.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		ah.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ah.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := ah.AValidator.ValidateSpec(spec); err != nil {
		ah.HandleError(c, err)
		return
	}

	count, err := ah.AQuery.Count(c.Request.Context(), spec)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	countResp := response.CountResponse{Count: count}
	ah.Response.Success(c, http.StatusOK, countResp)
}

// Update godoc
// @Summary Update an annotation by ID
// @Tags Annotations
// @Accept json
// @Produce json
// @Param id path string true "Annotation ID"
// @Param request body request.UpdateAnnotationRequest true "Annotation update request"
// @Success 204 "Annotation updated successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations/{id} [put]
func (ah *AnnotationHandler) Update(c *gin.Context) {
	annotationID := c.Param("id")
	if annotationID == "" {
		ah.HandleError(c, errors.NewValidationError("invalid annotation ID", map[string]interface{}{
			"id": "Annotation ID cannot be empty",
		}))
		return
	}

	var req request.UpdateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ah.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	cmd := command.UpdateAnnotationCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{
			ID:   annotationID,
			Name: nil,
		},
		Value:    req.Value,
		IsGlobal: req.IsGlobal,
	}

	err := ah.AUseCase.Update(c.Request.Context(), cmd)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	ah.Response.NoContent(c)
}

// SoftDelete godoc
// @Summary Soft delete annotation by ID
// @Tags Annotations
// @Accept json
// @Produce json
// @Param id path string true "Annotation ID"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations/{id}/soft-delete [delete]
func (ah *AnnotationHandler) SoftDelete(c *gin.Context) {
	annotationID := c.Param("id")
	if annotationID == "" {
		ah.HandleError(c, errors.NewValidationError("invalid annotation ID", map[string]interface{}{
			"id": "Annotation ID cannot be empty",
		}))
		return
	}

	err := ah.AQuery.SoftDelete(c.Request.Context(), annotationID)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	ah.Response.NoContent(c)
}

// SoftDeleteMany godoc
// @Summary Batch soft delete annotations by IDs
// @Tags Annotations
// @Accept json
// @Produce json
// @Param ids query []string true "Annotation IDs"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotations/soft-delete-many [delete]
func (ah *AnnotationHandler) SoftDeleteMany(c *gin.Context) {
	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		ah.HandleError(c, errors.NewValidationError("ids parameter is required", nil))
		return
	}

	err := ah.AQuery.SoftDeleteMany(c.Request.Context(), ids)
	if err != nil {
		ah.HandleError(c, err)
		return
	}

	ah.Response.NoContent(c)
}
