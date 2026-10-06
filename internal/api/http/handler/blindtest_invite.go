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
	appusecase "github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

// ─────────────────────────────────────────────────────────────────────────────
// Admin: invitation links of a blind test
// ─────────────────────────────────────────────────────────────────────────────

type BlindTestInviteHandler struct {
	helper.BaseHandler
	Invites port.BlindTestInviteUseCase
}

func NewBlindTestInviteHandler(invites port.BlindTestInviteUseCase, logger *slog.Logger) *BlindTestInviteHandler {
	return &BlindTestInviteHandler{Invites: invites, BaseHandler: helper.NewBaseHandler(logger)}
}

// Create godoc
// @Summary Make a shared invitation link to a blind test
// @Description Admins only. People who open /kor-test/katil/{token} join with a name and a 4-digit PIN, without an account, up to max_participants.
// @Tags Blind Tests
// @Accept json
// @Produce json
// @Param id path string true "Blind test ID"
// @Param request body request.CreateBlindTestInviteRequest true "Cap and optional expiry"
// @Success 201 {object} response.BlindTestInviteResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/invites [post]
func (h *BlindTestInviteHandler) Create(c *gin.Context) {
	adminID, err := middleware.GetAuthenticatedUserID(c)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	var req request.CreateBlindTestInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	inv, err := h.Invites.Create(c.Request.Context(), c.Param("id"), adminID, req.MaxParticipants, req.ExpiresAt)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusCreated, response.NewBlindTestInviteResponse(inv))
}

// List godoc
// @Summary List the invitation links of a blind test
// @Description Admins only. Newest first, with how many people joined each.
// @Tags Blind Tests
// @Produce json
// @Param id path string true "Blind test ID"
// @Success 200 {array} response.BlindTestInviteResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/invites [get]
func (h *BlindTestInviteHandler) List(c *gin.Context) {
	list, err := h.Invites.List(c.Request.Context(), c.Param("id"))
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestInviteResponses(list))
}

// Update godoc
// @Summary Open / close an invitation link or change its cap
// @Description Admins only. Closing stops everyone, joined or not; the cap cannot go below the people who joined.
// @Tags Blind Tests
// @Accept json
// @Produce json
// @Param id path string true "Blind test ID"
// @Param invite_id path string true "Invitation ID"
// @Param request body request.UpdateBlindTestInviteRequest true "Fields to change"
// @Success 200 {object} response.BlindTestInviteResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /blind-tests/{id}/invites/{invite_id} [put]
func (h *BlindTestInviteHandler) Update(c *gin.Context) {
	var req request.UpdateBlindTestInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	inv, err := h.Invites.Update(c.Request.Context(), c.Param("id"), c.Param("invite_id"), req.Active, req.MaxParticipants)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestInviteResponse(inv))
}

// ─────────────────────────────────────────────────────────────────────────────
// Public: people taking a blind test through an invitation link
// ─────────────────────────────────────────────────────────────────────────────

// BlindTestGuestHandler serves /public/blind-tests/invites/{token}/...: no
// platform user; the link token picks the invitation and, for the test itself,
// the X-Guest-Session header picks the person. Guests take the test exactly as
// users do, under user id guest_{id}; they never get results or the key.
type BlindTestGuestHandler struct {
	helper.BaseHandler
	Invites port.BlindTestInviteUseCase
	Tests   port.BlindTestUseCase
}

func NewBlindTestGuestHandler(invites port.BlindTestInviteUseCase, tests port.BlindTestUseCase, logger *slog.Logger) *BlindTestGuestHandler {
	return &BlindTestGuestHandler{Invites: invites, Tests: tests, BaseHandler: helper.NewBaseHandler(logger)}
}

const guestContextKey = "blind_test_guest"

// Info godoc
// @Summary What an invitation link leads to
// @Description Public. The set's name, how many people joined and whether the link takes new people.
// @Tags Blind Test Invitations
// @Produce json
// @Param token path string true "Invitation link token"
// @Success 200 {object} response.BlindTestInviteInfoResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /public/blind-tests/invites/{token} [get]
func (h *BlindTestGuestHandler) Info(c *gin.Context) {
	info, err := h.Invites.Info(c.Request.Context(), c.Param("token"))
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestInviteInfoResponse(info))
}

// Join godoc
// @Summary Join a blind test through an invitation link
// @Description Public. Name (unique in the link; kept in capitals, Turkish letters in ASCII), 4-digit PIN, consent. details.code on refusal: name_invalid, pin_invalid, consent_required, name_taken, invite_full, invite_closed, invite_expired.
// @Tags Blind Test Invitations
// @Accept json
// @Produce json
// @Param token path string true "Invitation link token"
// @Param request body request.BlindTestJoinRequest true "Participant"
// @Success 201 {object} response.BlindTestGuestSessionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse
// @Router /public/blind-tests/invites/{token}/join [post]
func (h *BlindTestGuestHandler) Join(c *gin.Context) {
	var req request.BlindTestJoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	s, err := h.Invites.Join(c.Request.Context(), c.Param("token"),
		port.BlindTestJoin{Name: req.Name, PIN: req.PIN, Consent: req.Consent})
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusCreated, response.BlindTestGuestSessionResponse{SessionToken: s.Token, Name: s.Guest.Name})
}

