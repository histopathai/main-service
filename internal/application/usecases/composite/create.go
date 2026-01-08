package composite

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type CreateUseCase[T port.Entity] struct {
	uowFactory port.UnitOfWorkFactory
}

func NewCreateUseCase[T port.Entity](uowFactory port.UnitOfWorkFactory) *CreateUseCase[T] {
	return &CreateUseCase[T]{
		uowFactory: uowFactory,
	}
}

func (uc *CreateUseCase[T]) Execute(ctx context.Context, entity T) (T, error) {
	var zero T
	switch entity.GetEntityType() {
	case vobj.EntityTypePatient:
		return uc.createPatient(ctx, entity)
	case vobj.EntityTypeImage:
		return uc.createImage(ctx, entity)
	case vobj.EntityTypeWorkspace:
		return uc.createWorkspace(ctx, entity)
	case vobj.EntityTypeAnnotationType:
		return uc.createAnnotationType(ctx, entity)
	case vobj.EntityTypeAnnotation:
		return uc.createAnnotation(ctx, entity)
	default:
		return zero, errors.NewValidationError("unsupported entity type for creation", nil)
	}
}

func (uc *CreateUseCase[T]) createPatient(ctx context.Context, entity T) (T, error) {
	var zero T
	patient, ok := any(entity).(model.Patient)
	if !ok {
		return zero, errors.NewValidationError("entity is not a Patient", nil)
	}

	var createdEntity T

	err := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		// Validate parent workspace
		parentID := entity.GetParent().GetID()
		parentEntity, err := uc.validateParent(txCtx, repos, parentID, vobj.EntityTypeWorkspace)
		if err != nil {
			return err
		}

		// Create patient
		created, err := repos.PatientRepo.Create(txCtx, &patient)
		if err != nil {
			return err
		}

		// Update parent child count
		if err := uc.incrementChildCount(txCtx, repos, parentID, vobj.EntityTypeWorkspace, parentEntity.GetChildCount()); err != nil {
			return err
		}

		// Convert back to generic type
		createdEntity, ok = any(*created).(T)
		if !ok {
			return errors.NewInternalError("failed to convert created patient to entity type", nil)
		}

		return nil
	})

	if err != nil {
		return zero, err
	}
	return createdEntity, nil
}

func (uc *CreateUseCase[T]) createWorkspace(ctx context.Context, entity T) (T, error) {
	var zero T
	workspace, ok := any(entity).(model.Workspace)
	if !ok {
		return zero, errors.NewValidationError("entity is not a Workspace", nil)
	}

	var createdEntity T

	err := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		// Check name uniqueness
		filters := []query.Filter{
			{
				Field:    constants.NameField,
				Operator: query.OpEqual,
				Value:    workspace.Name,
			},
		}
		count, err := repos.WorkspaceRepo.Count(txCtx, filters)
		if err != nil {
			return err
		}
		if count > 0 {
			details := map[string]any{
				"where": "CreateUseCase.createWorkspace",
				"type":  "uniqueness violation",
				"name":  workspace.Name,
			}
			return errors.NewConflictError("workspace with the given name already exists", details)
		}

		// Create workspace
		created, err := repos.WorkspaceRepo.Create(txCtx, &workspace)
		if err != nil {
			return err
		}

		// Convert back to generic type
		createdEntity, ok = any(*created).(T)
		if !ok {
			return errors.NewInternalError("failed to convert created workspace to entity type", nil)
		}

		return nil
	})

	if err != nil {
		return zero, err
	}
	return createdEntity, nil
}

func (uc *CreateUseCase[T]) createImage(ctx context.Context, entity T) (T, error) {
	var zero T
	image, ok := any(entity).(model.Image)
	if !ok {
		return zero, errors.NewValidationError("entity is not an Image", nil)
	}

	var createdEntity T

	err := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		// Validate parent patient
		parentID := entity.GetParent().GetID()
		parentEntity, err := uc.validateParent(txCtx, repos, parentID, vobj.EntityTypePatient)
		if err != nil {
			return err
		}

		// Create image
		created, err := repos.ImageRepo.Create(txCtx, &image)
		if err != nil {
			return err
		}

		// Update parent child count
		if err := uc.incrementChildCount(txCtx, repos, parentID, vobj.EntityTypePatient, parentEntity.GetChildCount()); err != nil {
			return err
		}

		// Convert back to generic type
		createdEntity, ok = any(*created).(T)
		if !ok {
			return errors.NewInternalError("failed to convert created image to entity type", nil)
		}

		return nil
	})

	if err != nil {
		return zero, err
	}
	return createdEntity, nil
}

