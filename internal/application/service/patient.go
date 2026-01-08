package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/commands"
	"github.com/histopathai/main-service/internal/application/usecases/common"
	"github.com/histopathai/main-service/internal/application/usecases/composite"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type PatientService struct {
	*BaseService[*model.Patient]
	transferUc *composite.TransferUseCase
}

func NewPatientService(
	patientRepo port.Repository[*model.Patient],
	uowFactory port.UnitOfWorkFactory,
) *PatientService {

	createUc := composite.NewCreateUseCase[*model.Patient](uowFactory)
	deleteUc := composite.NewDeleteUseCase(uowFactory)
	transferUc := composite.NewTransferUseCase(uowFactory)
	// updateUc := composite.NewUpdateUseCase[*model.Patient](uowFactory)

	baseSvc := NewBaseService(
		common.NewReadUseCase(patientRepo),
		common.NewListUseCase(patientRepo),
		common.NewCountUseCase(patientRepo),
		common.NewSoftDeleteUseCase(patientRepo),
		common.NewFilterUseCase(patientRepo),
		common.NewFilterByParentUseCase(patientRepo),
		common.NewFilterByCreatorUseCase(patientRepo),
		common.NewFilterByNameUseCase(patientRepo),
		deleteUc,
		createUc,
		// updateUc,
		vobj.EntityTypePatient,
	)

	return &PatientService{
		BaseService: baseSvc,
		transferUc:  transferUc,
	}
}

func (s *PatientService) Transfer(ctx context.Context, cmd commands.TransferCommand) error {
	return s.transferUc.Execute(ctx, cmd.ID, cmd.NewParentID, vobj.EntityTypePatient)
}

func (s *PatientService) TransferMany(ctx context.Context, cmd commands.TransferManyCommand) error {
	return s.transferUc.ExecuteMany(ctx, cmd.IDs, cmd.NewParentID, vobj.EntityTypePatient)
}
