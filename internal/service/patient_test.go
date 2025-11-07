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

func setupPatientService(t *testing.T) (
	*service.PatientService,
	*mocks.MockWorkspaceRepository,
	*mocks.MockPatientRepository,
	*mocks.MockImageRepository,
	*mocks.MockUnitOfWorkFactory,
) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockPatientRepo := mocks.NewMockPatientRepository(ctrl)
	mockImageRepo := mocks.NewMockImageRepository(ctrl)
	mockUOW := mocks.NewMockUnitOfWorkFactory(ctrl)

	mockUOW.EXPECT().WithTx(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context, repos *repository.Repositories) error) error {
			return fn(ctx, &repository.Repositories{
				WorkspaceRepo: mockWorkspaceRepo,
				PatientRepo:   mockPatientRepo,
				ImageRepo:     mockImageRepo,
			})
		},
	)

	pService := service.NewPatientService(mockPatientRepo, mockWorkspaceRepo, mockUOW)
	return pService, mockWorkspaceRepo, mockPatientRepo, mockImageRepo, mockUOW
}

func TestCreateNewPatient_Success(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _, _ := setupPatientService(t)

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

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), input.WorkspaceID).
		Return(&model.Workspace{
			ID:               input.WorkspaceID,
			Name:             "Test Workspace",
			AnnotationTypeID: ptrString("at-id-123"),
		}, nil)

	mockPatientRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(&model.Patient{
			ID:          "patient-123",
			WorkspaceID: input.WorkspaceID,
			Name:        input.Name,
		}, nil)

	createdPatient, err := pService.CreateNewPatient(ctx, input)

	require.NoError(t, err)
	require.NotNil(t, createdPatient)
	require.Equal(t, "patient-123", createdPatient.ID)
	require.Equal(t, input.WorkspaceID, createdPatient.WorkspaceID)
	require.Equal(t, input.Name, createdPatient.Name)
}

func TestCreateNewPatient_Failure_WorkspaceNotReady(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _, _ := setupPatientService(t)

	ctx := context.Background()

	input := service.CreatePatientInput{
		WorkspaceID: "workspace-123",
		Name:        "John Doe",
	}

	mockPatientRepo.EXPECT().
		FindByName(gomock.Any(), input.Name).
		Return(nil, nil)

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), input.WorkspaceID).
		Return(&model.Workspace{
			ID:               input.WorkspaceID,
			Name:             "Test Workspace",
			AnnotationTypeID: nil,
		}, nil)

	mockPatientRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Times(0)

	createdPatient, err := pService.CreateNewPatient(ctx, input)

	require.Error(t, err)
	require.Nil(t, createdPatient)

	var validationErr *errors.Err
	require.True(t, stderrors.As(err, &validationErr))
	assert.Equal(t, errors.ErrorTypeValidation, validationErr.Type)
}

func TestCreateNewPatient_Conflict(t *testing.T) {
	pService, _, mockPatientRepo, _, _ := setupPatientService(t)

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
	pService, mockWorkspaceRepo, mockPatientRepo, _, _ := setupPatientService(t)

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

func TestTransferPatient_Success(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _, _ := setupPatientService(t)

	ctx := context.Background()
	patientID := "patient-123"
	newWorkspaceID := "workspace-456"

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), newWorkspaceID).
		Return(&model.Workspace{
			ID:   newWorkspaceID,
			Name: "New Workspace",
		}, nil)

	mockPatientRepo.EXPECT().
		Transfer(gomock.Any(), patientID, newWorkspaceID).
		Return(nil)

	err := pService.TransferPatientWorkspace(ctx, patientID, newWorkspaceID)

	require.NoError(t, err)
}

func TestTransferPatient_WorkspaceConflictFailure(t *testing.T) {
	pService, mockWorkspaceRepo, _, _, _ := setupPatientService(t)

	ctx := context.Background()
	patientID := "patient-123"
	newWorkspaceID := "workspace-456"

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), newWorkspaceID).
		Return(nil, errors.NewConflictError("Workspace not found", nil))

	err := pService.TransferPatientWorkspace(ctx, patientID, newWorkspaceID)

	require.Error(t, err)
	var conflictErr *errors.Err
	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)
}