// Resume godoc
// @Summary Come back to a blind test with name and PIN
// @Description Public. details.code on refusal: wrong_credentials (with attempts_left), locked (with locked_until), invite_closed.
// @Tags Blind Test Invitations
// @Accept json
// @Produce json
// @Param token path string true "Invitation link token"
// @Param request body request.BlindTestResumeRequest true "Name and PIN"
// @Success 200 {object} response.BlindTestGuestSessionResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /public/blind-tests/invites/{token}/resume [post]
func (h *BlindTestGuestHandler) Resume(c *gin.Context) {
	var req request.BlindTestResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	s, err := h.Invites.Resume(c.Request.Context(), c.Param("token"), req.Name, req.PIN)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.BlindTestGuestSessionResponse{SessionToken: s.Token, Name: s.Guest.Name})
}

// RequireGuest authenticates the X-Guest-Session header against the link.
func (h *BlindTestGuestHandler) RequireGuest(c *gin.Context) {
	g, err := h.Invites.Authenticate(c.Request.Context(), c.Param("token"), c.GetHeader("X-Guest-Session"))
	if err != nil {
		h.HandleError(c, err)
		c.Abort()
		return
	}
	c.Set(guestContextKey, g)
	c.Next()
}

func guestOf(c *gin.Context) (setID, userID string) {
	g := c.MustGet(guestContextKey).(*port.BlindTestGuest)
	return g.SetID, appusecase.BlindTestGuestUserID(*g)
}

// Get godoc
// @Summary The guest's test
// @Description Header X-Guest-Session. Image ids in the guest's own order, their answers and notes; no labels.
// @Tags Blind Test Invitations
// @Produce json
// @Param token path string true "Invitation link token"
// @Param X-Guest-Session header string true "Session token from join / resume"
// @Success 200 {object} response.BlindTestResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /public/blind-tests/invites/{token}/test [get]
func (h *BlindTestGuestHandler) Get(c *gin.Context) {
	setID, userID := guestOf(c)
	view, err := h.Tests.Get(c.Request.Context(), setID, userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestResponse(view))
}

// Image godoc
// @Summary An image of the guest's test
// @Tags Blind Test Invitations
// @Produce png
// @Param token path string true "Invitation link token"
// @Param image_id path string true "Image ID"
// @Param X-Guest-Session header string true "Session token"
// @Success 200 {file} binary "PNG image"
// @Router /public/blind-tests/invites/{token}/test/images/{image_id} [get]
func (h *BlindTestGuestHandler) Image(c *gin.Context) {
	setID, _ := guestOf(c)
	reader, err := h.Tests.OpenImage(c.Request.Context(), setID, c.Param("image_id"))
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

// Answer godoc
// @Summary Answer one image (guest)
// @Tags Blind Test Invitations
// @Accept json
// @Produce json
// @Param token path string true "Invitation link token"
// @Param image_id path string true "Image ID"
// @Param X-Guest-Session header string true "Session token"
// @Param request body request.BlindTestAnswerRequest true "real or synthetic"
// @Success 200 {object} response.BlindTestProgressResponse
// @Router /public/blind-tests/invites/{token}/test/answers/{image_id} [put]
func (h *BlindTestGuestHandler) Answer(c *gin.Context) {
	var req request.BlindTestAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	setID, userID := guestOf(c)
	resp, err := h.Tests.Answer(c.Request.Context(), setID, userID, appusecase.BlindTestGuestRole, c.Param("image_id"), req.Label)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestProgressResponse(resp))
}

// Note godoc
// @Summary Write a note on one image (guest)
// @Tags Blind Test Invitations
// @Accept json
// @Produce json
// @Param token path string true "Invitation link token"
// @Param image_id path string true "Image ID"
// @Param X-Guest-Session header string true "Session token"
// @Param request body request.BlindTestNoteRequest true "Note text"
// @Success 200 {object} response.BlindTestProgressResponse
// @Router /public/blind-tests/invites/{token}/test/notes/{image_id} [put]
func (h *BlindTestGuestHandler) Note(c *gin.Context) {
	var req request.BlindTestNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.HandleError(c, errors.NewValidationError("invalid request payload", map[string]interface{}{"error": err.Error()}))
		return
	}
	setID, userID := guestOf(c)
	resp, err := h.Tests.Note(c.Request.Context(), setID, userID, appusecase.BlindTestGuestRole, c.Param("image_id"), req.Note)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestProgressResponse(resp))
}

// Complete godoc
// @Summary Complete the guest's test
// @Tags Blind Test Invitations
// @Produce json
// @Param token path string true "Invitation link token"
// @Param X-Guest-Session header string true "Session token"
// @Success 200 {object} response.BlindTestProgressResponse
// @Router /public/blind-tests/invites/{token}/test/complete [post]
func (h *BlindTestGuestHandler) Complete(c *gin.Context) {
	setID, userID := guestOf(c)
	resp, err := h.Tests.Complete(c.Request.Context(), setID, userID)
	if err != nil {
		h.HandleError(c, err)
		return
	}
	h.Response.Success(c, http.StatusOK, response.NewBlindTestProgressResponse(resp))
}
