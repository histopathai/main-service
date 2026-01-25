package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ContentUseCase struct {
	repo port.Repository[*model.Content]
	uow  port.UnitOfWorkFactory
}

func NewContentUseCase(
	repo port.Repository[*model.Content],
	uow port.UnitOfWorkFactory,
) *ContentUseCase {
	return &ContentUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *ContentUseCase) Create(ctx context.Context, content *model.Content) (*model.Content, error) {
	var createdContent *model.Content

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {

		// Validate parent exists
		if err := CheckParentExists(txCtx, &content.Parent, uc.uow); err != nil {
			return errors.NewValidationError("parent validation failed", map[string]interface{}{
				"parent_type": content.GetParent().Type,
				"parent_id":   content.GetParent().ID,
				"error":       err.Error(),
			})
		}

		// Create content
		contentRepo := repos[vobj.EntityTypeContent].(port.Repository[*model.Content])
		created, err := contentRepo.Create(txCtx, content)
		if err != nil {
			return err
		}

		// Define in Parent
		parentRepo := uc.uow.GetImageRepo()

		updates := map[string]interface{}{}
		if created.ContentType.IsThumbnail() {
			updates[fields.ImageThumbnailContentID.DomainName()] = created.ID
		} else if created.ContentType.IsIndexMap() {
			updates[fields.ImageIndexmapContentID.DomainName()] = created.ID
		} else if created.ContentType.IsArchive() {
			updates[fields.ImageZipTilesContentID.DomainName()] = created.ID
		} else if created.ContentType.IsDZI() {
			updates[fields.ImageDziContentID.DomainName()] = created.ID
		} else {
			return errors.NewInternalError("failed to define content in parent", nil)
		}

		if err := parentRepo.Update(txCtx, content.Parent.ID, updates); err != nil {
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

func (uc *ContentUseCase) Update(ctx context.Context, id string, updates map[string]interface{}) error {

	err := uc.repo.Update(ctx, id, updates)
	if err != nil {
		return errors.NewInternalError("failed to update content", err)
	}

	return nil
}
