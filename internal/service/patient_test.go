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
	*mocks.MockAnnotationRepository,
	*mocks.MockUnitOfWorkFactory,
) {
	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	mockWorkspaceRepo := mocks.NewMockWorkspaceRepository(ctrl)
	mockPatientRepo := mocks.NewMockPatientRepository(ctrl)
	mockImageRepo := mocks.NewMockImageRepository(ctrl)
	mockAnnotationRepo := mocks.NewMockAnnotationRepository(ctrl)
	mockUOW := mocks.NewMockUnitOfWorkFactory(ctrl)

	mockUOW.EXPECT().WithTx(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(ctx context.Context, fn func(ctx context.Context, repos *repository.Repositories) error) error {
			return fn(ctx, &repository.Repositories{
				WorkspaceRepo:  mockWorkspaceRepo,
				PatientRepo:    mockPatientRepo,
				ImageRepo:      mockImageRepo,
				AnnotationRepo: mockAnnotationRepo,
			})
		},
	)

	pService := service.NewPatientService(mockPatientRepo, mockWorkspaceRepo, mockUOW)
	return pService, mockWorkspaceRepo, mockPatientRepo, mockImageRepo, mockAnnotationRepo, mockUOW
}

func TestCreateNewPatient_Success(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupPatientService(t)

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
	pService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupPatientService(t)

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
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)

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
	pService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupPatientService(t)

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
	pService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupPatientService(t)

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
	pService, mockWorkspaceRepo, _, _, _, _ := setupPatientService(t)

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
	pService, _, mockPatientRepo, mockImageRepo, _, _ := setupPatientService(t)

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
	pService, _, _, mockImageRepo, _, _ := setupPatientService(t)

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
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)
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
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientID := "pat-123"
	input := service.UpdatePatientInput{}

	mockPatientRepo.EXPECT().Update(ctx, patientID, gomock.Any()).Times(0)

	err := pService.UpdatePatient(ctx, patientID, input)
	require.NoError(t, err)
}

func TestListPatients_Success(t *testing.T) {
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)
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
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)
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

