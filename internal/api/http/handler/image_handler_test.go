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
	"github.com/histopathai/main-service/internal/domain/storage"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupImageHandler(t *testing.T) (*gin.Engine, *mocks.MockIImageService) {
	ctrl := gomock.NewController(t)
	mockService := mocks.NewMockIImageService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	v := validator.NewRequestValidator()

	ih := handler.NewImageHandler(mockService, v, logger)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("authenticated_user_id", "test-user-id")
		c.Set("request_id", "test-request-id")
		c.Next()
	})

	r.POST("/images", ih.UploadImage)
	r.GET("/images/:image_id", ih.GetImageByID)
	r.DELETE("/images/:image_id", ih.DeleteImage)

	return r, mockService
}

func TestImageHandler_UploadImage_Success(t *testing.T) {
	r, mockService := setupImageHandler(t)

	uploadReq := request.UploadImageRequest{
		PatientID:   "pat-123",
		ContentType: "image/tiff",
		Name:        "slide.tiff",
		Format:      "TIFF",
	}
	body, _ := json.Marshal(uploadReq)

	mockService.EXPECT().
		UploadImage(gomock.Any(), gomock.Any()).
		Return(&storage.SignedURLPayload{
			URL:     "https://example.com/upload",
			Headers: map[string]string{"Key": "Value"},
		}, nil)

	req, _ := http.NewRequest(http.MethodPost, "/images", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "https://example.com/upload")
}

func TestImageHandler_UploadImage_ValidationError(t *testing.T) {
	r, _ := setupImageHandler(t)

	uploadReq := request.UploadImageRequest{
		ContentType: "image/tiff",
		Name:        "slide.tiff",
		Format:      "TIFF",
	}
	body, _ := json.Marshal(uploadReq)

	req, _ := http.NewRequest(http.MethodPost, "/images", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "UploadImageRequest.PatientID")
}

func TestImageHandler_GetImageByID_NotFound(t *testing.T) {
	r, mockService := setupImageHandler(t)

	mockService.EXPECT().
		GetImageByID(gomock.Any(), "img-404").
		Return(nil, errors.NewNotFoundError("not found"))

	req, _ := http.NewRequest(http.MethodGet, "/images/img-404", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestImageHandler_DeleteImage_Success(t *testing.T) {
	r, mockService := setupImageHandler(t)

	mockService.EXPECT().
		DeleteImageByID(gomock.Any(), "img-123").
		Return(nil)

	req, _ := http.NewRequest(http.MethodDelete, "/images/img-123", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusNoContent, w.Code)
}
