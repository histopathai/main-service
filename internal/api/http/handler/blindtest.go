package handler

import (
	"io"
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

// BlindTestHandler serves the blind test: participants tell real image patches
// from synthetic ones. Every user group may take a test; only admins get the
// results, which carry the answer key.
type BlindTestHandler struct {
	helper.BaseHandler
	UseCase port.BlindTestUseCase
}

func NewBlindTestHandler(useCase port.BlindTestUseCase, logger *slog.Logger) *BlindTestHandler {
	return &BlindTestHandler{UseCase: useCase, BaseHandler: helper.NewBaseHandler(logger)}
}

// List godoc
// @Summary List the active blind tests
// @Description Any user group. With the caller's own progress; no image ids, no labels.
// @Tags Blind Tests
// @Produce json
// @Success 200 {array} response.BlindTestSummaryResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests [get]
func (h *BlindTestHandler) List(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	list, err := h.UseCase.List(c.Request.Context(), userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestSummaryResponses(list))
}

// Get godoc
// @Summary Get a blind test to take
// @Description Any user group. The image ids in the caller's own order and the caller's answers; no labels.
// @Tags Blind Tests
// @Produce json
// @Param id path string true "Blind test ID"
// @Success 200 {object} response.BlindTestResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id} [get]
func (h *BlindTestHandler) Get(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	view, err := h.UseCase.Get(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestResponse(view))
}

// Answer godoc
// @Summary Answer one image of a blind test
// @Description Any user group. Records or changes the caller's answer; refused once the test is completed.
// @Tags Blind Tests
// @Accept json
// @Produce json
// @Param id path string true "Blind test ID"
// @Param image_id path string true "Image ID"
// @Param request body request.BlindTestAnswerRequest true "real or synthetic"
// @Success 200 {object} response.BlindTestProgressResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "The test is already completed"
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/answers/{image_id} [put]
func (h *BlindTestHandler) Answer(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	role, err := middleware.GetAuthenticatedUserRole(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	var req request.BlindTestAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	resp, err := h.UseCase.Answer(c.Request.Context(), c.Param("id"), userID, role, c.Param("image_id"), req.Label)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestProgressResponse(resp))
}

// Note godoc
// @Summary Write a note on one image of a blind test
// @Description Any user group. Why the caller took the image for real or synthetic; empty removes the note. Allowed after the test is completed too (only the answers are locked).
// @Tags Blind Tests
// @Accept json
// @Produce json
// @Param id path string true "Blind test ID"
// @Param image_id path string true "Image ID"
// @Param request body request.BlindTestNoteRequest true "Note text (at most 2000 characters)"
// @Success 200 {object} response.BlindTestProgressResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/notes/{image_id} [put]
func (h *BlindTestHandler) Note(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	role, err := middleware.GetAuthenticatedUserRole(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	var req request.BlindTestNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	resp, err := h.UseCase.Note(c.Request.Context(), c.Param("id"), userID, role, c.Param("image_id"), req.Note)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestProgressResponse(resp))
}

// Complete godoc
// @Summary Complete a blind test
// @Description Any user group. Locks the caller's answers; every image must be answered.
// @Tags Blind Tests
// @Produce json
// @Param id path string true "Blind test ID"
// @Success 200 {object} response.BlindTestProgressResponse
// @Failure 400 {object} response.ErrorResponse "Not every image is answered"
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/complete [post]
func (h *BlindTestHandler) Complete(c *gin.Context) {
	userID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	resp, err := h.UseCase.Complete(c.Request.Context(), c.Param("id"), userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestProgressResponse(resp))
}

// Image godoc
// @Summary Get an image of a blind test
// @Description Any user group. PNG; only ids that belong to the set are served.
// @Tags Blind Tests
// @Produce png
// @Param id path string true "Blind test ID"
// @Param image_id path string true "Image ID"
// @Success 200 {file} binary "PNG image"
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/images/{image_id} [get]
func (h *BlindTestHandler) Image(c *gin.Context) {
	reader, err := h.UseCase.OpenImage(c.Request.Context(), c.Param("id"), c.Param("image_id"))
	if err != nil {
		h.HandleError(c, err)
		return
	}
	defer reader.Close()
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "private, max-age=86400")
	c.Status(http.StatusOK)
	_, _ = io.Copy(c.Writer, reader)
}

// Results godoc
// @Summary Get the results of a blind test
// @Description Admins only. The answer key, every participant's score and the per-image votes.
// @Tags Blind Tests
// @Produce json
// @Param id path string true "Blind test ID"
// @Success 200 {object} response.BlindTestResultsResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/results [get]
func (h *BlindTestHandler) Results(c *gin.Context) {
	res, err := h.UseCase.Results(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestResultsResponse(res))
}
