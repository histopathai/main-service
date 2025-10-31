package service

import (
	"context"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	errors "github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
)

type AnnotationTypeService struct {
	annotationTypeRepo repository.AnnotationTypeRepository
	uow                repository.UnitOfWorkFactory
}

func NewAnnotationTypeService(
	annotationTypeRepo repository.AnnotationTypeRepository,
	uow repository.UnitOfWorkFactory,
) *AnnotationTypeService {
	return &AnnotationTypeService{
		annotationTypeRepo: annotationTypeRepo,
		uow:                uow,
	}
}

func (ats *AnnotationTypeService) ValidateAnnotationTypeCreation(ctx context.Context, input *CreateAnnotationTypeInput) error {

	if input.ScoreEnabled && input.ClassificationEnabled {
		details := map[string]interface{}{"annotation_type": "An annotation type cannot have both score and classification enabled."}
		return errors.NewValidationError("invalid annotation type configuration", details)
	}

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

type CreateAnnotationTypeInput struct {
	Name                  string
	Description           *string
	ScoreEnabled          bool
	ScoreName             *string
	ScoreMin              *float64
	ScoreMax              *float64
	ClassificationEnabled bool
	ClassList             []string
}

func (ats *AnnotationTypeService) CreateNewAnnotationType(ctx context.Context, input *CreateAnnotationTypeInput) (*model.AnnotationType, error) {

	_, err := ats.annotationTypeRepo.FindByName(ctx, input.Name)
	if err == nil {
		details := map[string]interface{}{"name": "An annotation type with the same name already exists."}
		return nil, errors.NewConflictError("annotation type name already exists", details)
	}

	if err := ats.ValidateAnnotationTypeCreation(ctx, input); err != nil {
		return nil, err
	}

	filter := []sharedQuery.Filter{
		{
			Field:    constants.AnnotationTypeNameField,
			Operator: sharedQuery.OpEqual,
			Value:    input.Name,
		},
	}
	existing, err := ats.annotationTypeRepo.FindByFilters(ctx, filter, &sharedQuery.Pagination{Limit: 1, Offset: 0})
	if err != nil {
		return nil, err
	}
	if len(existing.Data) > 0 {
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

func (ats *AnnotationTypeService) GetAllAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[*model.AnnotationType], error) {
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

type UpdateAnnotationTypeInput struct {
	Name        *string
	Description *string
}

func (ats *AnnotationTypeService) UpdateAnnotationType(ctx context.Context, id string, input *UpdateAnnotationTypeInput) error {
	updates := make(map[string]interface{})

	if input.Name != nil {
		updates[constants.AnnotationTypeNameField] = *input.Name
	}
	if input.Description != nil {
		updates[constants.AnnotationTypeDescField] = *input.Description
	}

	if len(updates) == 0 {
		return nil
	}

	return ats.annotationTypeRepo.Update(ctx, id, updates)
}

func (ats *AnnotationTypeService) DeleteAnnotationType(ctx context.Context, id string) error {

	uowErr := ats.uow.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {
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

func (ats *AnnotationTypeService) ListAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[*model.AnnotationType], error) {
	return ats.annotationTypeRepo.FindByFilters(ctx, []sharedQuery.Filter{}, paginationOpts)
}
