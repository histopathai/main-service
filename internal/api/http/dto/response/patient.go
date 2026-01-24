package response

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/shared/query"
)

type PatientResponse struct {
	ID         string            `json:"id"`
	EntityType string            `json:"entity_type"`
	Parent     ParentRefResponse `json:"parent"`
	Name       string            `json:"name"`
	Age        *int              `json:"age,omitempty"`
	Gender     *string           `json:"gender,omitempty"`
	Race       *string           `json:"race,omitempty"`
	Disease    *string           `json:"disease,omitempty"`
	Subtype    *string           `json:"subtype,omitempty"`
	Grade      *int              `json:"grade,omitempty"`
	History    *string           `json:"history,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func NewPatientResponse(p *model.Patient) *PatientResponse {
	parent := ParentRefResponse{
		ID:   p.Parent.ID,
		Type: p.Parent.Type.String(),
	}
	return &PatientResponse{
		ID:         p.ID,
		EntityType: p.EntityType.String(),
		Parent:     parent,
		Age:        p.Age,
		Name:       p.Name,
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

// Swagger documentation helper
type PatientDataResponse struct {
	Data PatientResponse `json:"data"`
}

type PatientListResponse struct {
	Data       []PatientResponse   `json:"data"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
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
