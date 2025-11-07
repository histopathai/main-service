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
	"github.com/histopathai/main-service/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupAnnotationTypeHandler(t *testing.T) (*gin.Engine, *mocks.MockIAnnotationTypeService) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockIAnnotationTypeService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	v := validator.NewRequestValidator()

	ath := handler.NewAnnotationTypeHandler(mockService, v, logger)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authenticated_user_id", "test-user-id")
		c.Set("request_id", "test-request-id")
		c.Next()
	})

	r.POST("/annotation-types", ath.CreateNewAnnotationType)
	r.GET("/annotation-types/:id", ath.GetAnnotationType)
	r.GET("/annotation-types/classification-enabled", ath.GetClassificationOptionedAnnotationTypes)
	r.DELETE("/annotation-types/:id", ath.DeleteAnnotationType)

	return r, mockService
}

func TestAnnotationTypeHandler_CreateNewAnnotationType_Success(t *testing.T) {
	r, mockService := setupAnnotationTypeHandler(t)

	createReq := request.CreateAnnotationTypeRequest{
		Name:                  "Test Type",
		ClassificationEnabled: true,
		ClassList:             &[]string{"A", "B"},
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().
		CreateNewAnnotationType(gomock.Any(), gomock.Any()).
		Return(&model.AnnotationType{
			ID:   "at-123",
			Name: createReq.Name,
		}, nil)

	req, _ := http.NewRequest(http.MethodPost, "/annotation-types", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "at-123")
}

func TestAnnotationTypeHandler_CreateNewAnnotationType_ValidationError(t *testing.T) {
	r, mockService := setupAnnotationTypeHandler(t)

	createReq := request.CreateAnnotationTypeRequest{
		Name: "",
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().CreateNewAnnotationType(gomock.Any(), gomock.Any()).Times(0)

	req, _ := http.NewRequest(http.MethodPost, "/annotation-types", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "CreateAnnotationTypeRequest.Name")
}

func TestAnnotationTypeHandler_GetAnnotationType_NotFound(t *testing.T) {
	r, mockService := setupAnnotationTypeHandler(t)

	mockService.EXPECT().
		GetAnnotationTypeByID(gomock.Any(), "at-404").
		Return(nil, errors.NewNotFoundError("not found"))

	req, _ := http.NewRequest(http.MethodGet, "/annotation-types/at-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestAnnotationTypeHandler_DeleteAnnotationType_Conflict(t *testing.T) {
	r, mockService := setupAnnotationTypeHandler(t)

	mockService.EXPECT().
		DeleteAnnotationType(gomock.Any(), "at-in-use").
		Return(errors.NewConflictError("in use", nil))

	req, _ := http.NewRequest(http.MethodDelete, "/annotation-types/at-in-use", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusConflict, w.Code)
}

func TestAnnotationTypeHandler_GetClassificationOptionedAnnotationTypes_Success(t *testing.T) {
	r, mockService := setupAnnotationTypeHandler(t)

	pagination := &query.Pagination{
		Limit:   20,
		Offset:  0,
		SortBy:  "created_at",
		SortDir: "desc",
	}

	mockService.EXPECT().
		GetClassificationAnnotationTypes(gomock.Any(), pagination).
		Return(&query.Result[*model.AnnotationType]{
			Data: []*model.AnnotationType{},
		}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/annotation-types/classification-enabled", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":[]`)
}
