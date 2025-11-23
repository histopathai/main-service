package service_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/domain/storage"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
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

	mockUOW.EXPECT().WithTx(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context, repos *repository.Repositories) error) error {
			return fn(ctx, &repository.Repositories{
				ImageRepo:   mockImageRepo,
				PatientRepo: mockPatientRepo,
			})
		},
	)

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

	mockPayload := &storage.SignedURLPayload{
		URL:     "https://signed-url.example.com",
		Headers: map[string]string{"x-goog-meta-image-id": "image-123"},
	}

	mockStorage.EXPECT().
		GenerateSignedURL(ctx, "test-bucket", gomock.Any(), gomock.Any(), input.ContentType, gomock.Any()).
		Return(mockPayload, nil)

	url, err := imgService.UploadImage(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, url)
	assert.Contains(t, url.URL, "https://")
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

func TestListImageByPatientID_Success(t *testing.T) {
	imgService, mockImageRepo, _, _, _ := setupImageService(t)
	ctx := context.Background()
	patientID := "pat-123"
	pagination := &query.Pagination{Limit: 10}

	expectedFilter := []query.Filter{
		{
			Field:    constants.ImagePatientIDField,
			Operator: query.OpEqual,
			Value:    patientID,
		},
	}
	mockImageRepo.EXPECT().
		FindByFilters(ctx, expectedFilter, pagination).
		Return(&query.Result[*model.Image]{Data: []*model.Image{
			{ID: "img-1", PatientID: patientID},
			{ID: "img-2", PatientID: patientID},
		}}, nil)

	result, err := imgService.ListImageByPatientID(ctx, patientID, pagination)
	require.NoError(t, err)
	require.Len(t, result.Data, 2)
}

func TestDeleteImageByID_Success(t *testing.T) {
	imgService, _, _, _, mockPublisher := setupImageService(t)
	ctx := context.Background()
	imageID := "img-123"

	mockPublisher.EXPECT().
		PublishImageDeletionRequested(ctx, gomock.Any()).
		Return(nil)

	err := imgService.DeleteImageByID(ctx, imageID)
	require.NoError(t, err)
}

func TestGetImageByID_NotFound(t *testing.T) {
	imgService, mockImageRepo, _, _, _ := setupImageService(t)
	ctx := context.Background()
	imageID := "img-not-found"

	mockImageRepo.EXPECT().Read(ctx, imageID).Return(nil, errors.NewNotFoundError("not found"))

	img, err := imgService.GetImageByID(ctx, imageID)
	require.Error(t, err)
	require.Nil(t, img)
	var notFoundErr *errors.Err
	require.True(t, stderrors.As(err, &notFoundErr))
	assert.Equal(t, errors.ErrorTypeNotFound, notFoundErr.Type)
}

func TestBatchTransferImage_Success(t *testing.T) {
	imgService, mockImageRepo, mockPatientRepo, _, _ := setupImageService(t)
	ctx := context.Background()
	imageIDs := []string{"img-1", "img-2", "img-3"}
	newPatientID := "new-patient-123"

	mockImageRepo.EXPECT().
		BatchTransfer(ctx, imageIDs, newPatientID).
		Return(nil)

	mockPatientRepo.EXPECT().
		Read(ctx, newPatientID).
		Return(&model.Patient{ID: newPatientID}, nil)

	err := imgService.BatchTransferImages(ctx, imageIDs, newPatientID)
	require.NoError(t, err)
}

func TestBatchTransferImage_PatientNotFound(t *testing.T) {
	imgService, _, mockPatientRepo, _, _ := setupImageService(t)
	ctx := context.Background()
	imageIDs := []string{"img-1", "img-2", "img-3"}
	newPatientID := "nonexistent-patient"

	mockPatientRepo.EXPECT().
		Read(ctx, newPatientID).
		Return(nil, errors.NewNotFoundError("patient not found"))

	err := imgService.BatchTransferImages(ctx, imageIDs, newPatientID)
	require.Error(t, err)
	var notFoundErr *errors.Err
	require.True(t, stderrors.As(err, &notFoundErr))
	assert.Equal(t, errors.ErrorTypeNotFound, notFoundErr.Type)
}

func TestUploadImage_StorageFailure(t *testing.T) {
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
		Return(nil, errors.NewInternalError("storage unreachable", nil))

	url, err := imgService.UploadImage(ctx, input)

	require.Error(t, err)
	require.Nil(t, url)
}

func TestConfirmUpload_RepoCreateFailure(t *testing.T) {
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
		Return(nil, errors.NewInternalError("db insert failed", nil))

	mockPublisher.EXPECT().
		PublishImageProcessingRequested(gomock.Any(), gomock.Any()).
		Times(0)

	err := imgService.ConfirmUpload(ctx, input)

	require.Error(t, err)
}

func TestGetImageByID_Success(t *testing.T) {
	imgService, mockImageRepo, _, _, _ := setupImageService(t)
	ctx := context.Background()
	imageID := "img-123"

	mockImageRepo.EXPECT().
		Read(ctx, imageID).
		Return(&model.Image{ID: imageID, Name: "Slide 1"}, nil)

	img, err := imgService.GetImageByID(ctx, imageID)
	require.NoError(t, err)
	assert.Equal(t, imageID, img.ID)
}

func TestBatchDeleteImages_Success(t *testing.T) {
	imgService, _, _, _, mockPublisher := setupImageService(t)
	ctx := context.Background()
	ids := []string{"img-1", "img-2"}

	mockPublisher.EXPECT().
		PublishImageDeletionRequested(ctx, gomock.Any()).
		Times(len(ids)).
		Return(nil)

	err := imgService.BatchDeleteImages(ctx, ids)
	require.NoError(t, err)
}
func TestBatchDeleteImages_Failure(t *testing.T) {
	imgService, _, _, _, mockPublisher := setupImageService(t)
	ctx := context.Background()
	ids := []string{"img-1"}

	mockPublisher.EXPECT().
		PublishImageDeletionRequested(ctx, gomock.Any()).
		Return(errors.NewInternalError("failed to publish event", nil))

	err := imgService.BatchDeleteImages(ctx, ids)
	require.Error(t, err)

	var internalErr *errors.Err
	require.True(t, stderrors.As(err, &internalErr))
}

func TestBatchTransferImages_RepoFailure(t *testing.T) {
	imgService, mockImageRepo, mockPatientRepo, _, _ := setupImageService(t)
	ctx := context.Background()
	imageIDs := []string{"img-1"}
	newPatientID := "pat-new"

	mockPatientRepo.EXPECT().
		Read(ctx, newPatientID).
		Return(&model.Patient{ID: newPatientID}, nil)

	mockImageRepo.EXPECT().
		BatchTransfer(ctx, imageIDs, newPatientID).
		Return(errors.NewInternalError("update failed", nil))

	err := imgService.BatchTransferImages(ctx, imageIDs, newPatientID)
	require.Error(t, err)
}

func TestCountImages_Success(t *testing.T) {
	imgService, mockImageRepo, _, _, _ := setupImageService(t)
	ctx := context.Background()
	filters := []query.Filter{{Field: constants.ImagePatientIDField, Value: "p1"}}

	mockImageRepo.EXPECT().
		Count(ctx, filters).
		Return(int64(15), nil)

	count, err := imgService.CountImages(ctx, filters)
	require.NoError(t, err)
	assert.Equal(t, int64(15), count)
}
