package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/api/http/validator"

	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeHandler struct {
	annotationTypeService *service.AnnotationTypeService
	validator             *validator.RequestValidator
	BaseHandler           // Embed the BaseHandler
}

func NewAnnotationTypeHandler(annotationTypeService *service.AnnotationTypeService, validator *validator.RequestValidator, logger *slog.Logger) *AnnotationTypeHandler {
	return &AnnotationTypeHandler{
		annotationTypeService: annotationTypeService,
		validator:             validator,
		BaseHandler:           BaseHandler{logger: logger},
	}
}

// CreateNewAnnotationType godoc
// @Summary Create a new annotation type
// @Description Create a new annotation type with the provided details
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param        request body request.CreateAnnotationTypeRequest true "Annotation Type creation request"
// @Success 201 {object} response.AnnotationTypeDataResponse "Annotation Type created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 409 {object} response.ErrorResponse "Annotation Type already exists"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotation-types [post]

func (ath *AnnotationTypeHandler) CreateNewAnnotationType(c *gin.Context) {

	creator_id, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ath.handleError(c, err)
		return
	}

	var req request.CreateAnnotationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ath.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}
	if err := ath.validator.ValidateStruct(&req); err != nil {
		ath.handleError(c, err)
		return
	}

	// DTO -> Service Input
	var classList []string
	if req.ClassList != nil {
		classList = *req.ClassList
	}
	input := service.CreateAnnotationTypeInput{
		CreatorID:             creator_id,
		Name:                  req.Name,
		Description:           req.Description,
		ScoreEnabled:          req.ScoreEnabled,
		ScoreName:             req.ScoreName,
		ScoreMin:              req.ScoreMin,
		ScoreMax:              req.ScoreMax,
		ClassificationEnabled: req.ClassificationEnabled,
		ClassList:             classList,
	}

	result, err := ath.annotationTypeService.CreateNewAnnotationType(c.Request.Context(), &input)
	if err != nil {
		ath.handleError(c, err)
		return
	}
	ath.logger.Info("Annotation type created successfully",
		slog.String("annotation_type_id", result.ID),
	)

	// Server Output -> DTO

	annotation_resp := response.NewAnnotationTypeResponse(result)

	ath.response.Created(c, annotation_resp)

}

// GetAnnotationType godoc
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
func (ath *AnnotationTypeHandler) GetAnnotationType(c *gin.Context) {
	id := c.Param("id")

	result, err := ath.annotationTypeService.GetAnnotationTypeByID(c.Request.Context(), id)
	if err != nil {
		ath.handleError(c, err)
		return
	}

	ath.logger.Info("Annotation type retrieved successfully",
		slog.String("annotation_type_id", id),
	)

	// Service Output -> DTO
	annotation_resp := response.NewAnnotationTypeResponse(result)

	ath.response.Success(c, http.StatusOK, annotation_resp)

}

// ListAnnotationTypes godoc
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
func (ath *AnnotationTypeHandler) ListAnnotationTypes(c *gin.Context) {
	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		ath.handleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	pagination := &query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := ath.annotationTypeService.ListAnnotationTypes(c.Request.Context(), pagination)
	if err != nil {
		ath.handleError(c, err)
		return
	}

	ath.logger.Info("Annotation types listed successfully")

	// Service Output -> DTO
	paginationResp := &response.PaginationResponse{
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		Total:   result.Total,
		HasMore: result.HasMore,
	}
	annotationResponses := make([]response.AnnotationTypeResponse, len(result.Data))
	for i, at := range result.Data {
		annotationResponses[i] = *response.NewAnnotationTypeResponse(at)
	}

	ath.response.SuccessList(c, annotationResponses, paginationResp)

}

