package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ContentUseCase struct {
	repo      port.ContentRepository
	uow       port.UnitOfWorkFactory
	validator *validator.ContentValidator
}

func NewContentUseCase(
	repo port.ContentRepository,
	uow port.UnitOfWorkFactory,
) *ContentUseCase {
	return &ContentUseCase{
		repo:      repo,
		uow:       uow,
		validator: validator.NewContentValidator(repo, uow),
	}
}

func (uc *ContentUseCase) Upload(ctx context.Context, cmd command.UploadContentCommand) (*model.Content, error) {
	var createdContent *model.Content

	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {

		content, err := cmd.ToEntity()
		if err != nil {
			return err
		}

		// Validate
		if err := uc.validator.ValidateCreate(txCtx, content); err != nil {
			return err
		}

		updates := map[string]interface{}{}
		if content.ContentType.IsThumbnail() {
			updates[fields.ImageThumbnailContentID.DomainName()] = content.ID
		} else if content.ContentType.IsIndexMap() {
			updates[fields.ImageIndexmapContentID.DomainName()] = content.ID
		} else if content.ContentType.IsArchive() {
			updates[fields.ImageZipTilesContentID.DomainName()] = content.ID
		} else if content.ContentType.IsDZI() {
			updates[fields.ImageDziContentID.DomainName()] = content.ID
		} else {
			return errors.NewInternalError("failed to define content in parent", nil)
		}

		// Create content
		created, err := uc.repo.Create(txCtx, content)
		if err != nil {
			return err
		}

		// Update status to uploaded

		if err := uc.uow.GetImageRepo().Update(txCtx, content.Parent.ID, updates); err != nil {
			return errors.NewInternalError("failed to update parent", err)
		}

		createdContent = created

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdContent, nil
}
