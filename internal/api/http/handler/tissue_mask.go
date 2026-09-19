package handler

import (
	"context"
	stderrors "errors"
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
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	validator "github.com/histopathai/main-service/internal/shared/query"
)

// TissueMaskHandler serves the tissue mask API, read-write for every user group. Masks are keyed by
// image ID.
type TissueMaskHandler struct {
	helper.BaseHandler
	TMQuery     port.TissueMaskQuery
	TMUseCase   port.TissueMaskUseCase
	TMValidator *validator.Validator
}

func NewTissueMaskHandler(query port.TissueMaskQuery, useCase port.TissueMaskUseCase, logger *slog.Logger) *TissueMaskHandler {
	return &TissueMaskHandler{
		TMQuery:     query,
		TMUseCase:   useCase,
		TMValidator: validator.NewValidator(fields.NewTissueMaskFieldSet()),
		BaseHandler: helper.NewBaseHandler(logger),
	}
}

// GetByImageID godoc
// @Summary Get the tissue mask of an image
// @Description Any user group (admin, pathologist, datascientist). Polygons are in level-0 pixels; the preview they were computed on is served at /proxy/{image_id}/tissue_preview.png.
// @Tags Tissue Masks
// @Produce json
// @Param image_id path string true "Image ID"
// @Success 200 {object} response.TissueMaskDataResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /tissue-masks/image/{image_id} [get]
func (h *TissueMaskHandler) GetByImageID(c *gin.Context) {
	mask, err := h.TMQuery.Get(c.Request.Context(), c.Param("image_id"))
	if err != nil {
		h.HandleError(c, err)
		return
	}
	if mask == nil || mask.Deleted {
		h.HandleError(c, errors.NewNotFoundError("tissue mask not found"))
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewTissueMaskResponse(mask))
}

// Save godoc
// @Summary Save an edited tissue mask
// @Description Any user group (admin, pathologist, datascientist). Replaces the mask of an image and sets its status to edited; a previous approval is cleared.
// @Tags Tissue Masks
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param request body request.SaveTissueMaskRequest true "Complete tissue mask"
// @Success 200 {object} response.TissueMaskDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "The mask changed since expected_revision"
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /tissue-masks/image/{image_id} [put]
func (h *TissueMaskHandler) Save(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}

	var req request.SaveTissueMaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
			"error": err.Error(),
		}))
		return
	}

	toPoints := func(pts []request.TissuePointRequest) []vobj.Point {
		out := make([]vobj.Point, len(pts))
		for i, p := range pts {
			out[i] = vobj.Point{X: p.X, Y: p.Y}
		}
		return out
	}
	polygons := make([]vobj.TissuePolygon, len(req.Polygons))
	for i, p := range req.Polygons {
		holes := make([][]vobj.Point, len(p.Holes))
		for j, hole := range p.Holes {
			holes[j] = toPoints(hole)
		}
		polygons[i] = vobj.TissuePolygon{Exterior: toPoints(p.Exterior), Holes: holes}
	}

	cmd := command.SaveTissueMaskCommand{
		ImageID:          c.Param("image_id"),
		UserID:           userID,
		ExpectedRevision: req.ExpectedRevision,
		TissueMaskData: command.TissueMaskData{
			AlgorithmVersion: req.AlgorithmVersion,
			Params: vobj.TissueParams{
				Method:              req.Params.Method,
				SaturationThreshold: req.Params.SaturationThreshold,
				GrayThreshold:       req.Params.GrayThreshold,
				ClosingRadius:       req.Params.ClosingRadius,
				MinObjectArea:       req.Params.MinObjectArea,
				MinHoleArea:         req.Params.MinHoleArea,
				SimplifyTolerance:   req.Params.SimplifyTolerance,
			},
			Polygons:        polygons,
			PreviewWidth:    req.PreviewWidth,
			PreviewHeight:   req.PreviewHeight,
			Level0Width:     req.Level0Width,
			Level0Height:    req.Level0Height,
			DownsampleX:     req.DownsampleX,
			DownsampleY:     req.DownsampleY,
			TissueAreaRatio: req.TissueAreaRatio,
		},
	}

	mask, err := h.TMUseCase.Save(c.Request.Context(), cmd)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewTissueMaskResponse(mask))
}

