package usecase

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
)

const maxOpsPerTx = 1000

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

func CheckNameUniqueUnderParent[T port.Entity](ctx context.Context, repo port.Repository[T], name string, parentID string, excludeID ...string) (bool, error) {
	filters := []query.Filter{
		{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    name,
		},
		{
			Field:    constants.ParentIDField,
			Operator: query.OpEqual,
			Value:    parentID,
		},
		{
			Field:    constants.DeletedField,
			Operator: query.OpEqual,
			Value:    false,
		},
	}

	if len(excludeID) > 0 && excludeID[0] != "" {
		filters = append(filters, query.Filter{
			Field:    constants.IDField,
			Operator: query.OpNotEqual,
			Value:    excludeID[0],
		})
	}

	count, err := repo.Count(ctx, filters)
	if err != nil {
		return false, fmt.Errorf("failed to check name uniqueness: %w", err)
	}

	return count == 0, nil
}

func CheckNameUniqueInCollection[T port.Entity](ctx context.Context, repo port.Repository[T], name string, excludeID ...string) (bool, error) {
	filters := []query.Filter{
		{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    name,
		},
		{
			Field:    constants.DeletedField,
			Operator: query.OpEqual,
			Value:    false,
		},
	}

	if len(excludeID) > 0 && excludeID[0] != "" {
		filters = append(filters, query.Filter{
			Field:    constants.IDField,
			Operator: query.OpNotEqual,
			Value:    excludeID[0],
		})
	}

	count, err := repo.Count(ctx, filters)
	if err != nil {
		return false, fmt.Errorf("failed to check name uniqueness in collection: %w", err)
	}

	return count == 0, nil
}
