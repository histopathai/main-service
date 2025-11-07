package service_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
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
	*mocks.MockWorkspaceRepository,
) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockAnnotationTypeRepo := mocks.NewMockAnnotationTypeRepository(ctrl)
	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockUOW := mocks.NewMockUnitOfWorkFactory(ctrl)

	mockUOW.EXPECT().WithTx(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context, repos *repository.Repositories) error) error {
			return fn(ctx, &repository.Repositories{
				AnnotationTypeRepo: mockAnnotationTypeRepo,
				WorkspaceRepo:      mockWorkspaceRepo,
			})
		},
	)

	aService := service.NewAnnotationTypeService(mockAnnotationTypeRepo, mockUOW)
	return aService, mockUOW, mockAnnotationTypeRepo, mockWorkspaceRepo
}

func TestCreateNewAnnotationType_Success(t *testing.T) {
	aService, _, mockAnnotationTypeRepo, _ := setupAnnotationTypeService(t)

	ctx := context.Background()

	input := service.CreateAnnotationTypeInput{
		Name:                  "Test Annotation Type",
		ScoreEnabled:          false,
		ClassificationEnabled: true,
		ClassList:             []string{"Class A", "Class B"},
	}
	mockAnnotationTypeRepo.EXPECT().
		FindByName(gomock.Any(), input.Name).
		Return(nil, nil)

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
	aService, _, _, _ := setupAnnotationTypeService(t)

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
	aService, _, _, _ := setupAnnotationTypeService(t)

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
	aService, _, _, _ := setupAnnotationTypeService(t)

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

func TestDeleteAnnotationType_Success(t *testing.T) {
	aService, _, mockAnnotationTypeRepo, mockWorkspaceRepo := setupAnnotationTypeService(t)
	ctx := context.Background()
	typeID := "at-123"

	mockWorkspaceRepo.EXPECT().
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		Return(&query.Result[*model.Workspace]{Data: []*model.Workspace{}}, nil)

	mockAnnotationTypeRepo.EXPECT().
		Delete(ctx, typeID).
		Return(nil)

	err := aService.DeleteAnnotationType(ctx, typeID)
	require.NoError(t, err)
}

func TestDeleteAnnotationType_Conflict(t *testing.T) {
	aService, _, _, mockWorkspaceRepo := setupAnnotationTypeService(t)
	ctx := context.Background()
	typeID := "at-123"

	mockWorkspaceRepo.EXPECT().
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		Return(&query.Result[*model.Workspace]{Data: []*model.Workspace{
			{ID: "ws-1"},
		}}, nil)

	err := aService.DeleteAnnotationType(ctx, typeID)
	require.Error(t, err)
	var conflictErr *errors.Err
	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)
}

func TestUpdateAnnotationType_Success(t *testing.T) {
	aService, _, mockAnnotationTypeRepo, _ := setupAnnotationTypeService(t)
	ctx := context.Background()
	typeID := "at-123"
	desc := "new description"
	input := &service.UpdateAnnotationTypeInput{
		Description: &desc,
	}

	expectedUpdates := map[string]interface{}{
		constants.AnnotationTypeDescField: desc,
	}

	mockAnnotationTypeRepo.EXPECT().
		Update(ctx, typeID, expectedUpdates).
		Return(nil)

	err := aService.UpdateAnnotationType(ctx, typeID, input)
	require.NoError(t, err)
}

func TestGetClassificationAnnotationTypes_Success(t *testing.T) {
	aService, _, mockAnnotationTypeRepo, _ := setupAnnotationTypeService(t)
	ctx := context.Background()
	pagination := &query.Pagination{Limit: 10}

	expectedFilter := []query.Filter{
		{
			Field:    constants.AnnotationTypeClassificationEnabledField,
			Operator: query.OpEqual,
			Value:    true,
		},
	}
	mockAnnotationTypeRepo.EXPECT().
		FindByFilters(ctx, expectedFilter, pagination).
		Return(&query.Result[*model.AnnotationType]{Data: []*model.AnnotationType{
			{ID: "at-1", ClassificationEnabled: true},
		}}, nil)

	result, err := aService.GetClassificationAnnotationTypes(ctx, pagination)
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
	assert.True(t, result.Data[0].ClassificationEnabled)
}

func TestGetScoreAnnotationTypes_Success(t *testing.T) {
	aService, _, mockAnnotationTypeRepo, _ := setupAnnotationTypeService(t)
	ctx := context.Background()
	pagination := &query.Pagination{Limit: 10}

	expectedFilter := []query.Filter{
		{
			Field:    constants.AnnotationTypeScoreEnabledField,
			Operator: query.OpEqual,
			Value:    true,
		},
	}
	mockAnnotationTypeRepo.EXPECT().
		FindByFilters(ctx, expectedFilter, pagination).
		Return(&query.Result[*model.AnnotationType]{Data: []*model.AnnotationType{
			{ID: "at-1", ScoreEnabled: true},
		}}, nil)

	result, err := aService.GetScoreAnnotationTypes(ctx, pagination)
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
	assert.True(t, result.Data[0].ScoreEnabled)
}
