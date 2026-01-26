package validator

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ContentValidator struct {
	repo port.ContentRepository
	uow  port.UnitOfWorkFactory
}

func NewContentValidator(repo port.ContentRepository, uow port.UnitOfWorkFactory) *ContentValidator {
	return &ContentValidator{repo: repo, uow: uow}
}

func (v *ContentValidator) ValidateCreate(ctx context.Context, content *model.Content) error {

	imageRepo := v.uow.GetImageRepo()

	image, err := imageRepo.Read(ctx, content.Parent.ID)
	if err != nil {
		return errors.NewInternalError("failed to read image", err)
	}

	if image == nil {
		return errors.NewInternalError("image not found", nil)
	}

	if content.ContentType.IsThumbnail() {
		if image.ThumbnailContentID != nil {
			return errors.NewValidationError("image already has thumbnail content", nil)
		}
	}

	if content.ContentType.IsIndexMap() {
		if image.IndexmapContentID != nil {
			return errors.NewValidationError("image already has indexmap content", nil)
		}
	}

	if content.ContentType.IsArchive() {
		if image.ZipTilesContentID != nil {
			return errors.NewValidationError("image already has archive content", nil)
		}
	}

	if content.ContentType.IsDZI() {
		if image.DziContentID != nil {
			return errors.NewValidationError("image already has dzi content", nil)
		}
	}
	if content.ContentType.IsTiles() {
		if image.TilesContentID != nil {
			return errors.NewValidationError("image already has tiles content", nil)
		}
	}

	return nil
}
