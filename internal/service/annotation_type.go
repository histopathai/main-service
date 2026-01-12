package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	errors "github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeService struct {
	annotationTypeRepo port.AnnotationTypeRepository
	uow                port.UnitOfWorkFactory
}

func NewAnnotationTypeService(
	annotationTypeRepo port.AnnotationTypeRepository,
	uow port.UnitOfWorkFactory,
) *AnnotationTypeService {
	return &AnnotationTypeService{
		annotationTypeRepo: annotationTypeRepo,
		uow:                uow,
	}
}

func (ats *AnnotationTypeService) CreateNewAnnotationType(ctx context.Context, input *port.CreateAnnotationTypeInput) (*model.AnnotationType, error) {

	existing, err := ats.annotationTypeRepo.FindByName(ctx, input.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		details := map[string]interface{}{"name": "An annotation type with the same name already exists."}
		return nil, errors.NewConflictError("annotation type name already exists", details)
	}

	entity, err := vobj.NewEntity(vobj.EntityTypeAnnotationType, &input.Name, input.CreatorID, nil)
	if err != nil {
		return nil, err
	}

	atModel := &model.AnnotationType{
		Entity:   *entity,
		Type:     input.Type,
		Global:   input.Global,
		Required: input.Required,
		Options:  input.Options,
		Min:      input.Min,
		Max:      input.Max,
		Color:    input.Color,
	}

	created, err := ats.annotationTypeRepo.Create(ctx, atModel)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (ats *AnnotationTypeService) GetAnnotationTypeByID(ctx context.Context, id string) (*model.AnnotationType, error) {
	return ats.annotationTypeRepo.Read(ctx, id)
}

func (ats *AnnotationTypeService) ListAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[*model.AnnotationType], error) {
	return ats.annotationTypeRepo.FindByFilters(ctx, []sharedQuery.Filter{}, paginationOpts)
}

func (ats *AnnotationTypeService) UpdateAnnotationType(ctx context.Context, id string, input *port.UpdateAnnotationTypeInput) error {
	updates := make(map[string]interface{})

	if input.Name != nil {
		updates[constants.NameField] = *input.Name
	}
	if input.Type != nil {
		updates[constants.TagTypeField] = string(*input.Type)
	}
	if input.Global != nil {
		updates[constants.TagGlobalField] = *input.Global
	}
	if input.Required != nil {
		updates[constants.TagRequiredField] = *input.Required
	}
	if len(input.Options) > 0 {
		updates[constants.TagOptionsField] = input.Options
	}
	if input.Min != nil {
		updates[constants.TagMinField] = *input.Min
	}
	if input.Max != nil {
		updates[constants.TagMaxField] = *input.Max
	}
	if input.Color != nil {
		updates[constants.TagColorField] = *input.Color
	}

	if len(updates) == 0 {
		return nil
	}

	return ats.annotationTypeRepo.Update(ctx, id, updates)
}

func (ats *AnnotationTypeService) DeleteAnnotationType(ctx context.Context, id string) error {

	uowErr := ats.uow.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		wsRepo := repos.WorkspaceRepo
		annotationTypeRepo := repos.AnnotationTypeRepo

		pagination := &sharedQuery.Pagination{Limit: 1, Offset: 0}

		wsfilter := []sharedQuery.Filter{
			{
				Field:    constants.ParentIDField,
				Operator: sharedQuery.OpEqual,
				Value:    id,
			},
		}
		result, err := wsRepo.FindByFilters(txCtx, wsfilter, pagination)
		if err != nil {
			return err
		}
		if len(result.Data) > 0 {
			details := map[string]interface{}{"annotation_type_id": "Annotation type is in use by one or more workspaces."}
			return errors.NewConflictError("annotation type in use", details)
		}

		if err := annotationTypeRepo.Delete(txCtx, id); err != nil {
			return err
		}
		return nil
	})

	return uowErr
}

func (ats *AnnotationTypeService) CountAnnotationTypes(ctx context.Context, filters []sharedQuery.Filter) (int64, error) {
	return ats.annotationTypeRepo.Count(ctx, filters)
}

func (ats *AnnotationTypeService) BatchDeleteAnnotationTypes(ctx context.Context, ids []string) error {
	return ats.annotationTypeRepo.BatchDelete(ctx, ids)
}
