package helper

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

func CheckNameUniqueUnderParent[T port.Entity](ctx context.Context, repo port.Repository[T], name string, parentID string, excludeID ...string) (bool, error) {
	builder := query.NewBuilder()
	builder.Where(fields.EntityName.DomainName(), query.OpEqual, name)
	builder.Where(fields.EntityParentID.DomainName(), query.OpEqual, parentID)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	if len(excludeID) > 0 && excludeID[0] != "" {
		builder.Where(fields.EntityID.DomainName(), query.OpNotEqual, excludeID[0])
	}

	count, err := repo.Count(ctx, builder.Build())
	if err != nil {
		return false, fmt.Errorf("failed to check name uniqueness: %w", err)
	}

	return count == 0, nil
}

func CheckNameUniqueInCollection[T port.Entity](ctx context.Context, repo port.Repository[T], name string, excludeID ...string) (bool, error) {
	builder := query.NewBuilder()
	builder.Where(fields.EntityName.DomainName(), query.OpEqual, name)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	if len(excludeID) > 0 && excludeID[0] != "" {
		builder.Where(fields.EntityID.DomainName(), query.OpNotEqual, excludeID[0])
	}

	count, err := repo.Count(ctx, builder.Build())
	if err != nil {
		return false, fmt.Errorf("failed to check name uniqueness in collection: %w", err)
	}

	return count == 0, nil
}

func CheckParentExists(ctx context.Context, parent *vobj.ParentRef, uow port.UnitOfWorkFactory) error {
	if parent == nil || parent.ID == "" {
		return nil
	}
	var err error

	switch parent.Type {
	case vobj.ParentTypeWorkspace:
		return nil
	case vobj.ParentTypePatient:
		repo := uow.GetPatientRepo()
		_, err = repo.Read(ctx, parent.ID)
	case vobj.ParentTypeImage:
		repo := uow.GetImageRepo()
		_, err = repo.Read(ctx, parent.ID)
	case vobj.ParentTypeAnnotation:
		repo := uow.GetAnnotationRepo()
		_, err = repo.Read(ctx, parent.ID)
	case vobj.ParentTypeAnnotationType:
		repo := uow.GetAnnotationTypeRepo()
		_, err = repo.Read(ctx, parent.ID)
	case vobj.ParentTypeContent:
		return nil
	default:
		return fmt.Errorf("unknown parent type: %s", parent.Type)
	}

	if err != nil {
		return fmt.Errorf("parent %s with ID %s not found: %w", parent.Type, parent.ID, err)
	}

	return nil
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func ValidateImageStatusTransition(currentStatus, newStatus vobj.ImageStatus) error {
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
