package validator

import (
	"context"

	"github.com/histopathai/main-service/internal/application/usecase/helper"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type WorkspaceValidator struct {
	repo port.WorkspaceRepository
	uow  port.UnitOfWorkFactory
}

func NewWorkspaceValidator(repo port.WorkspaceRepository, uow port.UnitOfWorkFactory) *WorkspaceValidator {
	return &WorkspaceValidator{repo: repo, uow: uow}
}

func (v *WorkspaceValidator) ValidateCreate(ctx context.Context, ws *model.Workspace) error {
	if isUnique, err := helper.CheckNameUniqueInCollection(ctx, v.repo, ws.Name); err != nil {
		return errors.NewInternalError("failed to check name uniqueness", err)
	} else if !isUnique {
		return errors.NewConflictError("workspace name already exists", map[string]interface{}{
			"name": ws.Name,
		})
	}
	return nil

}

func (v *WorkspaceValidator) ValidateUpdate(ctx context.Context, id string, updates map[string]interface{}) error {
	if name, ok := updates[fields.EntityName.DomainName()]; ok {
		if isUnique, err := helper.CheckNameUniqueInCollection(ctx, v.repo, name.(string), id); err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		} else if !isUnique {
			return errors.NewConflictError("workspace name already exists", map[string]interface{}{
				"name": name,
			})
		}
	}

	return nil
}

func (v *WorkspaceValidator) ValidateAnnotationTypeRemoval(ctx context.Context, workspaceId string, currentTypes, newTypes []string) error {
	removedTypes := v.getRemovedItems(currentTypes, newTypes)
	if len(removedTypes) == 0 {
		return nil
	}

	//Check if workspace has patients
	hasPatients, err := v.hasAnyPatients(ctx, workspaceId)
	if err != nil {
		return err
	}
	if !hasPatients {
		return nil // safe to remove
	}

	//Check if workspace has annotations
	hasAnnotations, err := v.hasAnyAnnotations(ctx, workspaceId)
	if err != nil {
		return err
	}
	if !hasAnnotations {
		return nil // safe to remove
	}

	if err := v.validateRemovedTypesNotInUse(ctx, removedTypes); err != nil {
		return err
	}

	return nil
}

// Helper Methods

func (v *WorkspaceValidator) getRemovedItems(currentTypes, newTypes []string) []string {
	removed := []string{}
	for _, currentType := range currentTypes {
		if !helper.Contains(newTypes, currentType) {
			removed = append(removed, currentType)
		}
	}
	return removed
}

func (v *WorkspaceValidator) hasAnyPatients(ctx context.Context, ws_id string) (bool, error) {
	patientRepo := v.uow.GetPatientRepo()

	builder := query.NewBuilder()
	builder.Where(fields.EntityParentID.DomainName(), query.OpEqual, ws_id)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
	count, err := patientRepo.Count(ctx, builder.Build())
	if err != nil {
		return false, errors.NewInternalError("failed to check name uniqueness", err)
	}

	return count == 0, nil

}

func (v *WorkspaceValidator) hasAnyAnnotations(ctx context.Context, ws_id string) (bool, error) {
	annotationRepo := v.uow.GetAnnotationRepo()

	builder := query.NewBuilder()
	builder.Where(fields.EntityParentID.DomainName(), query.OpEqual, ws_id)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
	count, err := annotationRepo.Count(ctx, builder.Build())
	if err != nil {
		return false, errors.NewInternalError("failed to check name uniqueness", err)
	}

	return count == 0, nil

}

func (v *WorkspaceValidator) validateRemovedTypesNotInUse(ctx context.Context, removedTypes []string) error {
	annotationRepo := v.uow.GetAnnotationRepo()

	builder := query.NewBuilder()
	builder.Where(fields.AnnotationTypeID.DomainName(), query.OpIn, removedTypes)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
	count, err := annotationRepo.Count(ctx, builder.Build())
	if err != nil {
		return errors.NewInternalError("failed to check annotation type in use", err)
	}

	if count > 0 {
		return errors.NewConflictError("annotation type in use", nil)
	}

	return nil

}
