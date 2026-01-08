package service

import (
	"context"

	"github.com/histopathai/main-service/internal/application/commands"
	"github.com/histopathai/main-service/internal/application/usecases/common"
	"github.com/histopathai/main-service/internal/application/usecases/composite"
	entityspecific "github.com/histopathai/main-service/internal/application/usecases/entity-specific"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type ImageService struct {
	*BaseService[model.Image]
	transferUc *composite.TransferUseCase
}

func NewImageService(
	uow port.UnitOfWorkFactory,
	imageRepo port.Repository[model.Image],
	deleteUc *composite.DeleteUseCase,
	transferUc *composite.TransferUseCase,
) *ImageService {

	createUc := entityspecific.NewCreateImageUseCase(uow)
	updateUc := entityspecific.NewUpdateImageUseCase(uow)

	baseSvc := NewBaseService(
		common.NewReadUseCase(imageRepo),
		common.NewListUseCase(imageRepo),
		common.NewCountUseCase(imageRepo),
		common.NewSoftDeleteUseCase(imageRepo),
		common.NewFilterUseCase(imageRepo),
		common.NewFilterByParentUseCase(imageRepo),
		common.NewFilterByCreatorUseCase(imageRepo),
		common.NewFilterByNameUseCase(imageRepo),
		deleteUc,
		createUc,
		updateUc,
		vobj.EntityTypeImage,
	)

	return &ImageService{
		BaseService: baseSvc,
		transferUc:  transferUc,
	}
}

func (s *ImageService) Transfer(ctx context.Context, cmd commands.TransferCommand) error {
	return s.transferUc.Execute(ctx, cmd.ID, cmd.NewParentID, vobj.EntityTypeImage)
}

func (s *ImageService) TransferMany(ctx context.Context, cmd commands.TransferManyCommand) error {
	return s.transferUc.ExecuteMany(ctx, cmd.IDs, cmd.NewParentID, vobj.EntityTypeImage)
}