func TransferPatientWorkspace_Success(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupPatientService(t)

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

func TransferPatientWorkspace_WorkspaceConflictFailure(t *testing.T) {
	pService, mockWorkspaceRepo, _, _, _, _ := setupPatientService(t)

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

func CascadeDeletePatient_Success(t *testing.T) {
	pService, _, mockPatientRepo, mockImageRepo, mockAnnotationRepo, _ := setupPatientService(t)

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

	mockAnnotationRepo.EXPECT().
		FindByFilters(gomock.Any(), filter, paginationOpts).
		Return(&query.Result[*model.Annotation]{Data: []*model.Annotation{}}, nil)

	mockAnnotationRepo.EXPECT().
		BatchDelete(gomock.Any(), []string{}).
		Return(nil)

	mockPatientRepo.EXPECT().
		BatchDelete(gomock.Any(), []string{patientID}).
		Return(nil)

	mockImageRepo.EXPECT().
		BatchDelete(gomock.Any(), []string{}).
		Return(nil)

	err := pService.CascadeDelete(ctx, patientID)

	require.NoError(t, err)
}

func TestGetPatientByID_Success(t *testing.T) {
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientID := "patient-123"
	expectedPatient := &model.Patient{
		ID:   patientID,
		Name: "Test Patient",
	}

	mockPatientRepo.EXPECT().
		Read(ctx, patientID).
		Return(expectedPatient, nil)

	result, err := pService.GetPatientByID(ctx, patientID)

	require.NoError(t, err)
	assert.Equal(t, expectedPatient, result)
}

func TestGetPatientByID_NotFound(t *testing.T) {
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientID := "patient-999"

	mockPatientRepo.EXPECT().
		Read(ctx, patientID).
		Return(nil, errors.NewNotFoundError("patient not found"))

	result, err := pService.GetPatientByID(ctx, patientID)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestCascadeDelete_ComplexWithImagesAndAnnotations(t *testing.T) {
	pService, _, mockPatientRepo, mockImageRepo, mockAnnotationRepo, _ := setupPatientService(t)
	ctx := context.Background()
	patientID := "patient-complex-del"

	image1 := &model.Image{ID: "img-1", PatientID: patientID}
	image2 := &model.Image{ID: "img-2", PatientID: patientID}
	anno1 := &model.Annotation{ID: "anno-1", ImageID: "img-1"}
	anno2 := &model.Annotation{ID: "anno-2", ImageID: "img-2"}

	imageFilter := []query.Filter{{Field: constants.ImagePatientIDField, Operator: query.OpEqual, Value: patientID}}

	mockImageRepo.EXPECT().
		FindByFilters(gomock.Any(), imageFilter, &query.Pagination{Limit: 100, Offset: 0}).
		Return(&query.Result[*model.Image]{
			Data:    []*model.Image{image1, image2},
			HasMore: false,
		}, nil)

	mockAnnotationRepo.EXPECT().
		FindByFilters(gomock.Any(),
			[]query.Filter{{Field: constants.AnnotationImageIDField, Operator: query.OpEqual, Value: image1.ID}},
			&query.Pagination{Limit: 100, Offset: 0}).
		Return(&query.Result[*model.Annotation]{
			Data:    []*model.Annotation{anno1},
			HasMore: false,
		}, nil)

	mockAnnotationRepo.EXPECT().
		FindByFilters(gomock.Any(),
			[]query.Filter{{Field: constants.AnnotationImageIDField, Operator: query.OpEqual, Value: image2.ID}},
			&query.Pagination{Limit: 100, Offset: 0}).
		Return(&query.Result[*model.Annotation]{
			Data:    []*model.Annotation{anno2},
			HasMore: false,
		}, nil)

	mockAnnotationRepo.EXPECT().
		BatchDelete(gomock.Any(), gomock.InAnyOrder([]string{"anno-1", "anno-2"})).
		Return(nil)

	mockImageRepo.EXPECT().
		BatchDelete(gomock.Any(), gomock.InAnyOrder([]string{"img-1", "img-2"})).
		Return(nil)

	mockPatientRepo.EXPECT().
		Delete(gomock.Any(), patientID).
		Return(nil)

	err := pService.CascadeDelete(ctx, patientID)
	require.NoError(t, err)
}

func TestBatchDelete_Success(t *testing.T) {
	pService, _, mockPatientRepo, mockImageRepo, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientIDs := []string{"p1", "p2"}

	for _, pid := range patientIDs {
		mockImageRepo.EXPECT().
			FindByFilters(gomock.Any(),
				[]query.Filter{{Field: constants.ImagePatientIDField, Operator: query.OpEqual, Value: pid}},
				gomock.Any()).
			Return(&query.Result[*model.Image]{Data: []*model.Image{}, HasMore: false}, nil)

		mockPatientRepo.EXPECT().Delete(gomock.Any(), pid).Return(nil)
	}

	err := pService.BatchDelete(ctx, patientIDs)
	require.NoError(t, err)
}

func TestBatchDelete_FailureOnOne(t *testing.T) {
	pService, _, _, mockImageRepo, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientIDs := []string{"p1", "p2"}

	mockImageRepo.EXPECT().
		FindByFilters(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.NewInternalError("db error", nil))

	err := pService.BatchDelete(ctx, patientIDs)
	require.Error(t, err)
}

func TestBatchTransfer_Success(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientIDs := []string{"p1", "p2", "p3"}
	newWorkspaceID := "target-ws"

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), newWorkspaceID).
		Return(&model.Workspace{ID: newWorkspaceID}, nil)

	for _, pid := range patientIDs {
		mockPatientRepo.EXPECT().
			Transfer(gomock.Any(), pid, newWorkspaceID).
			Return(nil)
	}

	err := pService.BatchTransfer(ctx, patientIDs, newWorkspaceID)
	require.NoError(t, err)
}

func TestBatchTransfer_WorkspaceNotFound(t *testing.T) {
	pService, mockWorkspaceRepo, _, _, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientIDs := []string{"p1"}
	newWorkspaceID := "non-existent-ws"

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), newWorkspaceID).
		Return(nil, errors.NewNotFoundError("workspace not found"))

	err := pService.BatchTransfer(ctx, patientIDs, newWorkspaceID)
	require.Error(t, err)
}

func TestBatchTransfer_RepoFailure(t *testing.T) {
	pService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupPatientService(t)
	ctx := context.Background()
	patientIDs := []string{"p1"}
	newWorkspaceID := "target-ws"

	mockWorkspaceRepo.EXPECT().
		Read(gomock.Any(), newWorkspaceID).
		Return(&model.Workspace{ID: newWorkspaceID}, nil)

	mockPatientRepo.EXPECT().
		Transfer(gomock.Any(), "p1", newWorkspaceID).
		Return(errors.NewInternalError("db error", nil))

	err := pService.BatchTransfer(ctx, patientIDs, newWorkspaceID)
	require.Error(t, err)
}

func TestCountPatients_Success(t *testing.T) {
	pService, _, mockPatientRepo, _, _, _ := setupPatientService(t)
	ctx := context.Background()
	filters := []query.Filter{{Field: "race", Value: "Asian"}}

	mockPatientRepo.EXPECT().
		Count(ctx, filters).
		Return(int64(42), nil)

	count, err := pService.CountPatients(ctx, filters)
	require.NoError(t, err)
	assert.Equal(t, int64(42), count)
}
