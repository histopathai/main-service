package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/api/http/dto/response"
	"github.com/histopathai/main-service/internal/api/http/handler/helper"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

// RecheckHandler serves Ek Kontrol: images an admin sent back to their expert,
// with the reason. Admins send and cancel; the expert marks a request done.
type RecheckHandler struct {
	helper.BaseHandler
	UseCase port.RecheckUseCase
}

func NewRecheckHandler(useCase port.RecheckUseCase, logger *slog.Logger) *RecheckHandler {
	return &RecheckHandler{UseCase: useCase, BaseHandler: helper.NewBaseHandler(logger)}
}

// List godoc
// @Summary List the Ek Kontrol requests
// @Description Any user group. By workspace, then image name.
// @Tags Recheck
// @Produce json
// @Param status query string false "open or done; all when empty"
// @Success 200 {array} response.RecheckResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /recheck-requests [get]
func (h *RecheckHandler) List(c *gin.Context) {
	list, err := h.UseCase.List(c.Request.Context(), c.Query("status"))
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewRecheckResponses(list))
}

// Request godoc
// @Summary Send an image to Ek Kontrol
// @Description Admins only. Adds the reason to the image's request (made if there is none, reopened if done); the same reason again replaces its note.
// @Tags Recheck
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param request body request.RecheckRequestRequest true "Reason and optional note"
// @Success 200 {object} response.RecheckResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /recheck-requests/{image_id}/reasons [post]
func (h *RecheckHandler) Request(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	var req request.RecheckRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	r, err := h.UseCase.Request(c.Request.Context(), c.Param("image_id"), userID, req.Reason, req.Note)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewRecheckResponse(r))
}

// SetStatus godoc
// @Summary Mark an Ek Kontrol request done, or open again
// @Description Admins and pathologists. Done needs the outcome (corrected, no_change, undecided, unsuitable); no_change and undecided need a note saying why.
// @Tags Recheck
// @Accept json
// @Produce json
// @Param image_id path string true "Image ID"
// @Param request body request.RecheckStatusRequest true "done: true or false"
// @Success 200 {object} response.RecheckResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /recheck-requests/{image_id}/status [put]
func (h *RecheckHandler) SetStatus(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	var req request.RecheckStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	r, err := h.UseCase.SetDone(c.Request.Context(), c.Param("image_id"), userID, *req.Done, req.Outcome, req.Note)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewRecheckResponse(r))
}

// Cancel godoc
// @Summary Take an image out of Ek Kontrol
// @Description Admins only. Removes the request with all its reasons; the image and its annotations are not touched.
// @Tags Recheck
// @Param image_id path string true "Image ID"
// @Success 204
// @Failure 403 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /recheck-requests/{image_id} [delete]
func (h *RecheckHandler) Cancel(c *gin.Context) {
	if err := h.UseCase.Cancel(c.Request.Context(), c.Param("image_id")); err != nil {
		h.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// RequestWorkspace godoc
// @Summary Send every image of a workspace to Ek Kontrol
// @Description Admins only. Adds the reason "dataset" with the note to each image (made, replaced or reopened as needed).
// @Tags Recheck
// @Accept json
// @Produce json
// @Param ws_id path string true "Workspace ID"
// @Param request body request.RecheckWorkspaceRequest true "Why the workspace is sent"
// @Success 200 {object} response.RecheckWorkspaceResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse "The workspace has no images"
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /recheck-workspaces/{ws_id} [post]
func (h *RecheckHandler) RequestWorkspace(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	var req request.RecheckWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	n, err := h.UseCase.RequestWorkspace(c.Request.Context(), c.Param("ws_id"), userID, req.Note)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.RecheckWorkspaceResponse{WsID: c.Param("ws_id"), Images: n})
}

// WithdrawWorkspace godoc
// @Summary Take a workspace out of Ek Kontrol
// @Description Admins only. Removes the reason "dataset" from its images; images with other reasons stay listed.
// @Tags Recheck
// @Produce json
// @Param ws_id path string true "Workspace ID"
// @Success 200 {object} response.RecheckWorkspaceResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /recheck-workspaces/{ws_id} [delete]
func (h *RecheckHandler) WithdrawWorkspace(c *gin.Context) {
	n, err := h.UseCase.WithdrawWorkspace(c.Request.Context(), c.Param("ws_id"))
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.RecheckWorkspaceResponse{WsID: c.Param("ws_id"), Images: n})
}
