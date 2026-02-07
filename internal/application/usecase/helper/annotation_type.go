// usecase/helper/annotation_type.go
package helper

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
)

// UpdateAnnotationNames updates all annotation names when annotation type name changes
func UpdateAnnotationNames(
	ctx context.Context,
	annotationRepo port.AnnotationRepository,
	oldName, newName string,
) error {
	builder := query.NewBuilder()
	builder.Where(fields.EntityName.DomainName(), query.OpEqual, oldName)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	const limit = 100
	offset := 0

	for {
		builder.Paginate(limit, offset)
		result, err := annotationRepo.Find(ctx, builder.Build())
		if err != nil {
			return fmt.Errorf("failed to fetch annotations: %w", err)
		}

		if len(result.Data) == 0 {
			break
		}

		// Update batch
		for _, annotation := range result.Data {
			err := annotationRepo.Update(ctx, annotation.GetID(), map[string]interface{}{
				fields.EntityName.DomainName(): newName,
			})
			if err != nil {
				return fmt.Errorf("failed to update annotation name: %w", err)
			}
		}

		if !result.HasMore {
			break
		}

		offset += limit
	}

	return nil
}

// CountAnnotationsByType counts annotations using a specific annotation type
func CountAnnotationsByType(
	ctx context.Context,
	annotationRepo port.AnnotationRepository,
	annotationTypeName string,
) (int64, error) {
	builder := query.NewBuilder()
	builder.Where(fields.EntityName.DomainName(), query.OpEqual, annotationTypeName)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)

	count, err := annotationRepo.Count(ctx, builder.Build())
	if err != nil {
		return 0, fmt.Errorf("failed to count annotations: %w", err)
	}

	return count, nil
}

// FindAnnotationTypeByNameAndWsID finds annotation type by name and workspace ID
func FindAnnotationTypeByNameAndWsID(
	ctx context.Context,
	annotationTypeRepo port.AnnotationTypeRepository,
	name string,
	wsID string,
) (string, error) {
	builder := query.NewBuilder()
	builder.Where(fields.EntityName.DomainName(), query.OpEqual, name)
	builder.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
	builder.Paginate(10, 0)

	result, err := annotationTypeRepo.Find(ctx, builder.Build())
	if err != nil {
		return "", fmt.Errorf("failed to find annotation type: %w", err)
	}

	if len(result.Data) == 0 {
		return "", fmt.Errorf("annotation type not found with name: %s", name)
	}

	// Find the one that matches workspace or is global
	for _, at := range result.Data {
		// Check if it's global or belongs to the workspace
		if at.IsGlobal {
			return at.GetID(), nil
		}
		// Check if it belongs to the workspace by checking parent
		if at.Parent.ID == wsID {
			return at.GetID(), nil
		}
	}

	// If no match found, return the first one (fallback)
	return result.Data[0].GetID(), nil
}
