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
	current_entity, err := ats.annotationTypeRepo.Read(ctx, id)
	if err != nil {
		return err
	}
	if current_entity == nil {
		return errors.NewNotFoundError("annotation type not found")
	}

	updates := make(map[string]interface{})

	// Name update
	if input.Name != nil && *input.Name != *current_entity.Name {
		updates[constants.NameField] = *input.Name
	}

	// Type update
	if input.Type != nil && input.Type.String() != current_entity.Type.String() {
		updates[constants.TagTypeField] = input.Type.String()
	}

	// Global update
	if input.Global != nil && *input.Global != current_entity.Global {
		updates[constants.TagGlobalField] = *input.Global
	}

	// Required update
	if input.Required != nil && *input.Required != current_entity.Required {
		updates[constants.TagRequiredField] = *input.Required
	}

	// Options update
	if len(input.Options) > 0 && !areOptionsEqual(current_entity.Options, input.Options) {
		updates[constants.TagOptionsField] = input.Options
	}

	// Min update
	if input.Min != nil && (current_entity.Min == nil || *input.Min != *current_entity.Min) {
		updates[constants.TagMinField] = *input.Min
	}

	// Max update
	if input.Max != nil && (current_entity.Max == nil || *input.Max != *current_entity.Max) {
		updates[constants.TagMaxField] = *input.Max
	}

	// Color update
	if input.Color != nil && (current_entity.Color == nil || *input.Color != *current_entity.Color) {
		updates[constants.TagColorField] = *input.Color
	}

	if len(updates) == 0 {
		return nil
	}

	return ats.annotationTypeRepo.Update(ctx, id, updates)
}

// areOptionsEqual
func areOptionsEqual(current, new []string) bool {
	if len(current) != len(new) {
		return false
	}

	currentMap := make(map[string]int, len(current))
	for _, v := range current {
		currentMap[v]++
	}

	for _, v := range new {
		if count, exists := currentMap[v]; !exists || count == 0 {
			return false
		}
		currentMap[v]--
	}

	return true
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