func (uc *CreateUseCase[T]) createAnnotation(ctx context.Context, entity T) (T, error) {
	var zero T
	annotation, ok := any(entity).(model.Annotation)
	if !ok {
		return zero, errors.NewValidationError("entity is not an Annotation", nil)
	}

	var createdEntity T

	err := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		// Validate parent image
		parentID := entity.GetParent().GetID()
		parentEntity, err := uc.validateParent(txCtx, repos, parentID, vobj.EntityTypeImage)
		if err != nil {
			return err
		}

		// Create annotation
		created, err := repos.AnnotationRepo.Create(txCtx, &annotation)
		if err != nil {
			return err
		}

		// Update parent child count
		if err := uc.incrementChildCount(txCtx, repos, parentID, vobj.EntityTypeImage, parentEntity.GetChildCount()); err != nil {
			return err
		}

		// Convert back to generic type
		createdEntity, ok = any(*created).(T)
		if !ok {
			return errors.NewInternalError("failed to convert created annotation to entity type", nil)
		}

		return nil
	})

	if err != nil {
		return zero, err
	}
	return createdEntity, nil
}

func (uc *CreateUseCase[T]) createAnnotationType(ctx context.Context, entity T) (T, error) {
	var zero T
	annotationType, ok := any(entity).(model.AnnotationType)
	if !ok {
		return zero, errors.NewValidationError("entity is not an AnnotationType", nil)
	}

	var createdEntity T

	err := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		// Create annotation type
		created, err := repos.AnnotationTypeRepo.Create(txCtx, &annotationType)
		if err != nil {
			return err
		}

		// Convert back to generic type
		createdEntity, ok = any(*created).(T)
		if !ok {
			return errors.NewInternalError("failed to convert created annotation type to entity type", nil)
		}

		return nil
	})

	if err != nil {
		return zero, err
	}
	return createdEntity, nil
}

// validateParent validates that a parent entity exists and is not deleted
func (uc *CreateUseCase[T]) validateParent(ctx context.Context, repos *port.Repositories, parentID string, parentType vobj.EntityType) (port.Entity, error) {
	var parent port.Entity
	var err error

	switch parentType {
	case vobj.EntityTypeWorkspace:
		parent, err = repos.WorkspaceRepo.Read(ctx, parentID)
	case vobj.EntityTypePatient:
		parent, err = repos.PatientRepo.Read(ctx, parentID)
	case vobj.EntityTypeImage:
		parent, err = repos.ImageRepo.Read(ctx, parentID)
	default:
		return nil, errors.NewValidationError("unsupported parent type", nil)
	}

	if err != nil {
		return nil, err
	}

	if parent == nil {
		details := map[string]any{
			"type":      "parent not found",
			"parent_id": parentID,
		}
		return nil, errors.NewValidationError("parent entity not found", details)
	}

	if parent.IsDeleted() {
		details := map[string]any{
			"type":      "parent deleted",
			"parent_id": parentID,
		}
		return nil, errors.NewValidationError("cannot add to deleted parent", details)
	}

	return parent, nil
}

// incrementChildCount updates the child count of a parent entity
func (uc *CreateUseCase[T]) incrementChildCount(ctx context.Context, repos *port.Repositories, parentID string, parentType vobj.EntityType, currentCount int64) error {
	updateData := map[string]any{
		constants.ChildCountField: currentCount + 1,
	}

	switch parentType {
	case vobj.EntityTypeWorkspace:
		return repos.WorkspaceRepo.Update(ctx, parentID, updateData)
	case vobj.EntityTypePatient:
		return repos.PatientRepo.Update(ctx, parentID, updateData)
	case vobj.EntityTypeImage:
		return repos.ImageRepo.Update(ctx, parentID, updateData)
	default:
		return errors.NewValidationError("unsupported parent type for child count update", nil)
	}
}
