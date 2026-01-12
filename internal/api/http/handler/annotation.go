package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/api/http/validator"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationHandler struct {
	annotationService port.IAnnotationService
	validator         *validator.RequestValidator
	BaseHandler       // Embed the BaseHandler
}

func NewAnnotationHandler(annotationService port.IAnnotationService, validator *validator.RequestValidator, logger *slog.Logger) *AnnotationHandler {
	return &AnnotationHandler{
		annotationService: annotationService,
		validator:         validator,
		BaseHandler:       BaseHandler{logger: logger},
	}
}

// CreateNewAnnotation [post] godoc
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
	entityInput := port.CreateEntityInput{
		Name: "",
		Type: vobj.EntityTypeAnnotation,
		Parent: &vobj.ParentRef{
			ID:   req.Parent.ID,
			Type: vobj.ParentTypeImage,
		},
		CreatorID: annotator_id,
	}

	tagInput := vobj.TagValue{
		Type:   vobj.TagType(req.Tag.TagType),
		Value:  req.Tag.Value,
		Color:  req.Tag.Color,
		Global: req.Tag.Global,
	}

	var polygon *[]vobj.Point
	if req.Polygon != nil && len(*req.Polygon) > 0 {
		points := make([]vobj.Point, len(*req.Polygon))
		for i, pr := range *req.Polygon {
			points[i] = vobj.NewPoint(pr.X, pr.Y)
		}
		polygon = &points
	}

	input := port.CreateAnnotationInput{
		CreateEntityInput: entityInput,
		TagValue:          tagInput,
		Polygon:           polygon,
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

// GetAnnotationByID [get] godoc
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
// @Router /annotations/{annotation_id} [get]
func (ah *AnnotationHandler) GetAnnotationByID(c *gin.Context) {
	annotationID := c.Param("annotation_id")
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

// GetAnnotationsByImageID [get] godoc
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

	pagination := queryReq.ToPagination()

	pagination.ApplyDefaults()

	if err := pagination.ValidateSortFields(request.ValidAnnotationSortFields); err != nil {
		ah.handleError(c, err)
		return
	}

	result, err := ah.annotationService.GetAnnotationsByImageID(c.Request.Context(), imageID, pagination)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	// Service Output -> DTO

	paginationResp := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	annotationsResp := make([]response.AnnotationResponse, len(result.Data))
	for i, at := range result.Data {
		annotationsResp[i] = *response.NewAnnotationResponse(at)
	}

	ah.response.SuccessList(c, annotationsResp, paginationResp)
}

// CountAnnotations [get] godoc
// @Summary Count annotations
// @Description Get the total count of annotations in the system
// @Tags Annotations
// @Accept json
// @Produce json
// @Success 200 {object} response.CountResponse "Total count of annotations"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotations/count-v1 [get]
func (ah *AnnotationHandler) CountV1Annotations(c *gin.Context) {
	count, err := ah.annotationService.CountAnnotations(c.Request.Context(), []query.Filter{})
	if err != nil {
		ah.handleError(c, err)
		return
	}

	countResp := response.CountResponse{
		Count: count,
	}

	ah.response.Success(c, http.StatusOK, countResp)
}

// UpdateAnnotation [put] godoc
// @Summary Update an annotation by ID
// @Description Update an annotation's details using its unique ID
// @Tags Annotations
// @Accept json
// @Produce json
// @Param id path string true "Annotation ID"
// @Param        request body request.UpdateAnnotationRequest true "Annotation update request"
// @Success 204 "Annotation updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 404 {object} response.ErrorResponse "Annotation not found"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotations/{annotation_id} [put]
func (ah *AnnotationHandler) UpdateAnnotation(c *gin.Context) {
	annotationID := c.Param("annotation_id")
	if annotationID == "" {
		ah.handleError(c, errors.NewValidationError("invalid annotation ID", map[string]interface{}{
			"annotation_id": "Annotation ID cannot be empty",
		}))
		return
	}

	var req request.UpdateAnnotationRequest
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

	// Entity input oluştur
	updateEntityInput := port.UpdateEntityInput{}

	// Parent varsa ekle
	if req.Parent != nil {
		parentRef, err := vobj.NewParentRef(req.Parent.ID, vobj.ParentType(req.Parent.Type))
		if err != nil {
			ah.handleError(c, errors.NewValidationError("invalid parent reference", map[string]interface{}{
				"error": err.Error(),
			}))
			return
		}
		updateEntityInput.Parent = parentRef
	}

	updateTagValue := port.UpdateTagValue{}
	if req.Tag != nil {
		if req.Tag.TagType != "" {
			tagType := vobj.TagType(req.Tag.TagType)
			updateTagValue.TagType = &tagType
		}
		if req.Tag.TagName != "" {
			updateTagValue.TagName = &req.Tag.TagName
		}
		if req.Tag.Value != nil {
			updateTagValue.Value = &req.Tag.Value
		}
		updateTagValue.Color = req.Tag.Color
		if req.Tag.Global {
			updateTagValue.Global = &req.Tag.Global
		}
	}

	var polygon *[]vobj.Point
	if req.Polygon != nil && len(*req.Polygon) > 0 {
		points := make([]vobj.Point, len(*req.Polygon))
		for i, pr := range *req.Polygon {
			points[i] = vobj.NewPoint(pr.X, pr.Y)
		}
		polygon = &points
	}

	input := port.UpdateAnnotationInput{
		UpdateEntityInput: updateEntityInput,
		Polygon:           polygon,
		UpdateTagValue:    updateTagValue,
	}

	err := ah.annotationService.UpdateAnnotation(c.Request.Context(), annotationID, &input)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	ah.logger.Info("Annotation updated successfully",
		slog.String("annotation_id", annotationID),
	)

	// No content to return
	ah.response.NoContent(c)
}

// DeleteAnnotation [delete] godoc
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
// @Router /annotations/{annotation_id} [delete]
func (ah *AnnotationHandler) DeleteAnnotation(c *gin.Context) {
	annotationID := c.Param("annotation_id")
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

// BatchDeleteAnnotations [post] godoc
// @Summary Batch delete annotations by IDs
// @Description Delete multiple annotations using their unique IDs
// @Tags Annotations
// @Accept json
// @Produce json
// @Param        request body request.BatchDeleteRequest true "Batch delete request"
// @Success 204 "Annotations deleted successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request payload"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Failure 401 {object} response.ErrorResponse "Unauthorized"
// @Security BearerAuth
// @Router /annotations/batch-delete [post]
func (ah *AnnotationHandler) BatchDeleteAnnotations(c *gin.Context) {
	var req request.BatchDeleteRequest
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

	err := ah.annotationService.BatchDeleteAnnotations(c.Request.Context(), req.IDs)
	if err != nil {
		ah.handleError(c, err)
		return
	}

	ah.logger.Info("Batch annotations deleted successfully",
		slog.Int("count", len(req.IDs)),
	)

	// No content to return
	ah.response.NoContent(c)
}
