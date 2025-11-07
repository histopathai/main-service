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

func setupPatientHandler(t *testing.T) (*gin.Engine, *mocks.MockIPatientService) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockIPatientService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	v := validator.NewRequestValidator()

	ph := handler.NewPatientHandler(mockService, v, logger)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authenticated_user_id", "test-user-id")
		c.Set("request_id", "test-request-id")
		c.Next()
	})

	r.POST("/patients", ph.CreateNewPatient)
	r.GET("/patients/:patient_id", ph.GetPatientByID)
	r.POST("/patients/:patient_id/transfer/:workspace_id", ph.TransferPatientWorkspace)

	return r, mockService
}

func TestPatientHandler_CreateNewPatient_Success(t *testing.T) {
	r, mockService := setupPatientHandler(t)

	createReq := request.CreatePatientRequest{
		Name:        "John Doe",
		WorkspaceID: "ws-123",
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().
		CreateNewPatient(gomock.Any(), gomock.Any()).
		Return(&model.Patient{
			ID:          "pat-123",
			Name:        createReq.Name,
			WorkspaceID: createReq.WorkspaceID,
		}, nil)

	req, _ := http.NewRequest(http.MethodPost, "/patients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "pat-123")
}

func TestPatientHandler_CreateNewPatient_MissingFields(t *testing.T) {
	r, _ := setupPatientHandler(t)

	createReq := request.CreatePatientRequest{
		Name: "John Doe",
	}
	body, _ := json.Marshal(createReq)

	req, _ := http.NewRequest(http.MethodPost, "/patients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "CreatePatientRequest.WorkspaceID")
}

func TestPatientHandler_CreateNewPatient_ServiceError(t *testing.T) {
	r, mockService := setupPatientHandler(t)

	createReq := request.CreatePatientRequest{
		Name:        "John Doe",
		WorkspaceID: "ws-123",
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().
		CreateNewPatient(gomock.Any(), gomock.Any()).
		Return(nil, errors.NewConflictError("name exists", nil))

	req, _ := http.NewRequest(http.MethodPost, "/patients", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestPatientHandler_TransferPatientWorkspace_Success(t *testing.T) {
	r, mockService := setupPatientHandler(t)

	mockService.EXPECT().
		TransferPatientWorkspace(gomock.Any(), "pat-123", "ws-456").
		Return(nil)

	req, _ := http.NewRequest(http.MethodPost, "/patients/pat-123/transfer/ws-456", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}
