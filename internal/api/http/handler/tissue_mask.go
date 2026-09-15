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

// TissueMaskHandler serves the admin-only tissue mask API. Masks are keyed by
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
// @Description Admin only. Polygons are in level-0 pixels; the preview they were computed on is served at /proxy/{image_id}/tissue_preview.png.
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
// @Description Admin only. Replaces the mask of an image and sets its status to edited; a previous approval is cleared.
// @Tags Tissue Masks
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param request body request.SaveTissueMaskRequest true "Complete tissue mask"
// @Success 200 {object} response.TissueMaskDataResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
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
		ImageID: c.Param("image_id"),
		UserID:  userID,
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
// @Description Admin only. Marks the current polygons as reviewed. Approving an approved mask is a no-op.
// @Tags Tissue Masks
// @Produce json
// @Param image_id path string true "Image ID"
// @Success 200 {object} response.TissueMaskDataResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /tissue-masks/image/{image_id}/approve [post]
func (h *TissueMaskHandler) Approve(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	mask, err := h.TMUseCase.Approve(c.Request.Context(), c.Param("image_id"), userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewTissueMaskResponse(mask))
}

// GetByWorkspaceID godoc
// @Summary List tissue masks of a workspace
// @Description Admin only. Returns summaries without polygons, e.g. to show review status next to images.
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
