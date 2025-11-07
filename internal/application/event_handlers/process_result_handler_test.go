package eventhandlers_test

import (
	"context"
	stderrors "errors"
	"io"
	"log/slog"
	"testing"
	"time"

	eventhandlers "github.com/histopathai/main-service/internal/application/event_handlers"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupProcessResultHandler(t *testing.T) (
	*eventhandlers.ProcessResultHandler,
	*mocks.MockImageRepository,
	*mocks.MockImageEventPublisher,
	events.EventSerializer,
) {
	ctrl := gomock.NewController(t)
	mockImageRepo := mocks.NewMockImageRepository(ctrl)
	mockPublisher := mocks.NewMockImageEventPublisher(ctrl)
	serializer := events.NewJSONEventSerializer()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	handler := eventhandlers.NewProcessResultHandler(
		mockImageRepo,
		serializer,
		mockPublisher,
		logger,
	)
	return handler, mockImageRepo, mockPublisher, serializer
}

func TestProcessResultHandler_Handle_Completed_Success(t *testing.T) {
	handler, mockImageRepo, _, serializer := setupProcessResultHandler(t)
	ctx := context.Background()

	event := eventhandlers.ImageProcessingCompletedEvent{
		BaseEvent:     events.NewBaseEvent(events.EventTypeImageProcessingCompleted),
		ImageID:       "img-123",
		ProcessedPath: "processed/path",
		Width:         1024,
		Height:        768,
		Size:          5000,
	}
	data, err := serializer.Serialize(event)
	require.NoError(t, err)

	attributes := map[string]string{
		"event_type": string(events.EventTypeImageProcessingCompleted),
	}

	expectedUpdates := map[string]interface{}{
		constants.ImageStatusField:        model.StatusProcessed,
		constants.ImageProcessedPathField: event.ProcessedPath,
		constants.ImageWidthField:         event.Width,
		constants.ImageHeightField:        event.Height,
		constants.ImageSizeField:          event.Size,
	}

	mockImageRepo.EXPECT().Update(ctx, event.ImageID, expectedUpdates).Return(nil)

	err = handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestProcessResultHandler_Handle_Completed_DeserializeError(t *testing.T) {
	handler, mockImageRepo, _, _ := setupProcessResultHandler(t)
	ctx := context.Background()

	data := []byte(`{"invalid_json": "value"`)
	attributes := map[string]string{
		"event_type": string(events.EventTypeImageProcessingCompleted),
	}

	mockImageRepo.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).Times(0)

	err := handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestProcessResultHandler_Handle_Failed_Retry(t *testing.T) {
	handler, mockImageRepo, mockPublisher, serializer := setupProcessResultHandler(t)
	ctx := context.Background()

	event := events.NewImageProcessingFailedEvent("img-123", "processing error")
	data, err := serializer.Serialize(event)
	require.NoError(t, err)

	attributes := map[string]string{
		"event_type": string(events.EventTypeImageProcessingFailed),
	}

	mockImage := &model.Image{
		ID:         "img-123",
		OriginPath: "origin/path",
		RetryCount: 1,
	}

	mockImageRepo.EXPECT().Read(ctx, event.ImageID).Return(mockImage, nil)

	mockImageRepo.EXPECT().Update(ctx, event.ImageID, gomock.Any()).DoAndReturn(
		func(ctx context.Context, id string, updates map[string]interface{}) error {
			assert.Equal(t, model.StatusProcessing, updates[constants.ImageStatusField])
			assert.Equal(t, "processing error", updates[constants.ImageFailureReasonField])
			assert.Equal(t, 2, updates[constants.ImageRetryCountField])
			assert.WithinDuration(t, time.Now(), updates[constants.ImageLastProcessedAtField].(time.Time), time.Second)
			return nil
		},
	)

	mockPublisher.EXPECT().PublishImageProcessingRequested(ctx, gomock.Any()).Return(nil)

	err = handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestProcessResultHandler_Handle_Failed_MaxRetriesReached(t *testing.T) {
	handler, mockImageRepo, mockPublisher, serializer := setupProcessResultHandler(t)
	ctx := context.Background()

	event := events.NewImageProcessingFailedEvent("img-123", "processing error")
	data, err := serializer.Serialize(event)
	require.NoError(t, err)

	attributes := map[string]string{
		"event_type": string(events.EventTypeImageProcessingFailed),
	}

	mockImage := &model.Image{
		ID:         "img-123",
		OriginPath: "origin/path",
		RetryCount: eventhandlers.MaxRetries,
	}

	mockImageRepo.EXPECT().Read(ctx, event.ImageID).Return(mockImage, nil)

	mockImageRepo.EXPECT().Update(ctx, event.ImageID, gomock.Any()).DoAndReturn(
		func(ctx context.Context, id string, updates map[string]interface{}) error {
			assert.Equal(t, model.StatusFailed, updates[constants.ImageStatusField])
			assert.Equal(t, eventhandlers.MaxRetries+1, updates[constants.ImageRetryCountField])
			return nil
		},
	)

	mockPublisher.EXPECT().PublishImageProcessingRequested(ctx, gomock.Any()).Times(0)

	err = handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}

func TestProcessResultHandler_Handle_Failed_ReadError(t *testing.T) {
	handler, mockImageRepo, _, serializer := setupProcessResultHandler(t)
	ctx := context.Background()

	event := events.NewImageProcessingFailedEvent("img-123", "processing error")
	data, err := serializer.Serialize(event)
	require.NoError(t, err)

	attributes := map[string]string{
		"event_type": string(events.EventTypeImageProcessingFailed),
	}

	mockImageRepo.EXPECT().Read(ctx, event.ImageID).Return(nil, errors.NewInternalError("db error", nil))

	err = handler.Handle(ctx, data, attributes)
	require.Error(t, err)
	var internalErr *errors.Err
	require.True(t, stderrors.As(err, &internalErr))
	assert.Equal(t, errors.ErrorTypeInternal, internalErr.Type)
}

func TestProcessResultHandler_Handle_UnknownEventType(t *testing.T) {
	handler, mockImageRepo, _, _ := setupProcessResultHandler(t)
	ctx := context.Background()

	data := []byte(`{}`)
	attributes := map[string]string{
		"event_type": "unknown.event.v1",
	}

	mockImageRepo.EXPECT().Read(ctx, gomock.Any()).Times(0)
	mockImageRepo.EXPECT().Update(ctx, gomock.Any(), gomock.Any()).Times(0)

	err := handler.Handle(ctx, data, attributes)
	require.NoError(t, err)
}
