package eventhandlers_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"

	stderrors "errors"

	eventhandlers "github.com/histopathai/main-service/internal/application/event_handlers"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupUploadStatusHandler(t *testing.T) (*eventhandlers.UploadStatusHandler, *mocks.MockIImageService) {
	ctrl := gomock.NewController(t)
	mockImageService := mocks.NewMockIImageService(ctrl)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := eventhandlers.NewUploadStatusHandler(mockImageService, logger)
	return handler, mockImageService
}

func TestUploadStatusHandler_Handle_Success(t *testing.T) {
	handler, mockImageService := setupUploadStatusHandler(t)
	ctx := context.Background()

	notification := eventhandlers.GCSNotification{
		Name:   "test-image.tiff",
		Bucket: "test-bucket",
		Metadata: map[string]string{
			"image-id":    "img-123",
			"patient-id":  "pat-123",
			"creator-id":  "user-123",
			"file-name":   "test-image.tiff",
			"format":      "TIFF",
			"origin-path": "img-123-test-image.tiff",
			"status":      string(model.StatusUploaded),
			"width":       "1024",
			"height":      "768",
			"size":        "5000",
		},
	}
	data, err := json.Marshal(notification)
	require.NoError(t, err)

	attributes := map[string]string{"event_id": "evt-1"}

	mockImageService.EXPECT().ConfirmUpload(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, input *service.ConfirmUploadInput) error {
			assert.Equal(t, "img-123", input.ImageID)
			assert.Equal(t, "pat-123", input.PatientID)
			assert.Equal(t, "user-123", input.CreatorID)
			assert.Equal(t, "test-image.tiff", input.Name)
			assert.Equal(t, "TIFF", input.Format)
			assert.Equal(t, "img-123-test-image.tiff", input.OriginPath)
			assert.Equal(t, model.StatusUploaded, input.Status)
			assert.Equal(t, 1024, *input.Width)
			assert.Equal(t, 768, *input.Height)
			assert.Equal(t, int64(5000), *input.Size)
			return nil
		},
	)

	err = handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestUploadStatusHandler_Handle_MissingImageId(t *testing.T) {
	handler, mockImageService := setupUploadStatusHandler(t)
	ctx := context.Background()

	notification := eventhandlers.GCSNotification{
		Name:   "test-image.tiff",
		Bucket: "test-bucket",
		Metadata: map[string]string{
			"patient-id": "pat-123",
		},
	}
	data, err := json.Marshal(notification)
	require.NoError(t, err)

	attributes := map[string]string{"event_id": "evt-1"}

	mockImageService.EXPECT().ConfirmUpload(ctx, gomock.Any()).Times(0)

	err = handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestUploadStatusHandler_Handle_MissingMetadata(t *testing.T) {
	handler, mockImageService := setupUploadStatusHandler(t)
	ctx := context.Background()

	notification := eventhandlers.GCSNotification{
		Name:     "test-image.tiff",
		Bucket:   "test-bucket",
		Metadata: nil,
	}
	data, err := json.Marshal(notification)
	require.NoError(t, err)

	attributes := map[string]string{"event_id": "evt-1"}

	mockImageService.EXPECT().ConfirmUpload(ctx, gomock.Any()).Times(0)

	err = handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestUploadStatusHandler_Handle_InvalidJson(t *testing.T) {
	handler, mockImageService := setupUploadStatusHandler(t)
	ctx := context.Background()

	data := []byte(`{"invalid" "json"}`)
	attributes := map[string]string{"event_id": "evt-1"}

	mockImageService.EXPECT().ConfirmUpload(ctx, gomock.Any()).Times(0)

	err := handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestUploadStatusHandler_Handle_ConfirmUploadError(t *testing.T) {
	handler, mockImageService := setupUploadStatusHandler(t)
	ctx := context.Background()

	notification := eventhandlers.GCSNotification{
		Name:   "test-image.tiff",
		Bucket: "test-bucket",
		Metadata: map[string]string{
			"image-id": "img-123",
		},
	}
	data, err := json.Marshal(notification)
	require.NoError(t, err)

	attributes := map[string]string{"event_id": "evt-1"}

	mockImageService.EXPECT().ConfirmUpload(ctx, gomock.Any()).Return(errors.NewInternalError("db error", nil))

	err = handler.Handle(ctx, data, attributes)
	require.Error(t, err)
	var internalErr *errors.Err
	require.True(t, stderrors.As(err, &internalErr))
	assert.Equal(t, errors.ErrorTypeInternal, internalErr.Type)
}

func TestUploadStatusHandler_Handle_InvalidMetadataValues(t *testing.T) {
	handler, mockImageService := setupUploadStatusHandler(t)
	ctx := context.Background()

	notification := eventhandlers.GCSNotification{
		Name:   "test-image.tiff",
		Bucket: "test-bucket",
		Metadata: map[string]string{
			"image-id": "img-123",
			"width":    "not-an-int",
			"height":   "not-an-int",
			"size":     "not-an-int",
		},
	}
	data, err := json.Marshal(notification)
	require.NoError(t, err)

	attributes := map[string]string{"event_id": "evt-1"}

	mockImageService.EXPECT().ConfirmUpload(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, input *service.ConfirmUploadInput) error {
			assert.Equal(t, "img-123", input.ImageID)
			assert.Nil(t, input.Width)
			assert.Nil(t, input.Height)
			assert.Nil(t, input.Size)
			return nil
		},
	)

	err = handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}