func TestDeletePatientByID_Success(t *testing.T) {
	pService, _, mockPatientRepo, mockImageRepo, _ := setupPatientService(t)

	ctx := context.Background()
	patientID := "patient-123"

	filter := []query.Filter{
		{
			Field:    "PatientID",
			Operator: query.OpEqual,
			Value:    patientID,
		},
	}
	paginationOpts := &query.Pagination{
		Limit:  1,
		Offset: 0,
	}
	mockImageRepo.EXPECT().
		FindByFilters(gomock.Any(), filter, paginationOpts).
		Return(&query.Result[*model.Image]{Data: []*model.Image{}}, nil)

	mockPatientRepo.EXPECT().
		Delete(gomock.Any(), patientID).
		Return(nil)

	err := pService.DeletePatientByID(ctx, patientID)

	require.NoError(t, err)
}

func TestDeletePatientByID_WithAssociatedImagesFailure(t *testing.T) {
	pService, _, _, mockImageRepo, _ := setupPatientService(t)

	ctx := context.Background()
	patientID := "patient-123"

	filter := []query.Filter{
		{
			Field:    "PatientID",
			Operator: query.OpEqual,
			Value:    patientID,
		},
	}
	paginationOpts := &query.Pagination{
		Limit:  1,
		Offset: 0,
	}
	mockImageRepo.EXPECT().
		FindByFilters(gomock.Any(), filter, paginationOpts).
		Return(&query.Result[*model.Image]{Data: []*model.Image{
			{ID: "image-1", PatientID: patientID},
		}}, nil)

	err := pService.DeletePatientByID(ctx, patientID)

	require.Error(t, err)
	var conflictErr *errors.Err
	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)
}

func TestUpdatePatient_Success(t *testing.T) {
	pService, _, mockPatientRepo, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientID := "pat-123"
	newName := "New Name"
	newAge := 30
	input := service.UpdatePatientInput{
		Name: &newName,
		Age:  &newAge,
	}

	expectedUpdates := map[string]interface{}{
		constants.PatientNameField: newName,
		constants.PatientAgeField:  newAge,
	}

	mockPatientRepo.EXPECT().Update(ctx, patientID, expectedUpdates).Return(nil)

	err := pService.UpdatePatient(ctx, patientID, input)
	require.NoError(t, err)
}

func TestUpdatePatient_NoUpdates(t *testing.T) {
	pService, _, mockPatientRepo, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientID := "pat-123"
	input := service.UpdatePatientInput{}

	mockPatientRepo.EXPECT().Update(ctx, patientID, gomock.Any()).Times(0)

	err := pService.UpdatePatient(ctx, patientID, input)
	require.NoError(t, err)
}

func TestListPatients_Success(t *testing.T) {
	pService, _, mockPatientRepo, _, _ := setupPatientService(t)
	ctx := context.Background()
	pagination := &query.Pagination{Limit: 10}

	mockPatientRepo.EXPECT().
		FindByFilters(ctx, []query.Filter{}, pagination).
		Return(&query.Result[*model.Patient]{Data: []*model.Patient{
			{ID: "pat-1"},
		}}, nil)

	result, err := pService.ListPatients(ctx, pagination)
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
}

func TestGetPatientsByWorkspaceID_Success(t *testing.T) {
	pService, _, mockPatientRepo, _, _ := setupPatientService(t)
	ctx := context.Background()
	workspaceID := "ws-123"
	pagination := &query.Pagination{Limit: 10}

	expectedFilter := []query.Filter{
		{
			Field:    constants.PatientWorkspaceIDField,
			Operator: query.OpEqual,
			Value:    workspaceID,
		},
	}
	mockPatientRepo.EXPECT().
		FindByFilters(ctx, expectedFilter, pagination).
		Return(&query.Result[*model.Patient]{Data: []*model.Patient{
			{ID: "pat-1", WorkspaceID: workspaceID},
		}}, nil)

	result, err := pService.GetPatientsByWorkspaceID(ctx, workspaceID, pagination)
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
}
