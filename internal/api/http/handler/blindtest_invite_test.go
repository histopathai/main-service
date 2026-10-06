package handler_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/handler"
	"github.com/histopathai/main-service/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type guestInvites struct{ port.BlindTestInviteUseCase }

func (guestInvites) Authenticate(context.Context, string, string) (*port.BlindTestGuest, error) {
	return &port.BlindTestGuest{ID: "g1", SetID: "s1"}, nil
}

func (guestInvites) Info(context.Context, string) (*port.BlindTestInviteInfo, error) {
	return &port.BlindTestInviteInfo{SetName: "Set A", Images: 200, Max: 10, Joinable: true}, nil
}

type guestTests struct{ port.BlindTestUseCase }

func (guestTests) Get(_ context.Context, setID, _ string) (*port.BlindTestView, error) {
	return &port.BlindTestView{Set: port.BlindTestSet{ID: setID, Name: "Set A", Active: true, ImageIDs: []string{"a"},
		Description: "sd15-unet baseline, step_0130000"}}, nil
}

func TestBlindTestGuest_SeesNoSetDescription(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := handler.NewBlindTestGuestHandler(guestInvites{}, guestTests{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	engine := gin.New()
	engine.GET("/invites/:token", h.Info)
	engine.GET("/invites/:token/test", h.RequireGuest, h.Get)

	for _, path := range []string{"/invites/tok", "/invites/tok/test"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("X-Guest-Session", "g1.secret")
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, path)
		body := w.Body.String()
		assert.Contains(t, body, "Set A", path)
		assert.False(t, strings.Contains(body, "sd15") || strings.Contains(body, "step_"), "%s leaks: %s", path, body)
	}
}
