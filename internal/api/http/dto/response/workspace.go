package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

// WorkspaceResponse - Single workspace
type WorkspaceResponse struct {
	ID              string             `json:"id" example:"ws-123"`
	EntityType      string             `json:"entity_type" example:"workspace"`
	CreatorID       string             `json:"creator_id" example:"user-123"`
	Parent          *ParentRefResponse `json:"parent,omitempty"`
	Name            string             `json:"name" example:"Lung Cancer Study"`
	OrganType       string             `json:"organ_type" example:"lung"`
	Organization    string             `json:"organization" example:"Health Research Institute"`
	Description     string             `json:"description" example:"Research workspace"`
	License         string             `json:"license" example:"CC BY 4.0"`
	ResourceURL     *string            `json:"resource_url,omitempty" example:"https://example.com"`
	ReleaseYear     *int               `json:"release_year,omitempty" example:"2023"`
	AnnotationTypes []string           `json:"annotation_types,omitempty"`
	CreatedAt       time.Time          `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt       time.Time          `json:"updated_at" example:"2024-01-02T12:00:00Z"`
}

func NewWorkspaceResponse(ws *model.Workspace) *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:              ws.ID,
		EntityType:      ws.EntityType.String(),
		CreatorID:       ws.CreatorID,
		Parent:          NewParentRefResponse(&ws.Parent),
		Name:            ws.Name,
		OrganType:       ws.OrganType.String(),
		Organization:    ws.Organization,
		Description:     ws.Description,
		License:         ws.License,
		ResourceURL:     ws.ResourceURL,
		ReleaseYear:     ws.ReleaseYear,
		AnnotationTypes: ws.AnnotationTypes,
		CreatedAt:       ws.CreatedAt,
		UpdatedAt:       ws.UpdatedAt,
	}
}

func NewWorkspaceListResponse(result *query.Result[*model.Workspace]) *ListResponse[WorkspaceResponse] {
	data := make([]WorkspaceResponse, len(result.Data))
	for i, ws := range result.Data {
		data[i] = *NewWorkspaceResponse(ws)
	}

	return &ListResponse[WorkspaceResponse]{
		Data: data,
		Pagination: &PaginationResponse{
			Limit:   result.Limit,
			Offset:  result.Offset,
			HasMore: result.HasMore,
		},
	}
}

// ============================================================================
// Swagger Documentation Types (concrete types for swagger)
// ============================================================================

// WorkspaceDataResponse - For swagger documentation
type WorkspaceDataResponse struct {
	Data WorkspaceResponse `json:"data"`
}

// WorkspaceListResponseDoc - For swagger documentation
type WorkspaceListResponseDoc struct {
	Data       []WorkspaceResponse `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
