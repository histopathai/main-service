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

// Save stores an edited mask. Approval and rejection are cleared, so approved
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
		if err := checkRevision(existing, cmd.ExpectedRevision); err != nil {
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

// Approve marks the current mask as reviewed. Approving an approved mask
// changes nothing.
func (uc *TissueMaskUseCase) Approve(ctx context.Context, cmd command.ReviewTissueMaskCommand) (*model.TissueMask, error) {
	return uc.review(ctx, cmd, func(mask *model.TissueMask, now time.Time) bool {
		if mask.Status == vobj.TissueMaskStatusApproved {
			return false
		}
		mask.Status = vobj.TissueMaskStatusApproved
		mask.ApprovedBy, mask.ApprovedAt = &cmd.UserID, &now
		mask.RejectedBy, mask.RejectedAt, mask.RejectReason = nil, nil, nil
		return true
	})
}

// Reject marks the image as unusable for tissue-based work. Rejecting again
// only updates the reason.
func (uc *TissueMaskUseCase) Reject(ctx context.Context, cmd command.ReviewTissueMaskCommand) (*model.TissueMask, error) {
	if details, ok := cmd.Validate(); !ok {
		return nil, errors.NewValidationError("invalid rejection", details)
	}
	return uc.review(ctx, cmd, func(mask *model.TissueMask, now time.Time) bool {
		if mask.Status == vobj.TissueMaskStatusRejected && equalStrings(mask.RejectReason, cmd.Reason) {
			return false
		}
		mask.Status = vobj.TissueMaskStatusRejected
		mask.RejectedBy, mask.RejectedAt, mask.RejectReason = &cmd.UserID, &now, cmd.Reason
		mask.ApprovedBy, mask.ApprovedAt = nil, nil
		return true
	})
}

// review applies a status change to an existing mask; change reports whether
// anything changed.
func (uc *TissueMaskUseCase) review(ctx context.Context, cmd command.ReviewTissueMaskCommand, change func(*model.TissueMask, time.Time) bool) (*model.TissueMask, error) {
	if cmd.ImageID == "" || cmd.UserID == "" {
		return nil, errors.NewValidationError("image_id and user_id are required", nil)
	}

	var result *model.TissueMask
	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		existing, err := uc.readMask(txCtx, cmd.ImageID)
		if err != nil {
			return err
		}
		if existing == nil || existing.Deleted {
			return errors.NewNotFoundError("tissue mask not found")
		}
		if err := checkRevision(existing, cmd.ExpectedRevision); err != nil {
			return err
		}

		mask := *existing
		if !change(&mask, uc.now()) {
			result = existing
			return nil
		}
		if err := uc.write(txCtx, existing, &mask); err != nil {
			return err
		}
		result = &mask
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ApplyWorkerResult stores the worker's default mask unless a person has
// already edited, approved or rejected the image's mask. It reports whether it
// wrote.
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

// checkRevision refuses writes based on a revision other than the stored one.
// A nil expectation skips the check.
func checkRevision(existing *model.TissueMask, expected *int) error {
	if expected == nil {
		return nil
	}
	current := 0
	if existing != nil && !existing.Deleted {
		current = existing.Revision
	}
	if current == *expected {
		return nil
	}
	details := map[string]interface{}{
		"expected_revision": *expected,
		"current_revision":  current,
	}
	if existing != nil {
		details["status"] = existing.Status.String()
		details["updated_at"] = existing.UpdatedAt
	}
	return errors.NewConflictError("tissue mask was changed by someone else; reload it", details)
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

// write creates the document or replaces its mask fields (keeping created_at)
// and advances the revision.
func (uc *TissueMaskUseCase) write(ctx context.Context, existing, mask *model.TissueMask) error {
	mask.Revision = 1
	if existing != nil {
		mask.Revision = existing.Revision + 1
	}

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
	mask.Deleted = false
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

func equalStrings(a, b *string) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return *a == *b
}

func isNotFound(err error) bool {
	var appErr *errors.Err
	return stderrors.As(err, &appErr) && appErr.Type == errors.ErrorTypeNotFound
}
