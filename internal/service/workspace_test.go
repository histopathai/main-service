package service_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func ptrString(s string) *string {
	return &s
}

func setupWorkspaceService(t *testing.T) (
	*service.WorkspaceService,
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
				PatientRepo:   mockPatientRepo,
			})
		},
	)

	wsService := service.NewWorkspaceService(mockWorkspaceRepo, mockUOW)
	return wsService, mockWorkspaceRepo, mockPatientRepo, mockUOW
}

func TestCreateNewWorkspace_Success(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _ := setupWorkspaceService(t)

	ctx := context.Background()

	input := service.CreateWorkspaceInput{
		CreatorID:    "user-123",
		Name:         "Test Workspace",
		OrganType:    "Liver",
		Organization: "Test Org",
		Description:  "A workspace for testing",
		License:      "CC-BY",
	}

	mockWorkspaceRepo.EXPECT().
		FindByName(ctx, input.Name).
		Return(nil, nil)
	mockWorkspaceRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, ws *model.Workspace) (*model.Workspace, error) {
			assert.Equal(t, input.Name, ws.Name)
			ws.ID = "ws-123"
			return ws, nil
		})

	// --- Act ---
	createdWS, err := wsService.CreateNewWorkspace(ctx, &input)

	// --- Assert ---
	assert.NoError(t, err)
	assert.NotNil(t, createdWS)
	assert.Equal(t, "ws-123", createdWS.ID)
	assert.Equal(t, input.Name, createdWS.Name)
}

func TestCreateNewWorkspace_Conflict(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _ := setupWorkspaceService(t)
	ctx := context.Background()

	input := service.CreateWorkspaceInput{
		Name: "Existing Workspace",
	}

	// Expectation: mock the repository method to return an existing workspace
	mockWorkspaceRepo.EXPECT().
		FindByName(ctx, input.Name).
		Return(&model.Workspace{
			ID:   "ws-456",
			Name: input.Name,
		}, nil)

	// --- Act ---
	createdWS, err := wsService.CreateNewWorkspace(ctx, &input)

	// --- Assert ---
	require.Error(t, err)
	require.Nil(t, createdWS)
	var conflictErr *errors.Err
	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)

}

func TestDeleteWorkspace_Success_NoPatients(t *testing.T) {

	wsService, mockWorkspaceRepo, mockPatientRepo, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "workspace-123"

	// Expectation: mock the repository methods
	mockPatientRepo.EXPECT().
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		Return(&sharedQuery.Result[*model.Patient]{Data: []*model.Patient{}}, nil)

	mockWorkspaceRepo.EXPECT().
		Delete(ctx, workspaceID).
		Return(nil)

	// --- Act ---
	err := wsService.DeleteWorkspace(ctx, workspaceID)

	// --- Assert ---
	assert.NoError(t, err)
}

func TestDeleteWorkspace_Failure_HasPatients(t *testing.T) {

	wsService, _, mockPatientRepo, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "workspace-123"

	// Expectation: mock the repository methods
	mockPatientRepo.EXPECT().
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		Return(&sharedQuery.Result[*model.Patient]{Data: []*model.Patient{
			{ID: "patient-1"},
		}}, nil)

	// --- Act ---
	err := wsService.DeleteWorkspace(ctx, workspaceID)

	// --- Assert ---
	require.Error(t, err)
	var conflictErr *errors.Err
	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)
}

func TestUpdateWorkspace_Success(t *testing.T) {

	wsService, mockWorkspaceRepo, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "workspace-123"

	mockWorkspaceRepo.EXPECT().
		FindByName(ctx, "Updated Workspace").
		Return(nil, nil)
	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(&model.Workspace{}, nil)

	mockWorkspaceRepo.EXPECT().
		Update(ctx, workspaceID, gomock.Any()).
		Return(nil)

	input := service.UpdateWorkspaceInput{
		Name:        ptrString("Updated Workspace"),
		OrganType:   ptrString("Heart"),
		Description: ptrString("Updated description"),
	}

	// --- Act ---
	err := wsService.UpdateWorkspace(ctx, workspaceID, input)

	// --- Assert ---
	assert.NoError(t, err)
}

func TestUpdateWorkspace_Failure_IDNotFound(t *testing.T) {

	wsService, mockWorkspaceRepo, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "nonexistent-workspace"

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(nil, errors.NewNotFoundError("workspace not found"))

	input := service.UpdateWorkspaceInput{
		Name: ptrString("Updated Workspace"),
	}

	// --- Act ---
	err := wsService.UpdateWorkspace(ctx, workspaceID, input)
	// --- Assert ---
	require.Error(t, err)
	var notFoundErr *errors.Err
	require.True(t, stderrors.As(err, &notFoundErr))
	assert.Equal(t, errors.ErrorTypeNotFound, notFoundErr.Type)
}

func TestUpdateWorkspace_Failure_NameConflict(t *testing.T) {

	wsService, mockWorkspaceRepo, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "workspace-123"

	mockWorkspaceRepo.EXPECT().
		FindByName(ctx, "Conflicting Name").
		Return(&model.Workspace{
			ID:   "ws-456",
			Name: "Conflicting Name",
		}, nil)

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(&model.Workspace{}, nil)

	input := service.UpdateWorkspaceInput{
		Name: ptrString("Conflicting Name"),
	}
	// --- Act ---
	err := wsService.UpdateWorkspace(ctx, workspaceID, input)

	// --- Assert ---
	require.Error(t, err)
	var conflictErr *errors.Err
	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)
}

func TestListWorkspaces_Success(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	pagination := &sharedQuery.Pagination{Limit: 10}

	mockWorkspaceRepo.EXPECT().
		FindByFilters(ctx, []sharedQuery.Filter{}, pagination).
		Return(&sharedQuery.Result[*model.Workspace]{Data: []*model.Workspace{
			{ID: "ws-1"},
		}}, nil)

	result, err := wsService.ListWorkspaces(ctx, pagination)
	require.NoError(t, err)
	require.Len(t, result.Data, 1)
}

func TestGetWorkspaceByID_NotFound(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "ws-not-found"

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(nil, errors.NewNotFoundError("not found"))

	ws, err := wsService.GetWorkspaceByID(ctx, workspaceID)
	require.Error(t, err)
	require.Nil(t, ws)
}
