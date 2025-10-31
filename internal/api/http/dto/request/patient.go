package request

type CreatePatientRequest struct {
	WorkspaceID string  `json:"workspace_id" binding:"required,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string  `json:"name" binding:"required" example:"Patient_001"`
	Age         *int    `json:"age,omitempty" example:"45"`
	Gender      *string `json:"gender,omitempty" example:"Female"`
	Race        *string `json:"race,omitempty" example:"Asian"`
	Disease     *string `json:"disease,omitempty" example:"Glioblastoma"`
	Subtype     *string `json:"subtype,omitempty" example:"IDH-wildtype"`
	Grade       *int    `json:"grade,omitempty" example:"3"`
	History     *string `json:"history,omitempty" example:"No prior history of cancer."`
}

type UpdatePatientRequest struct {
	WorkspaceID *string `json:"workspace_id,omitempty" binding:"omitempty,uuid4" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        *string `json:"name,omitempty" example:"Patient_001"`
	Age         *int    `json:"age,omitempty" example:"45"`
	Gender      *string `json:"gender,omitempty" example:"Female"`
	Race        *string `json:"race,omitempty" example:"Asian"`
	Disease     *string `json:"disease,omitempty" example:"Glioblastoma"`
	Subtype     *string `json:"subtype,omitempty" example:"IDH-wildtype"`
	Grade       *int    `json:"grade,omitempty" example:"3"`
	History     *string `json:"history,omitempty" example:"No prior history of cancer."`
}

type ListPatientsRequest struct {
	Filters    []JSONFilterRequest   `json:"filters,omitempty" binding:"omitempty,dive"`
	Pagination JSONPaginationRequest `json:"pagination"`
}
