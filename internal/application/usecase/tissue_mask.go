package usecase

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type TissueMaskUseCase struct {
	uow port.UnitOfWorkFactory
	now func() time.Time
}

func NewTissueMaskUseCase(uow port.UnitOfWorkFactory) *TissueMaskUseCase {
	return &TissueMaskUseCase{uow: uow, now: time.Now}
}

// Save stores an edited mask. Any previous approval is cleared, so approved
// always means the current polygons were reviewed.
func (uc *TissueMaskUseCase) Save(ctx context.Context, cmd command.SaveTissueMaskCommand) (*model.TissueMask, error) {
	if details, ok := cmd.Validate(); !ok {
		return nil, errors.NewValidationError("invalid tissue mask", details)
	}

	var saved *model.TissueMask
	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		image, existing, err := uc.readImageAndMask(txCtx, cmd.ImageID)
		if err != nil {
			return err
		}

		now := uc.now()
		mask := newTissueMask(image, existing, cmd.UserID, cmd.TissueMaskData)
		mask.Status = vobj.TissueMaskStatusEdited
		mask.EditedBy, mask.EditedAt = &cmd.UserID, &now

		if err := uc.write(txCtx, existing, mask); err != nil {
			return err
		}
		saved = mask
		return nil
	})
	if err != nil {
		return nil, err
	}
	return saved, nil
}

// Approve marks the current mask as reviewed.
func (uc *TissueMaskUseCase) Approve(ctx context.Context, imageID, userID string) (*model.TissueMask, error) {
	if imageID == "" || userID == "" {
		return nil, errors.NewValidationError("image_id and user_id are required", nil)
	}

	var approved *model.TissueMask
	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		mask, err := uc.readMask(txCtx, imageID)
		if err != nil {
			return err
		}
		if mask == nil || mask.Deleted {
			return errors.NewNotFoundError("tissue mask not found")
		}
		if mask.Status == vobj.TissueMaskStatusApproved {
			approved = mask
			return nil
		}

		now := uc.now()
		updates := map[string]interface{}{
			fields.TissueMaskStatus.DomainName():     vobj.TissueMaskStatusApproved,
			fields.TissueMaskApprovedBy.DomainName(): &userID,
			fields.TissueMaskApprovedAt.DomainName(): &now,
		}
		if err := uc.uow.GetTissueMaskRepo().Update(txCtx, imageID, updates); err != nil {
			return errors.NewInternalError("failed to approve tissue mask", err)
		}
		mask.Status, mask.ApprovedBy, mask.ApprovedAt, mask.UpdatedAt = vobj.TissueMaskStatusApproved, &userID, &now, now
		approved = mask
		return nil
	})
	if err != nil {
		return nil, err
	}
	return approved, nil
}

// ApplyWorkerResult stores the worker's default mask unless a person has
// already edited or approved the image's mask. It reports whether it wrote.
func (uc *TissueMaskUseCase) ApplyWorkerResult(ctx context.Context, cmd command.ApplyWorkerTissueMaskCommand) (bool, error) {
	if details, ok := cmd.Validate(); !ok {
		return false, errors.NewValidationError("invalid tissue mask from worker", details)
	}

	applied := false
	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		applied = false
		image, existing, err := uc.readImageAndMask(txCtx, cmd.ImageID)
		if err != nil {
			return err
		}
		if existing != nil && !existing.Deleted && existing.Status != vobj.TissueMaskStatusAuto {
			return nil
		}

		mask := newTissueMask(image, existing, image.CreatorID, cmd.TissueMaskData)
		mask.Status = vobj.TissueMaskStatusAuto
		if err := uc.write(txCtx, existing, mask); err != nil {
			return err
		}
		applied = true
		return nil
	})
	return applied, err
}

func (uc *TissueMaskUseCase) readImageAndMask(ctx context.Context, imageID string) (*model.Image, *model.TissueMask, error) {
	image, err := uc.uow.GetImageRepo().Read(ctx, imageID)
	if err != nil {
		if isNotFound(err) {
			return nil, nil, errors.NewNotFoundError("image not found")
		}
		return nil, nil, errors.NewInternalError("failed to read image", err)
	}
	if image == nil || image.Deleted {
		return nil, nil, errors.NewNotFoundError("image not found")
	}
	mask, err := uc.readMask(ctx, imageID)
	if err != nil {
		return nil, nil, err
	}
	return image, mask, nil
}

func (uc *TissueMaskUseCase) readMask(ctx context.Context, imageID string) (*model.TissueMask, error) {
	mask, err := uc.uow.GetTissueMaskRepo().Read(ctx, imageID)
	if err != nil {
		if isNotFound(err) {
			return nil, nil
		}
		return nil, errors.NewInternalError("failed to read tissue mask", err)
	}
	return mask, nil
}

// write creates the document or replaces its mask fields, keeping created_at.
func (uc *TissueMaskUseCase) write(ctx context.Context, existing, mask *model.TissueMask) error {
	repo := uc.uow.GetTissueMaskRepo()
	if existing == nil {
		if _, err := repo.Create(ctx, mask); err != nil {
			return errors.NewInternalError("failed to create tissue mask", err)
		}
		return nil
	}
	updates := map[string]interface{}{
		"Mask":                              mask,
		fields.EntityIsDeleted.DomainName(): false,
	}
	if err := repo.Update(ctx, mask.ID, updates); err != nil {
		return errors.NewInternalError("failed to update tissue mask", err)
	}
	mask.UpdatedAt = uc.now()
	return nil
}

func newTissueMask(image *model.Image, existing *model.TissueMask, creatorID string, data command.TissueMaskData) *model.TissueMask {
	entity := vobj.Entity{
		ID:         image.ID,
		EntityType: vobj.EntityTypeTissueMask,
		Name:       image.Name,
		CreatorID:  creatorID,
		Parent:     vobj.ParentRef{ID: image.ID, Type: vobj.ParentTypeImage},
	}
	if existing != nil {
		entity.CreatorID = existing.CreatorID
		entity.CreatedAt = existing.CreatedAt
	}
	return &model.TissueMask{
		Entity:           entity,
		WsID:             image.WsID,
		AlgorithmVersion: data.AlgorithmVersion,
		Params:           data.Params,
		Polygons:         data.Polygons,
		PreviewWidth:     data.PreviewWidth,
		PreviewHeight:    data.PreviewHeight,
		Level0Width:      data.Level0Width,
		Level0Height:     data.Level0Height,
		DownsampleX:      data.DownsampleX,
		DownsampleY:      data.DownsampleY,
		TissueAreaRatio:  data.TissueAreaRatio,
	}
}

func isNotFound(err error) bool {
	var appErr *errors.Err
	return stderrors.As(err, &appErr) && appErr.Type == errors.ErrorTypeNotFound
}
