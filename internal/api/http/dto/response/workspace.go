package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceResponse struct {
	ID               string    `json:"id"`
	CreatorID        string    `json:"creator_id"`
	Name             string    `json:"name"`
	OrganType        string    `json:"organ_type"`
	AnnotationTypeID *string   `json:"annotation_type_id,omitempty"`
	Organization     string    `json:"organization"`
	Description      string    `json:"description"`
	License          string    `json:"license"`
	ResourceURL      *string   `json:"resource_url,omitempty"`
	ReleaseYear      *int      `json:"release_year,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func NewWorkspaceResponse(ws *model.Workspace) *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:               ws.ID,
		CreatorID:        ws.CreatorID,
		Name:             ws.Name,
		OrganType:        ws.OrganType,
		AnnotationTypeID: ws.AnnotationTypeID,
		Organization:     ws.Organization,
		Description:      ws.Description,
		License:          ws.License,
		ResourceURL:      ws.ResourceURL,
		ReleaseYear:      ws.ReleaseYear,
		CreatedAt:        ws.CreatedAt,
		UpdatedAt:        ws.UpdatedAt,
	}
}

func NewWorkspaceListResponse(result *query.Result[*model.Workspace]) *ListResponse[WorkspaceResponse] {

	data := make([]WorkspaceResponse, len(result.Data))
	for i, ws := range result.Data {
		dto := NewWorkspaceResponse(ws)
		data[i] = *dto
	}

	pagination := PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
	}

	return &ListResponse[WorkspaceResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}

// Added DTOs for swagger responses. Swagger requires a concrete type for response schemas.
type WorkspaceDataResponse struct {
	Data WorkspaceResponse `json:"data"`
}
type WorkspaceListResponse struct {
	Data       []WorkspaceResponse `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
