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

func setupAnnotationHandler(t *testing.T) (*gin.Engine, *mocks.MockIAnnotationService) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockIAnnotationService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	v := validator.NewRequestValidator()

	ah := handler.NewAnnotationHandler(mockService, v, logger)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authenticated_user_id", "test-user-id")
		c.Set("request_id", "test-request-id")
		c.Next()
	})

	r.POST("/annotations", ah.CreateNewAnnotation)
	r.GET("/annotations/:annotation_id", ah.GetAnnotationByID)
	r.GET("/annotations/image/:image_id", ah.GetAnnotationsByImageID)

	return r, mockService
}

func TestAnnotationHandler_CreateNewAnnotation_Success(t *testing.T) {
	r, mockService := setupAnnotationHandler(t)

	score := 0.9
	createReq := request.CreateAnnotationRequest{
		ImageID: "img-123",
		Polygon: []model.Point{{X: 1, Y: 1}},
		Score:   &score,
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().
		CreateNewAnnotation(gomock.Any(), gomock.Any()).
		Return(&model.Annotation{
			ID:      "anno-123",
			ImageID: createReq.ImageID,
			Score:   createReq.Score,
		}, nil)

	req, _ := http.NewRequest(http.MethodPost, "/annotations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "anno-123")
}

func TestAnnotationHandler_CreateNewAnnotation_ValidationError(t *testing.T) {
	r, mockService := setupAnnotationHandler(t)

	createReq := request.CreateAnnotationRequest{
		ImageID: "img-123",
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().CreateNewAnnotation(gomock.Any(), gomock.Any()).Times(0)

	req, _ := http.NewRequest(http.MethodPost, "/annotations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "CreateAnnotationRequest.Polygon")
}

func TestAnnotationHandler_CreateNewAnnotation_ServiceError(t *testing.T) {
	r, mockService := setupAnnotationHandler(t)

	score := 0.9
	createReq := request.CreateAnnotationRequest{
		ImageID: "img-123",
		Polygon: []model.Point{{X: 1, Y: 1}},
		Score:   &score,
	}
	body, _ := json.Marshal(createReq)

	mockService.EXPECT().
		CreateNewAnnotation(gomock.Any(), gomock.Any()).
		Return(nil, errors.NewValidationError("invalid input", nil))

	req, _ := http.NewRequest(http.MethodPost, "/annotations", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAnnotationHandler_GetAnnotationsByImageID_EmptyResults(t *testing.T) {
	r, mockService := setupAnnotationHandler(t)

	pagination := &query.Pagination{
		Limit:   20,
		Offset:  0,
		SortBy:  "created_at",
		SortDir: "desc",
	}

	mockService.EXPECT().
		GetAnnotationsByImageID(gomock.Any(), "img-123", pagination).
		Return(&query.Result[*model.Annotation]{
			Data:    []*model.Annotation{},
			Limit:   20,
			Offset:  0,
			HasMore: false,
		}, nil)

	req, _ := http.NewRequest(http.MethodGet, "/annotations/image/img-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"data":[]`)
}

func TestAnnotationHandler_GetAnnotationByID_NotFound(t *testing.T) {
	r, mockService := setupAnnotationHandler(t)

	mockService.EXPECT().
		GetAnnotationByID(gomock.Any(), "anno-404").
		Return(nil, errors.NewNotFoundError("not found"))

	req, _ := http.NewRequest(http.MethodGet, "/annotations/anno-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}
