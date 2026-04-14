package handler

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/handler/helper"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type AnnotationReviewHandler struct {
	helper.BaseHandler
	ARUseCase port.AnnotationReviewUseCase
}

func NewAnnotationReviewHandler(useCase port.AnnotationReviewUseCase, logger *slog.Logger) *AnnotationReviewHandler {
	return &AnnotationReviewHandler{
		ARUseCase:   useCase,
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
