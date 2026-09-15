package handler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/middleware"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTissueMaskUseCase struct {
	port.TissueMaskUseCase
	last command.ReviewTissueMaskCommand
	err  error
}

func (f *fakeTissueMaskUseCase) review(cmd command.ReviewTissueMaskCommand, status vobj.TissueMaskStatus) (*model.TissueMask, error) {
	f.last = cmd
	if f.err != nil {
		return nil, f.err
	}
	return &model.TissueMask{Entity: vobj.Entity{ID: cmd.ImageID}, Status: status, Revision: 7, RejectReason: cmd.Reason}, nil
}

func (f *fakeTissueMaskUseCase) Approve(_ context.Context, cmd command.ReviewTissueMaskCommand) (*model.TissueMask, error) {
	return f.review(cmd, vobj.TissueMaskStatusApproved)
}

func (f *fakeTissueMaskUseCase) Reject(_ context.Context, cmd command.ReviewTissueMaskCommand) (*model.TissueMask, error) {
	return f.review(cmd, vobj.TissueMaskStatusRejected)
}

func tissueRouter(uc port.TissueMaskUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h := NewTissueMaskHandler(nil, uc, logger)
	auth := middleware.NewAuthMiddleware(logger)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware(), auth.RequireAuth(), auth.RequireRole("admin"))
	r.POST("/tissue-masks/image/:image_id/approve", h.Approve)
	r.POST("/tissue-masks/image/:image_id/reject", h.Reject)
	return r
}

func post(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(http.MethodPost, path, reader)
	req.Header.Set("X-User-ID", "u1")
	req.Header.Set("X-User-Role", "admin")
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestTissueMaskReviewEndpoints(t *testing.T) {
	uc := &fakeTissueMaskUseCase{}
	r := tissueRouter(uc)

	w := post(r, "/tissue-masks/image/img-1/approve", "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	assert.Nil(t, uc.last.ExpectedRevision, "approve without a body skips the revision check")
	assert.Equal(t, "u1", uc.last.UserID)

	w = post(r, "/tissue-masks/image/img-1/reject", `{"reason":"IHC","expected_revision":6}`)
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	require.NotNil(t, uc.last.ExpectedRevision)
	assert.Equal(t, 6, *uc.last.ExpectedRevision)
	var body struct {
		Data struct {
			Status       string `json:"status"`
			Revision     int    `json:"revision"`
			RejectReason string `json:"reject_reason"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "rejected", body.Data.Status)
	assert.Equal(t, 7, body.Data.Revision)
	assert.Equal(t, "IHC", body.Data.RejectReason)

	uc.err = errors.NewConflictError("tissue mask was changed by someone else; reload it", map[string]interface{}{"current_revision": 8})
	w = post(r, "/tissue-masks/image/img-1/approve", `{"expected_revision":6}`)
	assert.Equal(t, http.StatusConflict, w.Code)

	w = post(r, "/tissue-masks/image/img-1/reject", `{"reason":`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
