package validator

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/helper"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ImageValidator struct {
	repo port.ImageRepository
	uow  port.UnitOfWorkFactory
}

func NewImageValidator(repo port.ImageRepository, uow port.UnitOfWorkFactory) *ImageValidator {
	return &ImageValidator{repo: repo, uow: uow}
}

func (v *ImageValidator) ValidateCreate(ctx context.Context, image *model.Image) error {
	if err := helper.CheckParentExists(ctx, &image.Parent, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}
	// Check WSID exists
	if err := helper.CheckParentExists(ctx, &vobj.ParentRef{ID: image.Parent.ID, Type: vobj.ParentTypeWorkspace}, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}

	return nil

}

func (v *ImageValidator) ValidateTransfer(ctx context.Context, command *command.TransferCommand) error {
	if err := helper.CheckParentExists(ctx, &vobj.ParentRef{ID: command.GetNewParent(), Type: vobj.ParentTypeWorkspace}, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}
	if err := helper.CheckParentExists(ctx, &vobj.ParentRef{ID: command.GetOldParent(), Type: vobj.ParentTypeWorkspace}, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}
	return nil
}
