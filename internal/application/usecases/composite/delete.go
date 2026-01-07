package composite

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type DeleteUseCase struct {
	uowFactory repository.UnitOfWorkFactory
}

func NewDeleteUseCase(uowFactory repository.UnitOfWorkFactory) *DeleteUseCase {
	return &DeleteUseCase{uowFactory: uowFactory}
}

type EntityIDCollection struct {
	WorkspaceIDs      []string
	PatientIDs        []string
	ImageIDs          []string
	AnnotationIDs     []string
	AnnotationTypeIDs []string
}

func (uc *DeleteUseCase) Execute(ctx context.Context, id string, entityType vobj.EntityType) error {
	repos, err := uc.uowFactory.WithoutTx(ctx)
	if err != nil {
		return err
	}

	switch entityType {
	case vobj.EntityTypeWorkspace:
		return uc.deleteWorkspace(ctx, id, repos)
	case vobj.EntityTypePatient:
		return uc.deletePatient(ctx, id, repos)
	case vobj.EntityTypeImage:
		return uc.deleteImage(ctx, id, repos)
	case vobj.EntityTypeAnnotation:
		return uc.deleteAnnotation(ctx, id, repos)
	case vobj.EntityTypeAnnotationType:
		return uc.deleteAnnotationType(ctx, id, repos)
	default:
		return errors.NewValidationError("unsupported entity type for delete", nil)
	}
}

func (uc *DeleteUseCase) deleteWorkspace(
	ctx context.Context,
	id string,
	repos *repository.Repositories,
) error {
	workspace, err := repos.WorkspaceRepo.Read(ctx, id)
	if err != nil {
		return err
	}

	if workspace == nil {
		return errors.NewNotFoundError("workspace not found")
	}

	if !workspace.IsDeleted() {
		return errors.NewValidationError("workspace is not marked as deleted", map[string]any{
			"where":        "DeleteUseCase.deleteWorkspace",
			"workspace_id": id,
		})
	}

	if workspace.GetChildCount() > 0 {
		allPatients, err := repos.PatientRepo.GetChildren(ctx, id, true)
		if err != nil {
			return fmt.Errorf("failed to get patients: %w", err)
		}

		allImageIDs := make([]string, 0)
		allAnnotationIDs := make([]string, 0)

		for _, patient := range allPatients {
			patientImages, err := repos.ImageRepo.GetChildren(ctx, patient.GetID(), true)
			if err != nil {
				return fmt.Errorf("failed to get images for patient %s: %w", patient.GetID(), err)
			}

			for _, image := range patientImages {
				allImageIDs = append(allImageIDs, image.GetID())

				if image.GetChildCount() > 0 {
					imageAnnotations, err := repos.AnnotationRepo.GetChildren(ctx, image.GetID(), true)
					if err != nil {
						return fmt.Errorf("failed to get annotations for image %s: %w", image.GetID(), err)
					}

					for _, annotation := range imageAnnotations {
						allAnnotationIDs = append(allAnnotationIDs, annotation.GetID())
					}
				}
			}
		}

		if len(allAnnotationIDs) > 0 {
			if err := repos.AnnotationRepo.DeleteMany(ctx, allAnnotationIDs); err != nil {
				return fmt.Errorf("failed to delete annotations: %w", err)
			}
		}

		if len(allImageIDs) > 0 {
			if err := repos.ImageRepo.DeleteMany(ctx, allImageIDs); err != nil {
				return fmt.Errorf("failed to delete images: %w", err)
			}
		}

		if len(allPatients) > 0 {
			patientIDs := make([]string, len(allPatients))
			for i, patient := range allPatients {
				patientIDs[i] = patient.GetID()
			}

			if err := repos.PatientRepo.DeleteMany(ctx, patientIDs); err != nil {
				return fmt.Errorf("failed to delete patients: %w", err)
			}
		}
	}

	annotationTypeID := workspace.GetParent().GetID()
	if annotationTypeID != "" {
		annotationType, err := repos.AnnotationTypeRepo.Read(ctx, annotationTypeID)
		if err != nil {
			return fmt.Errorf("failed to read annotation type: %w", err)
		}

		if annotationType != nil {
			childCount := annotationType.GetChildCount()

			if childCount == 1 {
				if err := repos.AnnotationTypeRepo.Delete(ctx, annotationTypeID); err != nil {
					return fmt.Errorf("failed to delete annotation type: %w", err)
				}
			} else if childCount > 1 {
				if err := repos.AnnotationTypeRepo.Update(ctx, annotationTypeID, map[string]any{
					constants.ChildCountField: childCount - 1,
				}); err != nil {
					return fmt.Errorf("failed to update annotation type child count: %w", err)
				}
			}
		}
	}

	return repos.WorkspaceRepo.Delete(ctx, id)
}