// Approve godoc
// @Summary Approve a tissue mask
// @Description Any user group (admin, pathologist, datascientist). Marks the current polygons as reviewed and clears a rejection. Approving an approved mask is a no-op.
// @Tags Tissue Masks
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param request body request.ReviewTissueMaskRequest false "Optional expected revision"
// @Success 200 {object} response.TissueMaskDataResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "The mask changed since expected_revision"
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /tissue-masks/image/{image_id}/approve [post]
func (h *TissueMaskHandler) Approve(c *gin.Context) {
	h.review(c, h.TMUseCase.Approve)
}

// Reject godoc
// @Summary Reject a tissue mask
// @Description Any user group (admin, pathologist, datascientist). Marks the image as unusable for tissue-based work (e.g. no tissue, failed stain) with an optional reason and clears an approval. The worker never overwrites a rejected mask; saving the mask again makes it edited.
// @Tags Tissue Masks
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param request body request.ReviewTissueMaskRequest false "Optional reason and expected revision"
// @Success 200 {object} response.TissueMaskDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "The mask changed since expected_revision"
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /tissue-masks/image/{image_id}/reject [post]
func (h *TissueMaskHandler) Reject(c *gin.Context) {
	h.review(c, h.TMUseCase.Reject)
}

func (h *TissueMaskHandler) review(c *gin.Context, apply func(context.Context, command.ReviewTissueMaskCommand) (*model.TissueMask, error)) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	var req request.ReviewTissueMaskRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil && !stderrors.Is(err, io.EOF) {
			h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{
				"error": err.Error(),
			}))
			return
		}
	}
	mask, err := apply(c.Request.Context(), command.ReviewTissueMaskCommand{
		ImageID:          c.Param("image_id"),
		UserID:           userID,
		ExpectedRevision: req.ExpectedRevision,
		Reason:           req.Reason,
	})
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewTissueMaskResponse(mask))
}

// GetByWorkspaceID godoc
// @Summary List tissue masks of a workspace
// @Description Any user group (admin, pathologist, datascientist). Returns summaries without polygons, e.g. to show review status next to images.
// @Tags Tissue Masks
// @Produce json
// @Param workspace_id path string true "Workspace ID"
// @Param limit query int false "Number of items per page" default(20) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param sort_by query string false "Field to sort by" default(created_at) Enums(created_at, updated_at, status, tissue_area_ratio)
// @Param sort_dir query string false "Sort direction" default(desc) Enums(asc, desc)
// @Success 200 {object} response.TissueMaskListResponseDoc
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /tissue-masks/workspace/{workspace_id} [get]
// GetWorkspaceStats godoc
// @Summary Per-workspace tissue mask completion stats
// @Description Any user group (admin, pathologist, datascientist). Returns a compact table: per workspace, how many images need masks and how many are done (approved/rejected) vs remaining.
// @Tags Tissue Masks
// @Produce json
// @Success 200 {object} response.TissueMaskWorkspaceStatsListResponseDoc
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /tissue-masks/workspace-stats [get]
func (h *TissueMaskHandler) GetWorkspaceStats(c *gin.Context) {
	stats, err := h.TMQuery.GetWorkspaceStats(c.Request.Context())
	if err != nil {
		h.HandleError(c, err)
		return
	}
	result := make([]response.TissueMaskWorkspaceStatsResponse, len(stats))
	for i, s := range stats {
		result[i] = response.TissueMaskWorkspaceStatsResponse{
			WorkspaceID:   s.WorkspaceID,
			WorkspaceName: s.WorkspaceName,
			TotalImages:   s.TotalImages,
			Approved:      s.Approved,
			Edited:        s.Edited,
			Auto:          s.Auto,
			Rejected:      s.Rejected,
			Missing:       s.Missing,
			Done:          s.Done,
			Remaining:     s.Remaining,
		}
	}
	h.Response.Success(c, http.StatusOK, result)
}

func (h *TissueMaskHandler) GetByWorkspaceID(c *gin.Context) {
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
	if err := h.TMValidator.ValidateSpec(spec); err != nil {
		h.HandleError(c, err)
		return
	}

	result, err := h.TMQuery.GetByWsID(c.Request.Context(), spec, c.Param("workspace_id"))
	if err != nil {
		h.HandleError(c, err)
		return
	}

	masks := make([]response.TissueMaskSummaryResponse, len(result.Data))
	for i, m := range result.Data {
		masks[i] = response.NewTissueMaskSummaryResponse(m)
	}
	h.Response.SuccessList(c, masks, &response.PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	})
}
