package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/helper"
	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type PatientUseCase struct {
	repo      port.PatientRepository
	uow       port.UnitOfWorkFactory
	validator *validator.PatientValidator
}

func NewPatientUseCase(repo port.PatientRepository, uow port.UnitOfWorkFactory) *PatientUseCase {
	return &PatientUseCase{
		validator: validator.NewPatientValidator(repo, uow),
		repo:      repo,
		uow:       uow,
	}
}

func (uc *PatientUseCase) Create(ctx context.Context, cmd command.CreatePatientCommand) (*model.Patient, error) {
	entity, err := cmd.ToEntity()
	if err != nil {
		return nil, errors.NewInternalError("failed to convert command to entity", err)
	}

	var createdPatient *model.Patient
	uowerr := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		patientRepo := uc.uow.GetPatientRepo()
		if err := uc.validator.ValidateCreate(txCtx, entity); err != nil {
			return err
		}

		createdPatient, err = patientRepo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create patient", err)
		}

		return nil
	})

	if uowerr != nil {
		return nil, uowerr
	}

	return createdPatient, nil
}

func (uc *PatientUseCase) Update(ctx context.Context, cmd command.UpdatePatientCommand) error {
	updates := cmd.GetUpdates()
	if updates == nil {
		return errors.NewInternalError("no updates provided", nil)
	}

	id := cmd.GetID()

	if err := uc.validator.ValidateUpdate(ctx, id, updates); err != nil {
		return err
	}

	if err := uc.repo.Update(ctx, id, updates); err != nil {
		return errors.NewInternalError("failed to update patient", err)
	}

	return nil
}

func (uc *PatientUseCase) Transfer(ctx context.Context, cmd command.TransferCommand) error {

	uowerr := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.validator.ValidateTransfer(txCtx, cmd); err != nil {
			return err
		}

		hiearachyService := helper.NewHierarchyService(uc.uow)

		childImageIDs, err := hiearachyService.GetChildIDs(txCtx, vobj.EntityTypePatient, cmd.GetID())

		if err != nil {
			return errors.NewInternalError("failed to get child image IDs", err)
		}

		annotationIDs := []string{}
		if len(childImageIDs) > 0 {
			for _, imageID := range childImageIDs {
				childAnnotationIDs, err := hiearachyService.GetChildIDs(txCtx, vobj.EntityTypeImage, imageID)
				if err != nil {
					return errors.NewInternalError("failed to get child annotation IDs", err)
				}
				annotationIDs = append(annotationIDs, childAnnotationIDs...)
			}
		}

		if len(childImageIDs) > 0 {
			updates := map[string]any{
				fields.ImageWsID.DomainName(): cmd.GetNewParent(),
			}

			imageRepo := uc.uow.GetImageRepo()
			if err := imageRepo.UpdateMany(txCtx, childImageIDs, updates); err != nil {
				return errors.NewInternalError("failed to update images", err)
			}
		}

		if len(annotationIDs) > 0 {
			updates := map[string]any{
				fields.AnnotationWsID.DomainName(): cmd.GetNewParent(),
			}
			annotationRepo := uc.uow.GetAnnotationRepo()
			if err := annotationRepo.UpdateMany(txCtx, annotationIDs, updates); err != nil {
				return errors.NewInternalError("failed to update annotations", err)
			}
		}

		if err := uc.repo.TransferMany(txCtx, []string{cmd.GetID()}, cmd.GetNewParent()); err != nil {
			return errors.NewInternalError("failed to transfer patient", err)
		}

		return nil

	})

	if uowerr != nil {
		return uowerr
	}

	return nil
}

func (uc *PatientUseCase) TransferMany(ctx context.Context, cmd command.TransferManyCommand) error {
	// Appply in channel

	ch := make(chan error)

	for _, id := range cmd.GetIDs() {
		go func(id string) {
			ch <- uc.Transfer(ctx, command.TransferCommand{
				NewParent:  cmd.GetNewParent(),
				ParentType: vobj.EntityTypePatient.String(),
				ID:         id,
			})
		}(id)
	}

	for range cmd.GetIDs() {
		if err := <-ch; err != nil {
			return err
		}
	}

	return nil
}
