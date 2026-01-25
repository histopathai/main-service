package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PatientResponse struct {
	ID         string             `json:"id" example:"patient-123"`
	EntityType string             `json:"entity_type" example:"patient"`
	CreatorID  string             `json:"creator_id" example:"user-123"`
	Parent     *ParentRefResponse `json:"parent,omitempty"`
	Name       string             `json:"name" example:"Patient_001"`
	Age        *int               `json:"age,omitempty" example:"45"`
	Gender     *string            `json:"gender,omitempty" example:"Female"`
	Race       *string            `json:"race,omitempty" example:"Asian"`
	Disease    *string            `json:"disease,omitempty" example:"Glioblastoma"`
	Subtype    *string            `json:"subtype,omitempty" example:"IDH-wildtype"`
	Grade      *int               `json:"grade,omitempty" example:"3"`
	History    *string            `json:"history,omitempty" example:"No prior history"`
	CreatedAt  time.Time          `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt  time.Time          `json:"updated_at" example:"2024-01-02T12:00:00Z"`
}

func NewPatientResponse(p *model.Patient) *PatientResponse {
	return &PatientResponse{
		ID:         p.ID,
		EntityType: p.EntityType.String(),
		CreatorID:  p.CreatorID,
		Parent:     NewParentRefResponse(&p.Parent),
		Name:       p.Name,
		Age:        p.Age,
		Gender:     p.Gender,
		Race:       p.Race,
		Disease:    p.Disease,
		Subtype:    p.Subtype,
		Grade:      p.Grade,
		History:    p.History,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func NewPatientListResponse(result *query.Result[*model.Patient]) *ListResponse[PatientResponse] {
	data := make([]PatientResponse, len(result.Data))
	for i, p := range result.Data {
		data[i] = *NewPatientResponse(p)
	}

	return &ListResponse[PatientResponse]{
		Data: data,
		Pagination: &PaginationResponse{
			Limit:   result.Limit,
			Offset:  result.Offset,
			HasMore: result.HasMore,
		},
	}
}

// Swagger docs
type PatientDataResponse struct {
	Data PatientResponse `json:"data"`
}

type PatientListResponseDoc struct {
	Data       []PatientResponse   `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}
