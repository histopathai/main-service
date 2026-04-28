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
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	validator "github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationReviewHandler struct {
	helper.BaseHandler
	ARQuery     port.AnnotationReviewQuery
	ARUseCase   port.AnnotationReviewUseCase
	ARValidator *validator.Validator
}

func NewAnnotationReviewHandler(query port.AnnotationReviewQuery, useCase port.AnnotationReviewUseCase, logger *slog.Logger) *AnnotationReviewHandler {
	return &AnnotationReviewHandler{
		ARQuery:     query,
		ARUseCase:   useCase,
		ARValidator: validator.NewValidator(fields.NewAnnotationReviewFieldSet()),
		BaseHandler: helper.NewBaseHandler(logger),
	}
}

// Create godoc
// @Summary Create an annotation review
// @Description Submit a review for an annotation. The reviewer cannot review their own manual annotation.
// @Tags Annotation Reviews
// @Accept json
// @Produce json
// @Param request body request.CreateAnnotationReviewRequest true "Annotation review request"
// @Success 201 {object} response.AnnotationReviewDataResponse "Review created successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-reviews [post]
func (h *AnnotationReviewHandler) Create(c *gin.Context) {
	reviewerID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	var req request.CreateAnnotationReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	var modifiedPolygon *[]command.CommandPoint
	if req.ModifiedPolygon != nil && len(*req.ModifiedPolygon) > 0 {
		points := make([]command.CommandPoint, len(*req.ModifiedPolygon))
		for i, p := range *req.ModifiedPolygon {
			points[i] = command.CommandPoint{X: p.X, Y: p.Y}
		}
		modifiedPolygon = &points
	}

	cmd := command.CreateAnnotationReviewCommand{
		AnnotationID:    req.AnnotationID,
		ReviewerID:      reviewerID,
		Status:          req.Status,
		Comments:        req.Comments,
		ModifiedValue:   req.ModifiedValue,
		ModifiedPolygon: modifiedPolygon,
	}

	if errDetails, ok := cmd.Validate(); !ok {
		h.HandleError(c, errors.NewValidationError("invalid command payload", errDetails))
		return
	}

	createdReview, err := h.ARUseCase.Create(c.Request.Context(), cmd)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.Response.Created(c, response.NewAnnotationReviewResponse(createdReview))
}

// Delete godoc
// @Summary Delete an annotation review
// @Description Soft-deletes a review. Can only be done by the reviewer who created it.
// @Tags Annotation Reviews
// @Produce json
// @Param id path string true "Review ID"
// @Success 204 "Review deleted successfully"
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-reviews/{id} [delete]
func (h *AnnotationReviewHandler) Delete(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		h.HandleError(c, errors.NewValidationError("review ID is required", nil))
		return
	}

	requesterID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	if err := h.ARUseCase.Delete(c.Request.Context(), reviewID, requesterID); err != nil {
		h.HandleError(c, err)
		return
	}

	h.Response.NoContent(c)
}

// Get godoc
// @Summary Get an annotation review by ID
// @Tags Annotation Reviews
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} response.AnnotationReviewDataResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-reviews/{id} [get]
func (h *AnnotationReviewHandler) Get(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		h.HandleError(c, errors.NewValidationError("review ID is required", nil))
		return
	}

	review, err := h.ARQuery.Get(c.Request.Context(), reviewID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.Response.Success(c, http.StatusOK, response.NewAnnotationReviewResponse(review))
}

// GetByAnnotationID godoc
// @Summary List reviews for an annotation
// @Description Get annotation reviews belonging to a specific annotation with optional filtering, sorting, and pagination
// @Tags Annotation Reviews
// @Produce json
// @Param annotation_id path string true "Annotation ID"
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, reviewer_id, status)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.AnnotationReviewListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-reviews/annotation/{annotation_id} [get]
func (h *AnnotationReviewHandler) GetByAnnotationID(c *gin.Context) {
	annotationID := c.Param("annotation_id")
	if annotationID == "" {
		h.HandleError(c, errors.NewValidationError("annotation ID is required", nil))
		return
	}

	var req request.ListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	spec, err := req.ToSpecification()
	if err != nil {
		h.HandleError(c, errors.NewValidationError(err.Error(), nil))
		return
	}

	h.ApplyVisibilityFilters(c, &spec)

	if err := h.ARValidator.ValidateSpec(spec); err != nil {
		h.HandleError(c, err)
		return
	}

	result, err := h.ARQuery.GetByParentID(c.Request.Context(), spec, annotationID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	pagination := &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	reviews := make([]response.AnnotationReviewResponse, len(result.Data))
	for i, r := range result.Data {
		reviews[i] = *response.NewAnnotationReviewResponse(r)
	}

	h.Response.SuccessList(c, reviews, pagination)
}

// Update godoc
// @Summary Update an annotation review
// @Description Update the status, comments, or modified polygon/value of an annotation review. Can only be done by the reviewer who created it.
// @Tags Annotation Reviews
// @Accept json
// @Produce json
// @Param id path string true "Review ID"
// @Param request body request.UpdateAnnotationReviewRequest true "Annotation review update request"
// @Success 200 {object} response.AnnotationReviewDataResponse "Review updated successfully"
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /annotation-reviews/{id} [put]
func (h *AnnotationReviewHandler) Update(c *gin.Context) {
	reviewID := c.Param("id")
	if reviewID == "" {
		h.HandleError(c, errors.NewValidationError("review ID is required", nil))
		return
	}

	requesterID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	var req request.UpdateAnnotationReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	var modifiedPolygon *[]command.CommandPoint
	if req.ModifiedPolygon != nil && len(*req.ModifiedPolygon) > 0 {
		points := make([]command.CommandPoint, len(*req.ModifiedPolygon))
		for i, p := range *req.ModifiedPolygon {
			points[i] = command.CommandPoint{X: p.X, Y: p.Y}
		}
		modifiedPolygon = &points
	}

	var status *fields.ReviewStatusField
	if req.Status != nil {
		s := fields.ReviewStatusField(*req.Status)
		status = &s
	}

	cmd := command.UpdateAnnotationReviewCommand{
		UpdateEntityCommand: command.UpdateEntityCommand{
			ID: reviewID,
		},
		RequesterID:     requesterID,
		Status:          status,
		Comments:        req.Comments,
		ModifiedValue:   req.ModifiedValue,
		ModifiedPolygon: modifiedPolygon,
	}

	if errDetails, ok := cmd.Validate(); !ok {
		h.HandleError(c, errors.NewValidationError("invalid command payload", errDetails))
		return
	}

	if err := h.ARUseCase.Update(c.Request.Context(), cmd); err != nil {
		h.HandleError(c, err)
		return
	}

	updatedReview, err := h.ARQuery.Get(c.Request.Context(), reviewID)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	h.Response.Success(c, http.StatusOK, response.NewAnnotationReviewResponse(updatedReview))
}
