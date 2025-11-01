package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/request"
	"github.com/histopathai/main-service-refactor/internal/api/http/dto/response"
	"github.com/histopathai/main-service-refactor/internal/api/http/middleware"
	"github.com/histopathai/main-service-refactor/internal/api/http/validator"
	"github.com/histopathai/main-service-refactor/internal/service"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type AnnotationHandler struct {
	annotationService *service.AnnotationService
	validator         *validator.RequestValidator
	BaseHandler       // Embed the BaseHandler
}

func NewAnnotationHandler(annotationService *service.AnnotationService, validator *validator.RequestValidator, logger *slog.Logger) *AnnotationHandler {
	return &AnnotationHandler{
		annotationService: annotationService,
		validator:         validator,
		BaseHandler:       BaseHandler{logger: logger},
	}
}

// CreateNewAnnotation godoc
// @Summary Create a new annotation
// @Description Create a new annotation with the provided details
// @Tags Annotations
// @Accept json
// @Produce json
// @Param        request body request.CreateAnnotationRequest true "Annotation creation request"
// @Success 201 {object} response.AnnotationDataResponse "Annotation created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotations [post]
func (ah *AnnotationHandler) CreateNewAnnotation(c *gin.Context) {
	annotator_id, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	var req request.CreateAnnotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ah.handleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	if err := ah.validator.ValidateStruct(&req); err != nil {
		ah.handleError(c, err)
		return
	}

	// DTO -> Service Input
	input := service.CreateAnnotationInput{
		ImageID:     req.ImageID,
		AnnotatorID: annotator_id,
		Polygon:     req.Polygon,
		Score:       req.Score,
		Class:       req.Class,
		Description: req.Description,
	}

	createdAnnotation, err := ah.annotationService.CreateNewAnnotation(c.Request.Context(), &input)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	ah.logger.Info("Annotation created successfully",
		slog.String("annotation_id", createdAnnotation.ID),
	)

	// Service Output -> DTO
	annotationResp := response.NewAnnotationResponse(createdAnnotation)

	ah.response.Created(c, annotationResp)

}

// GetAnnotationByID godoc
// @Summary Get an annotation by ID
// @Description Retrieve an annotation using its unique ID
// @Tags Annotations
// @Accept json
// @Produce json
// @Param id path string true "Annotation ID"
// @Success 200 {object} response.AnnotationDataResponse "Annotation retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid annotation ID"
// @Failure 404 {object} response.ErrorResponse "Annotation not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotations/{id} [get]
func (ah *AnnotationHandler) GetAnnotationByID(c *gin.Context) {
	annotationID := c.Param("id")
	if annotationID == "" {
		ah.handleError(c, errors.NewValidationError("invalid annotation ID", map[string]interface{}{
			"annotation_id": "Annotation ID cannot be empty",
		}))
		return
	}

	annotation, err := ah.annotationService.GetAnnotationByID(c.Request.Context(), annotationID)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	// Service Output -> DTO
	annotationResp := response.NewAnnotationResponse(annotation)

	ah.response.Success(c, http.StatusOK, annotationResp)
}

// GetAnnotationsByImageID godoc
// @Summary Get annotations by Image ID
// @Description Retrieve annotations associated with a specific image ID
// @Tags Annotations
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param        limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param        offset query int false "Number of items to skip" default(0) minimum(0)
// @Param        sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, name)
// @Param        sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.AnnotationListResponse "List of annotations retrieved successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request parameters"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotations/image/{image_id} [get]
func (ah *AnnotationHandler) GetAnnotationsByImageID(c *gin.Context) {
	imageID := c.Param("image_id")
	if imageID == "" {
		ah.handleError(c, errors.NewValidationError("invalid image ID", map[string]interface{}{
			"image_id": "Image ID cannot be empty",
		}))
		return
	}

	var queryReq request.QueryPaginationRequest
	if err := c.ShouldBindQuery(&queryReq); err != nil {
		ah.handleError(c, errors.NewValidationError("invalid query parameters",
			map[string]interface{}{"error": err.Error()}))
		return
	}

	pagination := &query.Pagination{
		Limit:   queryReq.Limit,
		Offset:  queryReq.Offset,
		SortBy:  queryReq.SortBy,
		SortDir: queryReq.SortDir,
	}

	result, err := ah.annotationService.GetAnnotationsByImageID(c.Request.Context(), imageID, pagination)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	// Service Output -> DTO

	paginationResp := &response.PaginationResponse{
		Total:  result.Total,
		Limit:  queryReq.Limit,
		Offset: queryReq.Offset,
	}

	annotationsResp := make([]response.AnnotationResponse, 0, len(result.Data))
	for i, at := range result.Data {
		annotationsResp[i] = *response.NewAnnotationResponse(at)
	}

	ah.response.SuccessList(c, annotationsResp, paginationResp)
}

// DeleteAnnotation godoc
// @Summary Delete an annotation by ID
// @Description Delete an annotation using its unique ID
// @Tags Annotations
// @Accept json
// @Produce json
// @Param id path string true "Annotation ID"
// @Success 204 "Annotation deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid annotation ID"
// @Failure 404 {object} response.ErrorResponse "Annotation not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotations/{id} [delete]
func (ah *AnnotationHandler) DeleteAnnotation(c *gin.Context) {
	annotationID := c.Param("id")
	if annotationID == "" {
		ah.handleError(c, errors.NewValidationError("invalid annotation ID", map[string]interface{}{
			"annotation_id": "Annotation ID cannot be empty",
		}))
		return
	}

	err := ah.annotationService.DeleteAnnotation(c.Request.Context(), annotationID)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	ah.logger.Info("Annotation deleted successfully",
		slog.String("annotation_id", annotationID),
	)

	// No content to return
	ah.response.NoContent(c)
}
