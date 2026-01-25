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

type PatientUseCase struct {
	repo port.Repository[*model.Patient]
	uow  port.UnitOfWorkFactory
}

func NewPatientUseCase(repo port.Repository[*model.Patient], uow port.UnitOfWorkFactory) *PatientUseCase {
	return &PatientUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *PatientUseCase) Create(ctx context.Context, entity *model.Patient) (*model.Patient, error) {
	var createdPatient *model.Patient

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		if err := CheckParentExists(txCtx, &entity.Parent, uc.uow); err != nil {
			return errors.NewValidationError("parent validation failed", map[string]interface{}{
				"parent_type": entity.GetParent().Type,
				"parent_id":   entity.GetParent().ID,
				"error":       err.Error(),
			})
		}

		parentWorkspace, err := uc.uow.GetWorkspaceRepo().Read(txCtx, entity.Parent.ID)
		if err != nil {
			return errors.NewInternalError("failed to read parent workspace", err)
		}

		if len(parentWorkspace.AnnotationTypes) == 0 {
			return errors.NewValidationError("parent workspace has no annotation types defined", map[string]interface{}{
				"parent_id": entity.Parent.ID,
			})
		}

		isUnique, err := CheckNameUniqueUnderParent(txCtx, uc.repo, entity.Name, entity.Parent.ID)
		if err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		}
		if !isUnique {
			return errors.NewConflictError("patient name already exists", map[string]interface{}{
				"name":      entity.Name,
				"parent_id": entity.Parent.ID,
			})
		}

		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create patient", err)
		}

		createdPatient = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdPatient, nil
}

func (uc *PatientUseCase) Update(ctx context.Context, patientID string, updates map[string]interface{}) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		if name, ok := updates[constants.NameField]; ok {
			currentPatient, err := uc.repo.Read(txCtx, patientID)
			if err != nil {
				return errors.NewInternalError("failed to read patient", err)
			}

			isUnique, err := CheckNameUniqueUnderParent(txCtx, uc.repo, name.(string), currentPatient.Parent.ID, patientID)
			if err != nil {
				return errors.NewInternalError("failed to check name uniqueness", err)
			}
			if !isUnique {
				return errors.NewConflictError("patient name already exists", map[string]interface{}{
					"name":      name,
					"parent_id": currentPatient.Parent.ID,
				})
			}
		}

		err := uc.repo.Update(txCtx, patientID, updates)
		if err != nil {
			return errors.NewInternalError("failed to update patient", err)
		}

		return nil
	})

	return err
}

