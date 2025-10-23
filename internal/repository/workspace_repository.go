package repository

import (
	"context"
	"time"

	"github.com/histopathai/models"
)

const WorkspacesCollection = "workspaces"

type WorkspaceQueryResult struct {
	Workspaces []*models.Workspace
	Total      int
	Limit      int
	Offset     int
	HasMore    bool
}

type WorkspaceRepository struct {
	repo *MainRepository
}

func NewWorkspaceRepository(repo *MainRepository) *WorkspaceRepository {
	return &WorkspaceRepository{
		repo: repo,
	}
}

func (wr *WorkspaceRepository) GetMainRepository() *MainRepository {
	return wr.repo
}
func (wr *WorkspaceRepository) Create(ctx context.Context, workspace *models.Workspace) (string, error) {

	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = time.Now()
	return wr.repo.Create(ctx, WorkspacesCollection, workspace.ToMap())

}

func (wr *WorkspaceRepository) Read(ctx context.Context, workspaceID string) (*models.Workspace, error) {
	data, err := wr.repo.Read(ctx, WorkspacesCollection, workspaceID)
	if err != nil {
		return nil, err
	}
	workspace := &models.Workspace{}
	workspace.FromMap(data)
	return workspace, nil
}

func (wr *WorkspaceRepository) Update(ctx context.Context, workspaceID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return wr.repo.Update(ctx, WorkspacesCollection, workspaceID, updates)
}

func (wr *WorkspaceRepository) Delete(ctx context.Context, workspaceID string) error {
	return wr.repo.Delete(ctx, WorkspacesCollection, workspaceID)
}

func (wr *WorkspaceRepository) List(ctx context.Context, filters []Filter, pagination Pagination) (*WorkspaceQueryResult, error) {

	result, err := wr.repo.List(ctx, WorkspacesCollection, filters, pagination)
	if err != nil {
		return nil, err
	}
	workspaces := make([]*models.Workspace, 0, len(result.Data))
	for _, data := range result.Data {
		workspace := &models.Workspace{}
		workspace.FromMap(data)
		workspaces = append(workspaces, workspace)
	}

	return &WorkspaceQueryResult{
		Workspaces: workspaces,
		Total:      result.Total,
		Limit:      result.Limit,
		Offset:     result.Offset,
		HasMore:    result.HasMore,
	}, nil
}

func (wr *WorkspaceRepository) Exists(ctx context.Context, workspaceID string) (bool, error) {
	return wr.repo.Exists(ctx, WorkspacesCollection, workspaceID)
}
