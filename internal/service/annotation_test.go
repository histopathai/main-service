package service_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/mocks"
	"github.com/histopathai/main-service-refactor/internal/service"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupAnnotationService(t *testing.T) (
	*service.AnnotationService,
	*mocks.MockAnnotationRepository,
	*mocks.MockUnitOfWorkFactory,
) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAnnotationRepo := mocks.NewMockAnnotationRepository(ctrl)
	mockUOW := mocks.NewMockUnitOfWorkFactory(ctrl)

	aService := service.NewAnnotationService(mockAnnotationRepo, mockUOW)
	return aService, mockAnnotationRepo, mockUOW
}

func TestCreateAnnotation_Success_WithScore(t *testing.T) {
	aService, mockAnnotationRepo, _ := setupAnnotationService(t)
	ctx := context.Background()

	score := 4.5
	input := &service.CreateAnnotationInput{
		ImageID:     "image-123",
		AnnotatorID: "annotator-123",
		Polygon: []model.Point{
			{X: 0, Y: 0},
			{X: 100, Y: 0},
			{X: 100, Y: 100},
			{X: 0, Y: 100},
		},
		Score: &score,
	}

	mockAnnotationRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(&model.Annotation{
			ID:          "annotation-123",
			ImageID:     input.ImageID,
			AnnotatorID: input.AnnotatorID,
			Polygon:     input.Polygon,
			Score:       input.Score,
		}, nil)

	created, err := aService.CreateNewAnnotation(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, "annotation-123", created.ID)
	assert.Equal(t, input.ImageID, created.ImageID)
	assert.Equal(t, *input.Score, *created.Score)
}

func TestCreateAnnotation_Success_WithClass(t *testing.T) {
	aService, mockAnnotationRepo, _ := setupAnnotationService(t)
	ctx := context.Background()

	class := "Malignant"
	input := &service.CreateAnnotationInput{
		ImageID:     "image-123",
		AnnotatorID: "annotator-123",
		Polygon: []model.Point{
			{X: 0, Y: 0},
			{X: 100, Y: 0},
			{X: 100, Y: 100},
			{X: 0, Y: 100},
		},
		Class: &class,
	}

	mockAnnotationRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(&model.Annotation{
			ID:          "annotation-123",
			ImageID:     input.ImageID,
			AnnotatorID: input.AnnotatorID,
			Polygon:     input.Polygon,
			Class:       input.Class,
		}, nil)

	created, err := aService.CreateNewAnnotation(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, *input.Class, *created.Class)
}

func TestCreateAnnotation_Failure_NoScoreOrClass(t *testing.T) {
	aService, _, _ := setupAnnotationService(t)
	ctx := context.Background()

	input := &service.CreateAnnotationInput{
		ImageID:     "image-123",
		AnnotatorID: "annotator-123",
		Polygon: []model.Point{
			{X: 0, Y: 0},
			{X: 100, Y: 0},
		},
		// Neither Score nor Class provided
	}

	created, err := aService.CreateNewAnnotation(ctx, input)

	require.Error(t, err)
	require.Nil(t, created)
	var validationErr *errors.Err
	require.True(t, stderrors.As(err, &validationErr))
	assert.Equal(t, errors.ErrorTypeValidation, validationErr.Type)
}

func TestGetAnnotationsByImageID_Success(t *testing.T) {
	aService, mockAnnotationRepo, _ := setupAnnotationService(t)
	ctx := context.Background()

	imageID := "image-123"
	score := 3.5

	mockAnnotationRepo.EXPECT().
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, filters []sharedQuery.Filter, pagination *sharedQuery.Pagination) (*sharedQuery.Result[*model.Annotation], error) {
			assert.Equal(t, "ImageID", filters[0].Field)
			assert.Equal(t, sharedQuery.OpEqual, filters[0].Operator)
			assert.Equal(t, imageID, filters[0].Value)

			return &sharedQuery.Result[*model.Annotation]{
				Data: []*model.Annotation{
					{
						ID:          "annotation-1",
						ImageID:     imageID,
						AnnotatorID: "annotator-1",
						Score:       &score,
					},
					{
						ID:          "annotation-2",
						ImageID:     imageID,
						AnnotatorID: "annotator-2",
						Score:       &score,
					},
				},
			}, nil
		})

	pagination := &sharedQuery.Pagination{
		Limit:  -1,
		Offset: 0,
	}
	result, err := aService.GetAnnotationsByImageID(ctx, imageID, pagination)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Data, 2)
	assert.Equal(t, imageID, result.Data[0].ImageID)
}

func TestGetAnnotationsByImageID_NoResults(t *testing.T) {
	aService, mockAnnotationRepo, _ := setupAnnotationService(t)
	ctx := context.Background()

	imageID := "image-123"

	mockAnnotationRepo.EXPECT().
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		Return(&sharedQuery.Result[*model.Annotation]{
			Data: []*model.Annotation{},
		}, nil)

	pagination := &sharedQuery.Pagination{
		Limit:  -1,
		Offset: 0,
	}
	result, err := aService.GetAnnotationsByImageID(ctx, imageID, pagination)

	require.NoError(t, err)
	require.Nil(t, result)
}

func TestDeleteAnnotation_Success(t *testing.T) {
	aService, mockAnnotationRepo, _ := setupAnnotationService(t)
	ctx := context.Background()

	annotationID := "annotation-123"

	mockAnnotationRepo.EXPECT().
		Delete(ctx, annotationID).
		Return(nil)

	err := aService.DeleteAnnotation(ctx, annotationID)

	require.NoError(t, err)
}

func TestGetAnnotationByID_Success(t *testing.T) {
	aService, mockAnnotationRepo, _ := setupAnnotationService(t)
	ctx := context.Background()

	annotationID := "annotation-123"
	score := 4.0

	mockAnnotationRepo.EXPECT().
		Read(ctx, annotationID).
		Return(&model.Annotation{
			ID:          annotationID,
			ImageID:     "image-123",
			AnnotatorID: "annotator-123",
			Score:       &score,
		}, nil)

	annotation, err := aService.GetAnnotationByID(ctx, annotationID)

	require.NoError(t, err)
	require.NotNil(t, annotation)
	assert.Equal(t, annotationID, annotation.ID)
}

func TestGetAnnotationByID_NotFound(t *testing.T) {
	aService, mockAnnotationRepo, _ := setupAnnotationService(t)
	ctx := context.Background()

	annotationID := "nonexistent"

	mockAnnotationRepo.EXPECT().
		Read(ctx, annotationID).
		Return(nil, errors.NewNotFoundError("annotation not found"))

	annotation, err := aService.GetAnnotationByID(ctx, annotationID)

	require.Error(t, err)
	require.Nil(t, annotation)
}