func (uc *DeleteUseCase) deletePatient(
	ctx context.Context,
	id string,
	repos *repository.Repositories,
) error {
	patient, err := repos.PatientRepo.Read(ctx, id)
	if err != nil {
		return err
	}

	if patient == nil {
		return errors.NewNotFoundError("patient not found")
	}

	if patient.GetChildCount() == 0 {
		return repos.PatientRepo.Delete(ctx, id)
	}

	allImages, err := repos.ImageRepo.GetChildren(ctx, id, true)
	if err != nil {
		return fmt.Errorf("failed to get images: %w", err)
	}

	allAnnotationIDs := make([]string, 0)

	for _, image := range allImages {
		if image.GetChildCount() > 0 {
			imageAnnotations, err := repos.AnnotationRepo.GetChildren(ctx, image.GetID(), true)
			if err != nil {
				return fmt.Errorf("failed to get annotations for image %s: %w", image.GetID(), err)
			}

			for _, annotation := range imageAnnotations {
				allAnnotationIDs = append(allAnnotationIDs, annotation.GetID())
			}
		}
	}

	if len(allAnnotationIDs) > 0 {
		if err := repos.AnnotationRepo.DeleteMany(ctx, allAnnotationIDs); err != nil {
			return fmt.Errorf("failed to delete annotations: %w", err)
		}
	}

	if len(allImages) > 0 {
		imageIDs := make([]string, len(allImages))
		for i, image := range allImages {
			imageIDs[i] = image.GetID()
		}

		if err := repos.ImageRepo.DeleteMany(ctx, imageIDs); err != nil {
			return fmt.Errorf("failed to delete images: %w", err)
		}
	}

	return repos.PatientRepo.Delete(ctx, id)
}

func (uc *DeleteUseCase) deleteImage(
	ctx context.Context,
	id string,
	repos *repository.Repositories,
) error {
	image, err := repos.ImageRepo.Read(ctx, id)
	if err != nil {
		return err
	}

	if image == nil {
		return errors.NewNotFoundError("image not found")
	}

	if image.GetChildCount() == 0 {
		return repos.ImageRepo.Delete(ctx, id)
	}

	allAnnotations, err := repos.AnnotationRepo.GetChildren(ctx, id, true)
	if err != nil {
		return fmt.Errorf("failed to get annotations: %w", err)
	}
	if len(allAnnotations) > 0 {
		annotationIDs := make([]string, len(allAnnotations))
		for i, annotation := range allAnnotations {
			annotationIDs[i] = annotation.GetID()
		}

		if err := repos.AnnotationRepo.DeleteMany(ctx, annotationIDs); err != nil {
			return fmt.Errorf("failed to delete annotations: %w", err)
		}
	}

	return repos.ImageRepo.Delete(ctx, id)
}

func (uc *DeleteUseCase) deleteAnnotation(
	ctx context.Context,
	id string,
	repos *repository.Repositories,
) error {
	annotation, err := repos.AnnotationRepo.Read(ctx, id)
	if err != nil {
		return err
	}

	if annotation == nil {
		return errors.NewNotFoundError("annotation not found")
	}

	return repos.AnnotationRepo.Delete(ctx, id)
}

func (uc *DeleteUseCase) deleteAnnotationType(
	ctx context.Context,
	id string,
	repos *repository.Repositories,
) error {
	annotationType, err := repos.AnnotationTypeRepo.Read(ctx, id)
	if err != nil {
		return err
	}

	if annotationType == nil {
		return errors.NewNotFoundError("annotation type not found")
	}

	if annotationType.GetChildCount() > 0 {
		return errors.NewValidationError(
			"cannot delete annotation type that is still in use by workspaces",
			map[string]any{
				"annotation_type_id": id,
				"workspace_count":    annotationType.GetChildCount(),
			},
		)
	}

	return repos.AnnotationTypeRepo.Delete(ctx, id)
}

func (uc *DeleteUseCase) ExecuteMany(
	ctx context.Context,
	ids []string,
	entityType vobj.EntityType,
) error {
	for _, id := range ids {
		if err := uc.Execute(ctx, id, entityType); err != nil {
			return fmt.Errorf("failed to delete entity %s: %w", id, err)
		}
	}
	return nil
}
