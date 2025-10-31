package request

type CreateWorkspaceRequest struct {
	CreatorID        string  `json:"creator_id" binding:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name             string  `json:"name" binding:"required" example:"Lung Cancer Study"`
	OrganType        string  `json:"organ_type" binding:"required" example:"Lung"`
	AnnotationTypeID *string `json:"annotation_type_id,omitempty" binding:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Organization     string  `json:"organization" binding:"required" example:"Health Research Institute"`
	Description      string  `json:"description" binding:"required" example:"A workspace for lung cancer research."`
	License          string  `json:"license" binding:"required" example:"CC BY 4.0"`
	ResourceURL      *string `json:"resource_url,omitempty" binding:"omitempty,url" example:"https://example.com/dataset"`
	ReleaseYear      *int    `json:"release_year,omitempty" binding:"omitempty,gte=1900,lte=2100" example:"2023"`
}

type UpdateWorkspaceRequest struct {
	Name             *string `json:"name,omitempty" binding:"omitempty" example:"Lung Cancer Study"`
	OrganType        *string `json:"organ_type,omitempty" binding:"omitempty" example:"Lung"`
	Organization     *string `json:"organization,omitempty" binding:"omitempty" example:"Health Research Institute"`
	Description      *string `json:"description,omitempty" binding:"omitempty" example:"A workspace for lung cancer research."`
	License          *string `json:"license,omitempty" binding:"omitempty" example:"CC BY 4.0"`
	ResourceURL      *string `json:"resource_url,omitempty" binding:"omitempty,url" example:"https://example.com/dataset"`
	ReleaseYear      *int    `json:"release_year,omitempty" binding:"omitempty,gte=1900,lte=2100" example:"2023"`
	AnnotationTypeID *string `json:"annotation_type_id,omitempty" binding:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
}

type ListWorkspacesRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}
