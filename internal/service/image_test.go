package service_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupImageService(t *testing.T) (
	*service.ImageService,
	*mocks.MockImageRepository,
	*mocks.MockPatientRepository,
	*mocks.MockObjectStorage,
	*mocks.MockImageEventPublisher,
) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockImageRepo := mocks.NewMockImageRepository(ctrl)
	mockPatientRepo := mocks.NewMockPatientRepository(ctrl)
	mockStorage := mocks.NewMockObjectStorage(ctrl)
	mockPublisher := mocks.NewMockImageEventPublisher(ctrl)
	mockUOW := mocks.NewMockUnitOfWorkFactory(ctrl)

	imgService := service.NewImageService(
		mockImageRepo,
		mockUOW,
		mockStorage,
		"test-bucket",
		mockPublisher,
		mockPatientRepo,
	)

	return imgService, mockImageRepo, mockPatientRepo, mockStorage, mockPublisher
}

func TestUploadImage_Success(t *testing.T) {
	imgService, _, mockPatientRepo, mockStorage, _ := setupImageService(t)
	ctx := context.Background()

	input := &service.UploadImageInput{
		PatientID:   "patient-123",
		CreatorID:   "creator-123",
		ContentType: "image/tiff",
		Name:        "test.tiff",
		Format:      "TIFF",
	}

	mockPatientRepo.EXPECT().
		Read(ctx, input.PatientID).
		Return(&model.Patient{ID: input.PatientID}, nil)

	mockStorage.EXPECT().
		GenerateSignedURL(ctx, "test-bucket", gomock.Any(), gomock.Any(), input.ContentType, gomock.Any()).
		Return("https://signed-url.example.com", nil)

	url, err := imgService.UploadImage(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, url)
	assert.Contains(t, *url, "https://")
}

func TestUploadImage_InvalidPatient(t *testing.T) {
	imgService, _, mockPatientRepo, _, _ := setupImageService(t)
	ctx := context.Background()

	input := &service.UploadImageInput{
		PatientID:   "invalid-patient",
		CreatorID:   "creator-123",
		ContentType: "image/tiff",
		Name:        "test.tiff",
		Format:      "TIFF",
	}

	mockPatientRepo.EXPECT().
		Read(ctx, input.PatientID).
		Return(nil, errors.NewNotFoundError("patient not found"))

	url, err := imgService.UploadImage(ctx, input)

	require.Error(t, err)
	require.Nil(t, url)
}

func TestConfirmUpload_Success(t *testing.T) {
	imgService, mockImageRepo, _, _, mockPublisher := setupImageService(t)
	ctx := context.Background()

	input := &service.ConfirmUploadInput{
		ImageID:    "image-123",
		PatientID:  "patient-123",
		CreatorID:  "creator-123",
		Name:       "test.tiff",
		Format:     "TIFF",
		Status:     model.StatusUploaded,
		OriginPath: "gcs://bucket/image-123-test.tiff",
	}

	mockImageRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(&model.Image{
			ID:         input.ImageID,
			PatientID:  input.PatientID,
			CreatorID:  input.CreatorID,
			Name:       input.Name,
			Format:     input.Format,
			Status:     input.Status,
			OriginPath: input.OriginPath,
		}, nil)

	mockPublisher.EXPECT().
		PublishImageProcessingRequested(ctx, gomock.Any()).
		Return(nil)

	err := imgService.ConfirmUpload(ctx, input)

	require.NoError(t, err)
}

func TestConfirmUpload_PublishEventFailure(t *testing.T) {
	imgService, mockImageRepo, _, _, mockPublisher := setupImageService(t)
	ctx := context.Background()

	input := &service.ConfirmUploadInput{
		ImageID:    "image-123",
		PatientID:  "patient-123",
		CreatorID:  "creator-123",
		Name:       "test.tiff",
		Format:     "TIFF",
		Status:     model.StatusUploaded,
		OriginPath: "gcs://bucket/image-123-test.tiff",
	}

	mockImageRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(&model.Image{ID: input.ImageID}, nil)

	mockPublisher.EXPECT().
		PublishImageProcessingRequested(ctx, gomock.Any()).
		Return(errors.NewInternalError("failed to publish event", nil))

	err := imgService.ConfirmUpload(ctx, input)

	require.Error(t, err)
	var internalErr *errors.Err
	require.True(t, stderrors.As(err, &internalErr))
	assert.Equal(t, errors.ErrorTypeInternal, internalErr.Type)
}
