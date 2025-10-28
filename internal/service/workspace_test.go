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
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

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

	// Expectation: mock the repository method
	mockWorkspaceRepo.EXPECT().
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		Return(&sharedQuery.Result[*model.Workspace]{Data: []*model.Workspace{}}, nil)

	mockWorkspaceRepo.EXPECT().
		Create(ctx, gomock.Any()).
		DoAndReturn(func(ctx context.Context, ws *model.Workspace) (*model.Workspace, error) {
			assert.Equal(t, input.Name, ws.Name)
			ws.ID = "ws-123"
			return ws, nil
		})

	// --- Act ---
	createdWS, err := wsService.CreateNewWorkspace(ctx, input)

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
		FindByFilters(ctx, gomock.Any(), gomock.Any()).
		Return(&sharedQuery.Result[*model.Workspace]{Data: []*model.Workspace{
			{Name: "Existing Workspace"},
		}}, nil)

	// --- Act ---
	createdWS, err := wsService.CreateNewWorkspace(ctx, input)

	// --- Assert ---
	require.Error(t, err)
	require.Nil(t, createdWS)
	var conflictErr *errors.Err
	require.True(t, stderrors.As(err, &conflictErr))
	assert.Equal(t, errors.ErrorTypeConflict, conflictErr.Type)
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
