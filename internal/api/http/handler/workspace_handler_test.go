package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/histopathai/main-service/internal/api/http/dto/request"
	"github.com/histopathai/main-service/internal/api/http/handler"
	"github.com/histopathai/main-service/internal/api/http/validator"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupWorkspaceHandler(t *testing.T) (*gin.Engine, *mocks.MockIWorkspaceService) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockIWorkspaceService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	v := validator.NewRequestValidator()

	wh := handler.NewWorkspaceHandler(mockService, v, logger)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authenticated_user_id", "test-user-id")
		c.Set("request_id", "test-request-id")
		c.Next()
	})

	r.POST("/workspaces", wh.CreateNewWorkspace)
	r.GET("/workspaces/:id", wh.GetWorkspaceByID)
	r.DELETE("/workspaces/:id", wh.DeleteWorkspace)

	return r, mockService
}

func TestWorkspaceHandler_CreateNewWorkspace_Success(t *testing.T) {
	r, mockService := setupWorkspaceHandler(t)

	createReq := request.CreateWorkspaceRequest{
		Name:         "Test WS",
		OrganType:    "brain",
		Organization: "Org",
		Description:  "Desc",
		License:      "MIT",
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().
		CreateNewWorkspace(gomock.Any(), gomock.Any()).
		Return(&model.Workspace{
			ID:        "ws-123",
			Name:      createReq.Name,
			OrganType: createReq.OrganType,
		}, nil)

	req, _ := http.NewRequest(http.MethodPost, "/workspaces", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "ws-123")
}

func TestWorkspaceHandler_CreateNewWorkspace_InvalidOrganType(t *testing.T) {
	r, mockService := setupWorkspaceHandler(t)

	createReq := request.CreateWorkspaceRequest{
		Name:         "Test WS",
		OrganType:    "invalid-organ",
		Organization: "Org",
		Description:  "Desc",
		License:      "MIT",
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().CreateNewWorkspace(gomock.Any(), gomock.Any()).Times(0)

	req, _ := http.NewRequest(http.MethodPost, "/workspaces", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid organ type")
}

func TestWorkspaceHandler_CreateNewWorkspace_ValidationErrors(t *testing.T) {
	r, _ := setupWorkspaceHandler(t)

	createReq := request.CreateWorkspaceRequest{
		Name: "",
	}
	body, _ := json.Marshal(createReq)

	req, _ := http.NewRequest(http.MethodPost, "/workspaces", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "CreateWorkspaceRequest.Name")
	assert.Contains(t, w.Body.String(), "CreateWorkspaceRequest.OrganType")
}

func TestWorkspaceHandler_GetWorkspaceByID_Success(t *testing.T) {
	r, mockService := setupWorkspaceHandler(t)

	mockService.EXPECT().
		GetWorkspaceByID(gomock.Any(), "ws-123").
		Return(&model.Workspace{ID: "ws-123", Name: "Test WS"}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/workspaces/ws-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ws-123")
}

func TestWorkspaceHandler_GetWorkspaceByID_NotFound(t *testing.T) {
	r, mockService := setupWorkspaceHandler(t)

	mockService.EXPECT().
		GetWorkspaceByID(gomock.Any(), "ws-404").
		Return(nil, errors.NewNotFoundError("not found"))

	req, _ := http.NewRequest(http.MethodGet, "/workspaces/ws-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestWorkspaceHandler_DeleteWorkspace_Conflict(t *testing.T) {
	r, mockService := setupWorkspaceHandler(t)

	mockService.EXPECT().
		DeleteWorkspace(gomock.Any(), "ws-in-use").
		Return(errors.NewConflictError("in use", nil))

	req, _ := http.NewRequest(http.MethodDelete, "/workspaces/ws-in-use", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}