func (uc *PatientUseCase) Transfer(ctx context.Context, patientID string, newParent vobj.ParentRef) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		currentPatient, err := uc.repo.Read(txCtx, patientID)
		if err != nil {
			return errors.NewInternalError("failed to read patient", err)
		}

		newParentWorkspace, err := uc.uow.GetWorkspaceRepo().Read(txCtx, newParent.ID)
		if err != nil {
			return errors.NewValidationError("new parent workspace does not exist", map[string]interface{}{
				"parent_id": newParent.ID,
				"error":     err.Error(),
			})
		}

		if len(newParentWorkspace.AnnotationTypes) == 0 {
			return errors.NewValidationError("new parent workspace has no annotation types defined", map[string]interface{}{
				"parent_id": newParent.ID,
			})
		}

		// Check Old parent's annotation types are subset of new parent's annotation types
		oldParentWorkspace, err := uc.uow.GetWorkspaceRepo().Read(txCtx, currentPatient.Parent.ID)
		if err != nil {
			return errors.NewInternalError("failed to read old parent workspace", err)
		}

		// NEW workspace annotation types set
		newAnnotationTypeSet := make(map[string]struct{})
		for _, at := range newParentWorkspace.AnnotationTypes {
			newAnnotationTypeSet[at] = struct{}{}
		}

		// Check each OLD annotation type exists in NEW workspace
		for _, oldType := range oldParentWorkspace.AnnotationTypes {
			if _, exists := newAnnotationTypeSet[oldType]; !exists {
				return errors.NewValidationError("new parent workspace does not contain all annotation types of old parent", map[string]interface{}{
					"old_parent_id": currentPatient.Parent.ID,
					"new_parent_id": newParent.ID,
					"missing_type":  oldType,
				})
			}
		}

		// Check name uniqueness in new parent
		isUnique, err := CheckNameUniqueUnderParent(txCtx, uc.repo, currentPatient.Name, newParent.ID)
		if err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		}
		if !isUnique {
			return errors.NewConflictError("patient with same name already exists in new parent", map[string]interface{}{
				"name":      currentPatient.Name,
				"parent_id": newParent.ID,
			})
		}

		// Transfer patient first (parent entity)
		err = uc.repo.Transfer(txCtx, patientID, newParent.ID)
		if err != nil {
			return errors.NewInternalError("failed to transfer patient", err)
		}

		// Get all image IDs under this patient
		imageIDs, err := uc.getImageIDsUnderPatient(txCtx, patientID, currentPatient.Parent.ID)
		if err != nil {
			return err
		}

		// Transfer images if any exist
		if len(imageIDs) > 0 {
			// Check Firestore transaction limit (500 operations)
			totalOps := 1 + len(imageIDs) // patient update + image updates

			// Get annotation count to calculate total operations
			annotationIDs, err := uc.getAnnotationIDsUnderImages(txCtx, imageIDs, currentPatient.Parent.ID)
			if err != nil {
				return err
			}
			totalOps += len(annotationIDs)

			if totalOps > maxOpsPerTx {
				return errors.NewValidationError("transfer operation exceeds transaction limit", map[string]interface{}{
					"image_count":      len(imageIDs),
					"annotation_count": len(annotationIDs),
					"total_operations": totalOps,
					"limit":            maxOpsPerTx,
					"message":          "Patient has too many images/annotations for atomic transfer. Please contact support.",
				})
			}

			// Transfer images
			imageRepo := uc.uow.GetImageRepo()
			err = imageRepo.UpdateMany(txCtx, imageIDs, map[string]interface{}{
				constants.WsIDField: newParent.ID,
			})
			if err != nil {
				return errors.NewInternalError("failed to transfer images to new workspace", err)
			}

			// Transfer annotations if any exist
			if len(annotationIDs) > 0 {
				annotationRepo := uc.uow.GetAnnotationRepo()
				err = annotationRepo.UpdateMany(txCtx, annotationIDs, map[string]interface{}{
					constants.WsIDField: newParent.ID,
				})
				if err != nil {
					return errors.NewInternalError("failed to transfer annotations to new workspace", err)
				}
			}
		}

		return nil
	})

	return err
}

func (uc *PatientUseCase) getImageIDsUnderPatient(ctx context.Context, patientID string, oldWsID string) ([]string, error) {
	imageRepo := uc.uow.GetImageRepo()

	builder := query.NewBuilder()
	builder.Where(constants.ParentIDField, query.OpEqual, patientID)
	builder.Where(constants.WsIDField, query.OpEqual, oldWsID)
	builder.Where("is_deleted", query.OpEqual, false)

	const limit = 1000
	offset := 0
	var allImageIDs []string

	for {
		builder.Paginate(limit, offset)
		spec := builder.Build()

		result, err := imageRepo.Find(ctx, spec)
		if err != nil {
			return nil, errors.NewInternalError("failed to fetch images", err)
		}

		for _, img := range result.Data {
			allImageIDs = append(allImageIDs, img.GetID())
		}

		if !result.HasMore {
			break
		}

		offset += limit
	}

	return allImageIDs, nil
}

func (uc *PatientUseCase) getAnnotationIDsUnderImages(ctx context.Context, imageIDs []string, oldWsID string) ([]string, error) {
	if len(imageIDs) == 0 {
		return []string{}, nil
	}

	annotationRepo := uc.uow.GetAnnotationRepo()

	builder := query.NewBuilder()
	builder.Where(constants.ParentIDField, query.OpIn, imageIDs)
	builder.Where(constants.WsIDField, query.OpEqual, oldWsID)
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
