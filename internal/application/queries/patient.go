package queries

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type PatientQuery struct {
	*BaseQuery[*model.Patient]
	*HierarchicalQueries[*model.Patient]
}

func NewPatientQuery(repo port.PatientRepository) *PatientQuery {
	return &PatientQuery{
		BaseQuery: &BaseQuery[*model.Patient]{
			repo: repo,
		},
		HierarchicalQueries: &HierarchicalQueries[*model.Patient]{
			repo: repo,
		},
	}
}
