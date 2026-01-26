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

type AnnotationTypeHandler struct {
	helper.BaseHandler
	ATQuery     port.AnnotationTypeQuery
	ATUseCase   port.AnnotationTypeUseCase
	ATValidator *validator.Validator
}

func NewAnnotationTypeHandler(query port.AnnotationTypeQuery, useCase port.AnnotationTypeUseCase, logger *slog.Logger) *AnnotationTypeHandler {
	return &AnnotationTypeHandler{
		ATQuery:     query,
		ATUseCase:   useCase,
		ATValidator: validator.NewValidator(fields.NewAnnotationTypeFieldSet()),
		BaseHandler: helper.NewBaseHandler(logger),
	}
}

// Create godoc
// @Summary Create a new annotation type
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param        request body request.CreateAnnotationTypeRequest true "Annotation Type creation request"
// @Success 201 {object} response.AnnotationTypeDataResponse "Annotation Type created successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-types [post]
func (ath *AnnotationTypeHandler) Create(c *gin.Context) {
	creator_id, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	var req request.CreateAnnotationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ath.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	cmd := command.CreateAnnotationTypeCommand{
		CreateEntityCommand: command.CreateEntityCommand{
			Name:       req.Name,
			EntityType: vobj.EntityTypeAnnotationType.String(),
			CreatorID:  creator_id,
			ParentID:   "",
			ParentType: vobj.ParentTypeNone.String(),
		},
		TagType:    req.TagType,
		IsGlobal:   req.IsGlobal,
		IsRequired: req.IsRequired,
		Options:    req.Options,
		Min:        req.Min,
		Max:        req.Max,
		Color:      req.Color,
	}

	errDetails, ok := cmd.Validate()
	if !ok {
		ath.HandleError(c, errors.NewValidationError("invalid command payload", errDetails))
		return
	}

	result, err := ath.ATUseCase.Create(c.Request.Context(), cmd)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	annotation_resp := response.NewAnnotationTypeResponse(result)
	ath.Response.Created(c, annotation_resp)
}

// Get [get] godoc
// @Summary Get an annotation type by ID
// @Description Retrieve the details of an annotation type using its ID
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param id path string true "Annotation Type ID"
// @Success 200 {object} response.AnnotationTypeDataResponse "Annotation Type retrieved successfully"
// @Failure 404 {object} response.ErrorResponse "Annotation Type not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotation-types/{id} [get]
func (ath *AnnotationTypeHandler) Get(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		ath.HandleError(c, errors.NewValidationError("invalid id", map[string]interface{}{
			"id": "ID cannot be empty",
		}))
		return
	}

	result, err := ath.ATQuery.Get(c.Request.Context(), id)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	annotation_resp := response.NewAnnotationTypeResponse(result)
	ath.Response.Success(c, http.StatusOK, annotation_resp)
}

// List [get] godoc
// @Summary List all annotation types
// @Description Retrieve a list of all annotation types
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.AnnotationTypeListResponse "List of annotation types retrieved successfully"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotation-types [get]
func (ath *AnnotationTypeHandler) List(c *gin.Context) {
	var req request.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = request.ListRequest{}
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ath.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := ath.ATValidator.ValidateSpec(spec); err != nil {
		ath.HandleError(c, err)
		return
	}

	result, err := ath.ATQuery.List(c.Request.Context(), spec)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	annotationResponses := make([]response.AnnotationTypeResponse, len(result.Data))
	for i, at := range result.Data {
		annotationResponses[i] = *response.NewAnnotationTypeResponse(at)
	}

	ath.Response.SuccessList(c, annotationResponses, paginationResp)
}

// Count godoc
// @Summary Count annotation types
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Success 200 {object} response.CountResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-types/count [post]
func (ath *AnnotationTypeHandler) Count(c *gin.Context) {
	var req request.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req = request.ListRequest{}
	}

	spec, err := req.ToSpecification()
	if err != nil {
		ath.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	if err := ath.ATValidator.ValidateSpec(spec); err != nil {
		ath.HandleError(c, err)
		return
	}

	count, err := ath.ATQuery.Count(c.Request.Context(), spec)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	countResp := &response.CountResponse{
		Count: count,
	}

	ath.Response.Success(c, http.StatusOK, countResp)
}

// Update [put] godoc
// @Summary Update an annotation type
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param id path string true "Annotation Type ID"
// @Param request body request.UpdateAnnotationTypeRequest true "Annotation Type update request"
// @Success 204  "Annotation Type updated successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-types/{id} [put]
func (ath *AnnotationTypeHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var req request.UpdateAnnotationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ath.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	cmd := command.UpdateAnnotationTypeCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{
			ID:   id,
			Name: req.Name,
		},
		IsGlobal:   req.IsGlobal,
		IsRequired: req.IsRequired,
		Options:    req.Options,
		Min:        req.Min,
		Max:        req.Max,
		Color:      req.Color,
	}

	err := ath.ATUseCase.Update(c.Request.Context(), cmd)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	ath.Response.NoContent(c)
}

// SoftDelete godoc
// @Summary Soft delete annotation type
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param id path string true "Annotation Type ID"
// @Success 204
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-types/{id}/soft-delete [delete]
func (ath *AnnotationTypeHandler) SoftDelete(c *gin.Context) {
	id := c.Param("id")

	err := ath.ATQuery.SoftDelete(c.Request.Context(), id)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	ath.Response.NoContent(c)
}

// SoftDeleteMany godoc
// @Summary Batch soft delete annotation types
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param ids query []string true "Annotation Type IDs"
// @Success 204
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-types/soft-delete-many [delete]
func (ath *AnnotationTypeHandler) SoftDeleteMany(c *gin.Context) {

	ids := c.QueryArray("ids")
	if len(ids) == 0 {
		ath.HandleError(c, errors.NewValidationError("ids is required", nil))
		return
	}

	err := ath.ATQuery.SoftDeleteMany(c.Request.Context(), ids)
	if err != nil {
		ath.HandleError(c, err)
		return
	}

	ath.Response.NoContent(c)
}
