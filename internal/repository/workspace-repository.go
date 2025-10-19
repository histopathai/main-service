package repository

import (
	"context"
	"time"

	"github.com/histopathai/models"
)

const WorkspacesCollection = "workspaces"

type WorkspaceRepository struct {
	repo Repository
}

func NewWorkspaceRepository(repo Repository) *WorkspaceRepository {
	return &WorkspaceRepository{
		repo: repo,
	}
}

func (wr *WorkspaceRepository) CreateWorkspace(ctx context.Context, workspace *models.Workspace) (string, error) {
	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = time.Now()
	return wr.repo.Create(ctx, WorkspacesCollection, workspace.ToMap())
}

func (wr *WorkspaceRepository) ReadWorkspace(ctx context.Context, workspaceID string) (*models.Workspace, error) {
	data, err := wr.repo.Read(ctx, WorkspacesCollection, workspaceID)
	if err != nil {
		return nil, err
	}
	workspace := &models.Workspace{}
	workspace.FromMap(data)
	return workspace, nil
}

func (wr *WorkspaceRepository) UpdateWorkspace(ctx context.Context, workspaceID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return wr.repo.Update(ctx, WorkspacesCollection, workspaceID, updates)
}

func (wr *WorkspaceRepository) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	return wr.repo.Delete(ctx, WorkspacesCollection, workspaceID)
}

func (wr *WorkspaceRepository) QueryWorkspaces(ctx context.Context, filters map[string]interface{}) ([]*models.Workspace, error) {
	results, err := wr.repo.Query(ctx, WorkspacesCollection, filters)
	if err != nil {
		return nil, err
	}
	workspaces := make([]*models.Workspace, 0, len(results))
	for _, data := range results {
		workspace := &models.Workspace{}
		workspace.FromMap(data)
		workspaces = append(workspaces, workspace)
	}
	return workspaces, nil
}

func (wr *WorkspaceRepository) Exists(ctx context.Context, workspaceID string) (bool, error) {
	return wr.repo.Exists(ctx, WorkspacesCollection, workspaceID)
}
