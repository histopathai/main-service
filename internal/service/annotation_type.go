package service

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	errors "github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
)

type AnnotationTypeService struct {
	annotationTypeRepo repository.AnnotationTypeRepository
	logger             *slog.Logger
}

func NewAnnotationTypeService(
	annotationTypeRepo repository.AnnotationTypeRepository,
	logger *slog.Logger,
) *AnnotationTypeService {
	return &AnnotationTypeService{
		annotationTypeRepo: annotationTypeRepo,
		logger:             logger,
	}
}

type CreateAnnotationTypeInput struct {
	Name                  string
	Description           *string
	ScoreEnabled          bool
	ScoreName             *string
	ScoreMin              *float64
	ScoreMax              *float64
	ClassificationEnabled bool
	ClassList             *[]string
}

func (ats *AnnotationTypeService) validateAnnotationTypeCreation(ctx context.Context, input *CreateAnnotationTypeInput) error {

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
		if input.ClassList == nil || len(*input.ClassList) == 0 {
			details := map[string]interface{}{"class_list": "Class list must be provided when classification is enabled."}
			return errors.NewValidationError("class list is required", details)
		}
	}

	return nil
}

func (ats *AnnotationTypeService) CreateAnnotationType(ctx context.Context, input *CreateAnnotationTypeInput) (*model.AnnotationType, error) {

	err := ats.validateAnnotationTypeCreation(ctx, input)
	if err != nil {
		return nil, err
	}

	newAnnotationType := &model.AnnotationType{
		Name:                  input.Name,
		Desc:                  input.Description,
		ScoreEnabled:          input.ScoreEnabled,
		ScoreName:             input.ScoreName,
		ClassificationEnabled: input.ClassificationEnabled,
		ClassList:             input.ClassList,
	}

	if input.ScoreEnabled {
		newAnnotationType.ScoreRange = &[2]float64{*input.ScoreMin, *input.ScoreMax}
	}

	return ats.annotationTypeRepo.Create(ctx, newAnnotationType)
}

func (ats *AnnotationTypeService) GetAnnotationTypeByID(ctx context.Context, id string) (*model.AnnotationType, error) {
	return ats.annotationTypeRepo.GetByID(ctx, id)
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

func (ats *AnnotationTypeService) GetAllAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.AnnotationType], error) {
	return ats.annotationTypeRepo.GetByCriteria(ctx, []sharedQuery.Filter{}, paginationOpts)
}

func (ats *AnnotationTypeService) GetClassificationAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.AnnotationType], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    constants.AnnotationTypeClassificationEnabledField,
			Operator: sharedQuery.OpEqual,
			Value:    true,
		},
	}

	return ats.annotationTypeRepo.GetByCriteria(ctx, filters, paginationOpts)

}

func (ats *AnnotationTypeService) GetScoreAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.AnnotationType], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    constants.AnnotationTypeScoreEnabledField,
			Operator: sharedQuery.OpEqual,
			Value:    true,
		},
	}

	return ats.annotationTypeRepo.GetByCriteria(ctx, filters, paginationOpts)
}

func (ats *AnnotationTypeService) Delete(ctx context.Context, id string) error {
	return ats.annotationTypeRepo.DeleteWithValidation(ctx, id)
}
