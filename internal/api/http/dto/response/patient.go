package response

import (
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/shared/query"
)

type PatientResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Age       *int      `json:"age,omitempty"`
	Gender    *string   `json:"gender,omitempty"`
	Race      *string   `json:"race,omitempty"`
	Disease   *string   `json:"disease,omitempty"`
	Subtype   *string   `json:"subtype,omitempty"`
	Grade     *int      `json:"grade,omitempty"`
	History   *string   `json:"history,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewPatientResponse(p *model.Patient) *PatientResponse {
	return &PatientResponse{
		ID:        p.ID,
		Name:      p.Name,
		Age:       p.Age,
		Gender:    p.Gender,
		Race:      p.Race,
		Disease:   p.Disease,
		Subtype:   p.Subtype,
		Grade:     p.Grade,
		History:   p.History,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
func NewPatientListResponse(result *query.Result[model.Patient]) *ListResponse[PatientResponse] {

	data := make([]PatientResponse, len(result.Data))
	for i, p := range result.Data {
		dto := NewPatientResponse(&p)
		data[i] = *dto
	}

	pagination := PaginationResponse{
		Limit:   result.Limit,
		Offset:  result.Offset,
		HasMore: result.HasMore,
		Total:   result.Total,
	}

	return &ListResponse[PatientResponse]{
		Data:       data,
		Pagination: &pagination,
	}
}
