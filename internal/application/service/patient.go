package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type PatientService struct {
	*Service[*model.Patient]
	usecase *usecase.PatientUseCase
}

func NewPatientService(repo port.Repository[*model.Patient], uowFactory port.UnitOfWorkFactory) *PatientService {
	return &PatientService{
		Service: &Service[*model.Patient]{
			repo:       repo,
			uowFactory: uowFactory,
		},
		usecase: usecase.NewPatientUseCase(repo, uowFactory),
	}
}

func (s *PatientService) Create(ctx context.Context, cmd any) (*model.Patient, error) {

	// Type assertion
	createCmd, ok := cmd.(command.CreatePatientCommand)
	if !ok {
		return nil, errors.NewInternalError("invalid command type for creating patient", nil)
	}

	entity, err := createCmd.ToEntity()
	if err != nil {
		return nil, err
	}

	return s.usecase.Create(ctx, entity)
}

func (s *PatientService) Update(ctx context.Context, cmd any) error {

	// Type assertion
	updateCmd, ok := cmd.(command.UpdatePatientCommand)
	if !ok {
		return errors.NewInternalError("invalid command type for updating patient", nil)
	}

	id := updateCmd.GetID()
	updates := updateCmd.GetUpdates()
	if updates == nil {
		return errors.NewValidationError("no updates provided for patient", map[string]interface{}{
			"id": id,
		})
	}

	return s.usecase.Update(ctx, id, updates)
}
