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

func (wr *WorkspaceRepository) ListWorkspaces(ctx context.Context, filters []Filter, pagination Pagination) (*WorkspaceQueryResult, error) {

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

func (wr *WorkspaceRepository) GetWorkspaceByName(ctx context.Context, name string, pagination Pagination) (*WorkspaceQueryResult, error) {
	filters := []Filter{
		{
			Field: "name",
			Op:    OpEqual,
			Value: name,
		},
	}
	return wr.ListWorkspaces(ctx, filters, pagination)
}
func (wr *WorkspaceRepository) GetWorkspaceByCreator(ctx context.Context, creatorID string, pagination Pagination) (*WorkspaceQueryResult, error) {
	filters := []Filter{
		{
			Field: "creator_id",
			Op:    OpEqual,
			Value: creatorID,
		},
	}
	return wr.ListWorkspaces(ctx, filters, pagination)
}

func (wr *WorkspaceRepository) GetWorkspaceByOrganType(ctx context.Context, organType string, pagination Pagination) (*WorkspaceQueryResult, error) {
	filters := []Filter{
		{
			Field: "organ_type",
			Op:    OpEqual,
			Value: organType,
		},
	}
	return wr.ListWorkspaces(ctx, filters, pagination)
}

func (wr *WorkspaceRepository) GetWorkspaceByOrganization(ctx context.Context, organizationID string, pagination Pagination) (*WorkspaceQueryResult, error) {
	filters := []Filter{
		{
			Field: "organization",
			Op:    OpEqual,
			Value: organizationID,
		},
	}
	return wr.ListWorkspaces(ctx, filters, pagination)
}

func (wr *WorkspaceRepository) GetAllWorkspaces(ctx context.Context, pagination Pagination) (*WorkspaceQueryResult, error) {
	filters := []Filter{}
	return wr.ListWorkspaces(ctx, filters, pagination)
}