// UpdateAnnotationType godoc
// @Summary Update an annotation type
// @Description Update the details of an existing annotation type
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param id path string true "Annotation Type ID"
// @Param        request body request.UpdateAnnotationTypeRequest true "Annotation Type update request"
// @Success 204  "Annotation Type updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "Annotation Type not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotation-types/{id} [put]
func (ath *AnnotationTypeHandler) UpdateAnnotationType(c *gin.Context) {
	id := c.Param("id")

	var req request.UpdateAnnotationTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ath.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	// DTO -> Service Input
	input := service.UpdateAnnotationTypeInput{
		Name:        req.Name,
		Description: req.Description,
	}

	err := ath.annotationTypeService.UpdateAnnotationType(c.Request.Context(), id, &input)
	if err != nil {
		ath.handleError(c, err)
		return
	}

	ath.logger.Info("Annotation type updated successfully",
		slog.String("annotation_type_id", id),
	)

	// No content to return
	ath.response.NoContent(c)
}

// DeleteAnnotationType godoc
// @Summary Delete an annotation type
// @Description Delete an existing annotation type by ID
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param id path string true "Annotation Type ID"
// @Success 204 "Annotation Type deleted successfully"
// @Failure 404 {object} response.ErrorResponse "Annotation Type not found"
// @Failure 409 {object} response.ErrorResponse "Annotation Type in use"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotation-types/{id} [delete]
func (ath *AnnotationTypeHandler) DeleteAnnotationType(c *gin.Context) {
	id := c.Param("id")

	err := ath.annotationTypeService.DeleteAnnotationType(c.Request.Context(), id)
	if err != nil {
		ath.handleError(c, err)
		return
	}

	ath.logger.Info("Annotation type deleted successfully",
		slog.String("annotation_type_id", id),
	)

	// No content to return
	ath.response.NoContent(c)
}

// Get Classification Optionied Annotation Types godoc
// @Summary Get annotation types with classification enabled
// @Description Retrieve a list of annotation types that have classification enabled
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.AnnotationTypeListResponse "List of classification optioned annotation types retrieved successfully"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotation-types/classification-enabled [get]
func (ath *AnnotationTypeHandler) GetClassificationOptionedAnnotationTypes(c *gin.Context) {

	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		ath.handleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	pagination := &query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := ath.annotationTypeService.GetClassificationAnnotationTypes(c.Request.Context(), pagination)
	if err != nil {
		ath.handleError(c, err)
		return
	}

	ath.logger.Info("Classification optioned annotation types listed successfully")

	// Service Output -> DTO
	paginationResp := &response.PaginationResponse{
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		Total:   result.Total,
		HasMore: result.HasMore,
	}

	annotationResponses := make([]response.AnnotationTypeResponse, len(result.Data))
	for i, at := range result.Data {
		annotationResponses[i] = *response.NewAnnotationTypeResponse(at)
	}

	ath.response.SuccessList(c, annotationResponses, paginationResp)

}

// Get Score Optionied Annotation Types godoc
// @Summary Get annotation types with score enabled
// @Description Retrieve a list of annotation types that have score enabled
// @Tags Annotation Types
// @Accept json
// @Produce json
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.AnnotationTypeListResponse "List of score optioned annotation types retrieved successfully"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotation-types/score-enabled [get]
func (ath *AnnotationTypeHandler) GetScoreOptionedAnnotationTypes(c *gin.Context) {

	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		ath.handleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	pagination := &query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := ath.annotationTypeService.GetScoreAnnotationTypes(c.Request.Context(), pagination)
	if err != nil {
		ath.handleError(c, err)
		return
	}

	ath.logger.Info("Score optioned annotation types listed successfully")

	// Service Output -> DTO
	paginationResp := &response.PaginationResponse{
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		Total:   result.Total,
		HasMore: result.HasMore,
	}

	annotationResponses := make([]response.AnnotationTypeResponse, len(result.Data))
	for i, at := range result.Data {
		annotationResponses[i] = *response.NewAnnotationTypeResponse(at)
	}

	ath.response.SuccessList(c, annotationResponses, paginationResp)

}
