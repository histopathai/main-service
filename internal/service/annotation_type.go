package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
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

func (ats *AnnotationTypeService) ValidateAnnotationTypeCreation(ctx context.Context, input *port.CreateAnnotationTypeInput) error {

	if input.ScoreEnabled {
		if input.ScoreName == nil || *input.ScoreName == "" {
			details := map[string]interface{}{"score_name": "Score name must be provided when score is enabled."}
			return errors.NewValidationError("score name is required", details)
		}
		if input.ScoreMin == nil || input.ScoreMax == nil {
			details := map[string]interface{}{"score_range": "Score min and max must be provided when score is enabled."}
			return errors.NewValidationError("score range is required", details)
		}
		if *input.ScoreMin >= *input.ScoreMax {
			details := map[string]interface{}{"score_range": "Score min must be less than score max."}
			return errors.NewValidationError("invalid score range", details)
		}
	}

	if input.ClassificationEnabled {
		if len(input.ClassList) == 0 {
			details := map[string]interface{}{"class_list": "Class list must be provided when classification is enabled."}
			return errors.NewValidationError("class list is required", details)
		}
	}

	return nil
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

	atModel := &model.AnnotationType{
		Name:                  input.Name,
		ScoreEnabled:          input.ScoreEnabled,
		ClassificationEnabled: input.ClassificationEnabled,
	}
	if input.Description != nil {
		atModel.Description = input.Description
	}

	if input.ScoreEnabled {
		scoreRange := [2]float64{*input.ScoreMin, *input.ScoreMax}
		atModel.ScoreRange = &scoreRange
		atModel.ScoreName = input.ScoreName
	}

	if input.ClassificationEnabled {
		atModel.ClassList = input.ClassList
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

func (ats *AnnotationTypeService) GetClassificationAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[*model.AnnotationType], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    constants.AnnotationTypeClassificationEnabledField,
			Operator: sharedQuery.OpEqual,
			Value:    true,
		},
	}

	return ats.annotationTypeRepo.FindByFilters(ctx, filters, paginationOpts)

}

func (ats *AnnotationTypeService) GetScoreAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[*model.AnnotationType], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    constants.AnnotationTypeScoreEnabledField,
			Operator: sharedQuery.OpEqual,
			Value:    true,
		},
	}

	return ats.annotationTypeRepo.FindByFilters(ctx, filters, paginationOpts)
}

func (ats *AnnotationTypeService) UpdateAnnotationType(ctx context.Context, id string, input *port.UpdateAnnotationTypeInput) error {
	updates := make(map[string]interface{})

	if input.Name != nil {
		updates[constants.AnnotationTypeNameField] = *input.Name
	}
	if input.Description != nil {
		updates[constants.AnnotationTypeDescField] = *input.Description
	}
	if input.ScoreName != nil {
		updates[constants.AnnotationTypeScoreNameField] = *input.ScoreName
	}
	if input.ScoreMin != nil && input.ScoreMax != nil {
		scoreRange := [2]float64{*input.ScoreMin, *input.ScoreMax}
		updates[constants.AnnotationTypeScoreRangeField] = scoreRange
	}
	if input.ClassList != nil {
		updates[constants.AnnotationTypeClassListField] = *input.ClassList
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
				Field:    constants.WorkspaceAnnotationTypeIDField,
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
