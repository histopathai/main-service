package service_test

import (
	"context"
	stderrors "errors"
	"testing"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/mocks"
	"github.com/histopathai/main-service/internal/service"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func ptrString(s string) *string {
	return &s
}
func ptrInt(i int) *int {
	return &i
}

func setupWorkspaceService(t *testing.T) (
	*service.WorkspaceService,
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
		func(ctx context.Context, fn func(ctx context.Context, repos *port.Repositories) error) error {
			return fn(ctx, &port.Repositories{
				WorkspaceRepo:  mockWorkspaceRepo,
				PatientRepo:    mockPatientRepo,
				ImageRepo:      mockImageRepo,
				AnnotationRepo: mockAnnotationRepo,
			})
		},
	)

	wsService := service.NewWorkspaceService(mockWorkspaceRepo, mockUOW)
	return wsService, mockWorkspaceRepo, mockPatientRepo, mockImageRepo, mockAnnotationRepo, mockUOW
}

func TestCreateNewWorkspace_Success(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)

	ctx := context.Background()

	input := port.CreateWorkspaceInput{
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
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()

	input := port.CreateWorkspaceInput{
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

	wsService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupWorkspaceService(t)
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

	wsService, _, mockPatientRepo, _, _, _ := setupWorkspaceService(t)
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

	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "workspace-123"

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(&model.Workspace{}, nil)

	mockWorkspaceRepo.EXPECT().
		Update(ctx, workspaceID, gomock.Any()).
		Return(nil)

	input := port.UpdateWorkspaceInput{
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

	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "nonexistent-workspace"

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(nil, errors.NewNotFoundError("workspace not found"))

	input := port.UpdateWorkspaceInput{
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

func TestListWorkspaces_Success(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
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
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "ws-not-found"

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(nil, errors.NewNotFoundError("not found"))

	ws, err := wsService.GetWorkspaceByID(ctx, workspaceID)
	require.Error(t, err)
	require.Nil(t, ws)
}

func TestCreateNewWorkspace_InvalidYear(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()

	input := port.CreateWorkspaceInput{
		Name:        "Test Workspace",
		ReleaseYear: ptrInt(1800),
	}

	mockWorkspaceRepo.EXPECT().
		FindByName(ctx, input.Name).
		Return(nil, nil)

	createdWS, err := wsService.CreateNewWorkspace(ctx, &input)

	require.Error(t, err)
	assert.Nil(t, createdWS)
	var valErr *errors.Err
	require.True(t, stderrors.As(err, &valErr))
	assert.Equal(t, errors.ErrorTypeValidation, valErr.Type)
}

func TestGetWorkspaceByID_Success(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "ws-123"

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(&model.Workspace{ID: workspaceID, Name: "Found"}, nil)

	ws, err := wsService.GetWorkspaceByID(ctx, workspaceID)
	require.NoError(t, err)
	assert.NotNil(t, ws)
	assert.Equal(t, workspaceID, ws.ID)
}

func TestUpdateWorkspace_NoUpdates(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "ws-123"

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(&model.Workspace{ID: workspaceID}, nil)

	err := wsService.UpdateWorkspace(ctx, workspaceID, port.UpdateWorkspaceInput{})
	require.NoError(t, err)
}

func TestUpdateWorkspace_InvalidYear(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "ws-123"
	invalidYear := 2200

	mockWorkspaceRepo.EXPECT().
		Read(ctx, workspaceID).
		Return(&model.Workspace{ID: workspaceID}, nil)

	err := wsService.UpdateWorkspace(ctx, workspaceID, port.UpdateWorkspaceInput{
		ReleaseYear: &invalidYear,
	})
	require.Error(t, err)
}

func TestCascadeDeleteWorkspace_Complex(t *testing.T) {
	wsService, mockWorkspaceRepo, mockPatientRepo, mockImageRepo, mockAnnotationRepo, _ := setupWorkspaceService(t)
	ctx := context.Background()
	workspaceID := "ws-complex"

	mockWorkspaceRepo.EXPECT().Read(ctx, workspaceID).Return(&model.Workspace{ID: workspaceID}, nil)

	patient1 := &model.Patient{ID: "p1"}
	patient2 := &model.Patient{ID: "p2"}

	mockPatientRepo.EXPECT().
		FindByFilters(gomock.Any(),
			[]query.Filter{{Field: constants.PatientWorkspaceIDField, Operator: query.OpEqual, Value: workspaceID}},
			&query.Pagination{Limit: 100, Offset: 0}).
		Return(&query.Result[*model.Patient]{Data: []*model.Patient{patient1}, HasMore: true}, nil)

	mockPatientRepo.EXPECT().
		FindByFilters(gomock.Any(),
			[]query.Filter{{Field: constants.PatientWorkspaceIDField, Operator: query.OpEqual, Value: workspaceID}},
			&query.Pagination{Limit: 100, Offset: 100}).
		Return(&query.Result[*model.Patient]{Data: []*model.Patient{patient2}, HasMore: false}, nil)

	image1 := &model.Image{ID: "img1"}
	mockImageRepo.EXPECT().
		FindByFilters(gomock.Any(),
			[]query.Filter{{Field: constants.ImagePatientIDField, Operator: query.OpEqual, Value: patient1.ID}},
			gomock.Any()).
		Return(&query.Result[*model.Image]{Data: []*model.Image{image1}, HasMore: false}, nil)

	mockImageRepo.EXPECT().
		FindByFilters(gomock.Any(),
			[]query.Filter{{Field: constants.ImagePatientIDField, Operator: query.OpEqual, Value: patient2.ID}},
			gomock.Any()).
		Return(&query.Result[*model.Image]{Data: []*model.Image{}, HasMore: false}, nil)

	anno1 := &model.Annotation{ID: "anno1"}
	mockAnnotationRepo.EXPECT().
		FindByFilters(gomock.Any(),
			[]query.Filter{{Field: constants.AnnotationImageIDField, Operator: query.OpEqual, Value: image1.ID}},
			gomock.Any()).
		Return(&query.Result[*model.Annotation]{Data: []*model.Annotation{anno1}, HasMore: false}, nil)

	mockAnnotationRepo.EXPECT().BatchDelete(gomock.Any(), []string{"anno1"}).Return(nil)
	mockImageRepo.EXPECT().BatchDelete(gomock.Any(), []string{"img1"}).Return(nil)
	mockPatientRepo.EXPECT().BatchDelete(gomock.Any(), []string{"p1", "p2"}).Return(nil)
	mockWorkspaceRepo.EXPECT().Delete(gomock.Any(), workspaceID).Return(nil)

	err := wsService.CascadeDeleteWorkspace(ctx, workspaceID)
	require.NoError(t, err)
}

func TestBatchDeleteWorkspaces_Success(t *testing.T) {
	wsService, mockWorkspaceRepo, mockPatientRepo, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	ids := []string{"ws-1", "ws-2"}

	for _, id := range ids {
		mockWorkspaceRepo.EXPECT().Read(ctx, id).Return(&model.Workspace{ID: id}, nil)

		mockPatientRepo.EXPECT().
			FindByFilters(gomock.Any(),
				[]query.Filter{{Field: constants.PatientWorkspaceIDField, Operator: query.OpEqual, Value: id}},
				gomock.Any()).
			Return(&query.Result[*model.Patient]{Data: []*model.Patient{}, HasMore: false}, nil)

		mockWorkspaceRepo.EXPECT().Delete(gomock.Any(), id).Return(nil)
	}

	err := wsService.BatchDeleteWorkspaces(ctx, ids)
	require.NoError(t, err)
}

func TestBatchDeleteWorkspaces_Failure(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	ids := []string{"ws-1"}

	mockWorkspaceRepo.EXPECT().Read(ctx, "ws-1").Return(nil, errors.NewInternalError("db fail", nil))

	err := wsService.BatchDeleteWorkspaces(ctx, ids)
	require.Error(t, err)
}

func TestCountWorkspaces_Success(t *testing.T) {
	wsService, mockWorkspaceRepo, _, _, _, _ := setupWorkspaceService(t)
	ctx := context.Background()
	filters := []query.Filter{{Field: "organ_type", Value: "Lung"}}

	mockWorkspaceRepo.EXPECT().
		Count(ctx, filters).
		Return(int64(5), nil)

	count, err := wsService.CountWorkspaces(ctx, filters)
	require.NoError(t, err)
	assert.Equal(t, int64(5), count)
}
