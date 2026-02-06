package request

type CreateWorkspaceRequest struct {
	Name            string   `json:"name" binding:"required" example:"Lung Cancer Study"`
	OrganType       string   `json:"organ_type" binding:"required" example:"lung"`
	Organization    string   `json:"organization" binding:"required" example:"Health Research Institute"`
	Description     string   `json:"description" binding:"required" example:"A workspace for lung cancer research"`
	License         string   `json:"license" binding:"required" example:"CC BY 4.0"`
	ResourceURL     *string  `json:"resource_url,omitempty" binding:"omitempty,url" example:"https://example.com/dataset"`
	ReleaseYear     *int     `json:"release_year,omitempty" binding:"omitempty,gte=1900,lte=2100" example:"2023"`
	AnnotationTypes []string `json:"annotation_types,omitempty" binding:"omitempty,dive" example:"['550e8400-e29b-41d4-a716-446655440000']"`
}

type UpdateWorkspaceRequest struct {
	CreatorID       *string  `json:"creator_id" example:"1"`
	Name            *string  `json:"name,omitempty" example:"Lung Cancer Study Updated"`
	OrganType       *string  `json:"organ_type,omitempty" example:"lung"`
	Organization    *string  `json:"organization,omitempty" example:"Health Research Institute"`
	Description     *string  `json:"description,omitempty" example:"Updated description"`
	License         *string  `json:"license,omitempty" example:"CC BY 4.0"`
	ResourceURL     *string  `json:"resource_url,omitempty" binding:"omitempty,url" example:"https://example.com/dataset"`
	ReleaseYear     *int     `json:"release_year,omitempty" binding:"omitempty,gte=1900,lte=2100" example:"2023"`
	AnnotationTypes []string `json:"annotation_types,omitempty" binding:"omitempty,dive" example:"['550e8400-e29b-41d4-a716-446655440000']"`
}
