package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
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

	// ... existing methods override ?
	// The view showed Create and Update. Be careful not to remove them.
	// I will append new methods.
	return s.usecase.Update(ctx, id, updates)
}

func (s *PatientService) Transfer(ctx context.Context, cmd command.TransferCommand) error {
	// Validate command if needed
	if _, ok := cmd.Validate(); !ok {
		// details handling?
		return errors.NewValidationError("invalid transfer command", nil)
	}

	newParentType, err := vobj.NewParentTypeFromString(cmd.ParentType)
	if err != nil {
		return err
	}
	parentRef := vobj.ParentRef{
		ID:   cmd.NewParent,
		Type: newParentType,
	}

	for _, id := range cmd.IDs {
		err := s.usecase.Transfer(ctx, id, parentRef)
		if err != nil {
			return err
		}
	}
	return nil
}
