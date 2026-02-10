package request

type CreatePatientRequest struct {
	Parent  ParentRefRequest `json:"parent" binding:"required"`
	Name    string           `json:"name" binding:"required" example:"Patient_001"`
	Age     *int             `json:"age,omitempty" example:"45"`
	Gender  *string          `json:"gender,omitempty" example:"Female"`
	Race    *string          `json:"race,omitempty" example:"Asian"`
	Disease *string          `json:"disease,omitempty" example:"Glioblastoma"`
	Subtype *string          `json:"subtype,omitempty" example:"IDH-wildtype"`
	Grade   *int             `json:"grade,omitempty" example:"3"`
	History *string          `json:"history,omitempty" example:"No prior history of cancer"`
}

type UpdatePatientRequest struct {
	CreatorID *string `json:"creator_id" binding:"omitempty" example:"1"`
	Name      *string `json:"name,omitempty" example:"Patient_001"`
	Age       *int    `json:"age,omitempty" example:"45"`
	Gender    *string `json:"gender,omitempty" example:"Female"`
	Race      *string `json:"race,omitempty" example:"Asian"`
	Disease   *string `json:"disease,omitempty" example:"Glioblastoma"`
	Subtype   *string `json:"subtype,omitempty" example:"IDH-wildtype"`
	Grade     *int    `json:"grade,omitempty" example:"3"`
	History   *string `json:"history,omitempty" example:"No prior history of cancer"`
}
