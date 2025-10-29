package service_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/mocks"
	"github.com/histopathai/main-service-refactor/internal/service"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func ptrFloat64(f float64) *float64 {
	return &f
}

func setupAnnotationTypeService(t *testing.T) (
	*service.AnnotationTypeService,
	*mocks.MockUnitOfWorkFactory,
	*mocks.MockAnnotationTypeRepository,
) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAnnotationTypeRepo := mocks.NewMockAnnotationTypeRepository(ctrl)
	mockUOW := mocks.NewMockUnitOfWorkFactory(ctrl)

	mockUOW.EXPECT().WithTx(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context, repos *repository.Repositories) error) error {
			return fn(ctx, &repository.Repositories{
				AnnotationTypeRepo: mockAnnotationTypeRepo,
			})
		},
	)

	aService := service.NewAnnotationTypeService(mockAnnotationTypeRepo, mockUOW)
	return aService, mockUOW, mockAnnotationTypeRepo
}

func TestCreateNewAnnotationType_Success(t *testing.T) {
	aService, _, mockAnnotationTypeRepo := setupAnnotationTypeService(t)

	ctx := context.Background()

	input := service.CreateAnnotationTypeInput{
		Name:                  "Test Annotation Type",
		ScoreEnabled:          false,
		ClassificationEnabled: true,
		ClassList:             []string{"Class A", "Class B"},
	}
	mockAnnotationTypeRepo.EXPECT().
		FindByName(gomock.Any(), input.Name).
		Return(nil, stderrors.New("not found"))

	mockAnnotationTypeRepo.EXPECT().
		FindByFilters(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&query.Result[*model.AnnotationType]{Data: []*model.AnnotationType{}}, nil)

	classList := []string{"Class A", "Class B"}
	mockAnnotationTypeRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(&model.AnnotationType{
			ID:                    "annotation-type-123",
			Name:                  "Test Annotation Type",
			ScoreEnabled:          false,
			ClassificationEnabled: true,
			ClassList:             classList,
		}, nil)

	created, err := aService.CreateNewAnnotationType(ctx, &input)

	require.NoError(t, err)
	require.NotNil(t, created)
	require.Equal(t, "annotation-type-123", created.ID)
	require.Equal(t, input.Name, created.Name)
	require.Equal(t, input.ScoreEnabled, created.ScoreEnabled)
	require.Equal(t, input.ClassificationEnabled, created.ClassificationEnabled)
	require.Equal(t, input.ClassList, created.ClassList)
}

func TestValidateAnnotationTypeCreation_Failure_MissingClassList(t *testing.T) {
	aService, _, _ := setupAnnotationTypeService(t)

	ctx := context.Background()

	input := service.CreateAnnotationTypeInput{
		Name:                  "Invalid Annotation Type",
		ScoreEnabled:          false,
		ClassificationEnabled: true,
		ClassList:             []string{},
	}

	err := aService.ValidateAnnotationTypeCreation(ctx, &input)

	require.Error(t, err)
	var validationError *errors.Err
	require.True(t, stderrors.As(err, &validationError))
	assert.Equal(t, errors.ErrorTypeValidation, validationError.Type)

}

func TestValidateAnnotationTypeCreation_Failure_MissingScoreRange(t *testing.T) {
	aService, _, _ := setupAnnotationTypeService(t)

	ctx := context.Background()

	input := service.CreateAnnotationTypeInput{
		Name:                  "Invalid Annotation Type",
		ScoreEnabled:          true,
		ClassificationEnabled: false,
	}

	err := aService.ValidateAnnotationTypeCreation(ctx, &input)

	require.Error(t, err)
	var validationError *errors.Err
	require.True(t, stderrors.As(err, &validationError))
	assert.Equal(t, errors.ErrorTypeValidation, validationError.Type)
}

func TestValidateAnnotationTypeCreation_Failure_BothTypesEnabled(t *testing.T) {
	aService, _, _ := setupAnnotationTypeService(t)

	ctx := context.Background()

	input := service.CreateAnnotationTypeInput{
		Name:                  "Valid Annotation Type",
		ScoreEnabled:          true,
		ScoreMin:              ptrFloat64(0.0),
		ScoreMax:              ptrFloat64(1.0),
		ClassificationEnabled: true,
		ClassList:             []string{"Class A", "Class B"},
	}

	err := aService.ValidateAnnotationTypeCreation(ctx, &input)

	require.Error(t, err)
	var validationError *errors.Err
	require.True(t, stderrors.As(err, &validationError))
	assert.Equal(t, errors.ErrorTypeValidation, validationError.Type)
}
