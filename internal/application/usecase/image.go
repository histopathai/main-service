package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type ImageUseCase struct {
	repo port.Repository[*model.Image]
	uow  port.UnitOfWorkFactory
}

func NewImageUseCase(repo port.Repository[*model.Image], uow port.UnitOfWorkFactory) *ImageUseCase {
	return &ImageUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *ImageUseCase) Create(ctx context.Context, entity *model.Image) (*model.Image, error) {
	var createdImage *model.Image

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// Validate parent exists (must be a Patient)
		if err := CheckParentExists(txCtx, &entity.Parent, uc.uow); err != nil {
			return errors.NewValidationError("parent validation failed", map[string]interface{}{
				"parent_type": entity.GetParent().Type,
				"parent_id":   entity.GetParent().ID,
				"error":       err.Error(),
			})
		}

		// Validate parent type is Patient
		if entity.Parent.Type != vobj.ParentTypePatient {
			return errors.NewValidationError("image parent must be a patient", map[string]interface{}{
				"parent_type": entity.Parent.Type,
				"expected":    vobj.ParentTypePatient,
			})
		}

		// Get workspace ID from parent patient if not set
		if entity.WsID == "" {
			parentRepo := uc.uow.GetPatientRepo()
			parentPatient, err := parentRepo.Read(txCtx, entity.Parent.ID)
			if err != nil {
				return errors.NewInternalError("failed to read parent patient", err)
			}
			entity.WsID = parentPatient.Parent.ID
		}

		// Ensure processing status is set
		if entity.Processing.Status == "" {
			// Default to PENDING for web uploads
			entity.Processing.Status = vobj.StatusPending
		}

		// Create image
		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create image", err)
		}

		createdImage = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdImage, nil
}

func (uc *ImageUseCase) Transfer(ctx context.Context, imageID string, newParent vobj.ParentRef) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// Read current image
		currentImage, err := uc.repo.Read(txCtx, imageID)
		if err != nil {
			return errors.NewInternalError("failed to read image", err)
		}

		// Validate new parent exists and is a Patient
		if err := CheckParentExists(txCtx, &newParent, uc.uow); err != nil {
			return errors.NewValidationError("new parent validation failed", map[string]interface{}{
				"parent_type": newParent.Type,
				"parent_id":   newParent.ID,
				"error":       err.Error(),
			})
		}

		if newParent.Type != vobj.ParentTypePatient {
			return errors.NewValidationError("image parent must be a patient", map[string]interface{}{
				"parent_type": newParent.Type,
				"expected":    vobj.ParentTypePatient,
			})
		}

		// Get new parent patient to check workspace
		newParentPatient, err := uc.uow.GetPatientRepo().Read(txCtx, newParent.ID)
		if err != nil {
			return errors.NewInternalError("failed to read new parent patient", err)
		}

		// Transfer image (updates parent_id)
		err = uc.repo.Transfer(txCtx, imageID, newParent.ID)
		if err != nil {
			return errors.NewInternalError("failed to transfer image", err)
		}

		// Get annotation IDs under this image
		annotationIDs, err := uc.getAnnotationIDsUnderImage(txCtx, imageID, currentImage.WsID)
		if err != nil {
			return err
		}

		// Transfer annotations to new workspace if any exist
		if len(annotationIDs) > 0 {
			// Check transaction limit
			totalOps := 1 + len(annotationIDs) // image transfer + annotation transfers

			if totalOps > maxOpsPerTx {
				return errors.NewValidationError("transfer operation exceeds transaction limit", map[string]interface{}{
					"annotation_count": len(annotationIDs),
					"total_operations": totalOps,
					"limit":            maxOpsPerTx,
					"message":          "Image has too many annotations for atomic transfer. Please contact support.",
				})
			}

			annotationRepo := uc.uow.GetAnnotationRepo()
			err = annotationRepo.UpdateMany(txCtx, annotationIDs, map[string]interface{}{
				constants.WsIDField: newParentPatient.Parent.ID,
			})
			if err != nil {
				return errors.NewInternalError("failed to transfer annotations to new workspace", err)
			}
		}

		return nil
	})

	return err
}

func (uc *ImageUseCase) UpdateStatus(ctx context.Context, imageID string, newStatus vobj.ImageStatus) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// Read current image
		currentImage, err := uc.repo.Read(txCtx, imageID)
		if err != nil {
			return errors.NewInternalError("failed to read image", err)
		}

		// Validate status transition
		if err := uc.validateStatusTransition(currentImage.Processing.Status, newStatus); err != nil {
			return err
		}

		// Update status
		updates := map[string]interface{}{
			constants.ImageProcessingStatusField: newStatus.String(),
		}
		err = uc.repo.Update(txCtx, imageID, updates)
		if err != nil {
			return errors.NewInternalError("failed to update image status", err)
		}

		return nil
	})

	return err
}

// validateStatusTransition validates if status transition is allowed
func (uc *ImageUseCase) validateStatusTransition(currentStatus, newStatus vobj.ImageStatus) error {
	// Define allowed transitions
	allowedTransitions := map[vobj.ImageStatus][]vobj.ImageStatus{
		vobj.StatusPending: {
			vobj.StatusProcessing,
			vobj.StatusDeleting,
		},
		vobj.StatusProcessing: {
			vobj.StatusProcessed,
			vobj.StatusFailed,
			vobj.StatusDeleting,
		},
		vobj.StatusProcessed: {
			vobj.StatusDeleting,
			vobj.StatusProcessing, // Re-processing allowed (e.g., V1 -> V2 migration)
		},
		vobj.StatusFailed: {
			vobj.StatusProcessing, // Retry
			vobj.StatusDeleting,
		},
		vobj.StatusDeleting: {
			// No transitions allowed from DELETING
		},
	}

	allowedStates, exists := allowedTransitions[currentStatus]
	if !exists {
		return errors.NewValidationError("invalid current status", map[string]interface{}{
			"current_status": currentStatus,
		})
	}

	for _, allowed := range allowedStates {
		if newStatus == allowed {
			return nil
		}
	}

	return errors.NewValidationError("invalid status transition", map[string]interface{}{
		"current_status": currentStatus,
		"new_status":     newStatus,
		"allowed":        allowedStates,
	})
}

func (uc *ImageUseCase) getAnnotationIDsUnderImage(ctx context.Context, imageID string, wsID string) ([]string, error) {
	annotationRepo := uc.uow.GetAnnotationRepo()

	builder := query.NewBuilder()
	builder.Where(constants.ParentIDField, query.OpEqual, imageID)
	builder.Where(constants.WsIDField, query.OpEqual, wsID)
	builder.Where("is_deleted", query.OpEqual, false)

	const limit = 1000
	offset := 0
	var allAnnotationIDs []string

	for {
		builder.Paginate(limit, offset)
		spec := builder.Build()

		result, err := annotationRepo.Find(ctx, spec)
		if err != nil {
			return nil, errors.NewInternalError("failed to fetch annotations", err)
		}

		for _, ann := range result.Data {
			allAnnotationIDs = append(allAnnotationIDs, ann.GetID())
		}

		if !result.HasMore {
			break
		}

		offset += limit
	}

	return allAnnotationIDs, nil
}
