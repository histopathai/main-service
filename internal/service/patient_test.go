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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupPatientService(t *testing.T) (
	*service.PatientService,
	*mocks.MockWorkspaceRepository,
	*mocks.MockPatientRepository,
	*mocks.MockUnitOfWorkFactory,
) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockPatientRepo := mocks.NewMockPatientRepository(ctrl)
	mockUOW := mocks.NewMockUnitOfWorkFactory(ctrl)

	mockUOW.EXPECT().WithTx(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context, repos *repository.Repositories) error) error {
			return fn(ctx, &repository.Repositories{
				WorkspaceRepo: mockWorkspaceRepo,
			})
		},
	)

	pService := service.NewPatientService(mockPatientRepo, mockWorkspaceRepo, mockUOW)
	return pService, mockWorkspaceRepo, mockPatientRepo, mockUOW
}

func TestCreateNewPatient_Success(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _ := setupPatientService(t)

	ctx := context.Background()

	input := service.CreatePatientInput{
		WorkspaceID: "workspace-123",
		Name:        "John Doe",
		Age:         nil,
		Gender:      nil,
		Race:        nil,
		Disease:     nil,
		Subtype:     nil,
		Grade:       nil,
		History:     nil,
	}

	mockPatientRepo.EXPECT().
		FindByName(gomock.Any(), input.Name).
		Return(nil, nil)

	mockPatientRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(&model.Patient{
			ID:          "patient-123",
			WorkspaceID: input.WorkspaceID,
			Name:        input.Name,
		}, nil)

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), input.WorkspaceID).
		Return(&model.Workspace{
			ID:   input.WorkspaceID,
			Name: "Test Workspace",
		}, nil)

	createdPatient, err := pService.CreateNewPatient(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, createdPatient)
	require.Equal(t, "patient-123", createdPatient.ID)
	require.Equal(t, input.WorkspaceID, createdPatient.WorkspaceID)
	require.Equal(t, input.Name, createdPatient.Name)
}

func TestCreateNewPatient_Conflict(t *testing.T) {
	pService, _, mockPatientRepo, _ := setupPatientService(t)

	ctx := context.Background()

	input := service.CreatePatientInput{
		WorkspaceID: "workspace-123",
		Name:        "John Doe",
	}

	mockPatientRepo.EXPECT().
		FindByName(gomock.Any(), input.Name).
		Return(&model.Patient{
			ID:          "patient-456",
			WorkspaceID: input.WorkspaceID,
			Name:        input.Name,
		}, nil)

	createdPatient, err := pService.CreateNewPatient(ctx, input)

	require.Error(t, err)
	require.Nil(t, createdPatient)

	var conflictErr *errors.Err

	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)
}

func TestCreateNewPatient_WorkspaceValidationFailure(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _ := setupPatientService(t)

	ctx := context.Background()

	input := service.CreatePatientInput{
		WorkspaceID: "invalid-workspace",
		Name:        "John Doe",
	}

	mockPatientRepo.EXPECT().
		FindByName(gomock.Any(), input.Name).
		Return(nil, nil)

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), input.WorkspaceID).
		Return(nil, stderrors.New("workspace not found"))

	createdPatient, err := pService.CreateNewPatient(ctx, input)

	require.Error(t, err)
	require.Nil(t, createdPatient)

	var validationErr *errors.Err

	require.True(t, stderrors.As(err, &validationErr))
	assert.Equal(t, errors.ErrorTypeValidation, validationErr.Type)
}
