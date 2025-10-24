package service

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
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

	created, err := ats.annotationTypeRepo.Create(ctx, newAnnotationType)
	if err != nil {
		ats.logger.Error("Failed to create annotation type", "error", err, "annotationTypeName", input.Name)
		return nil, errors.NewInternalError("failed to create annotation type", err)
	}

	return created, nil
}

func (ats *AnnotationTypeService) GetAnnotationTypeByID(ctx context.Context, id string) (*model.AnnotationType, error) {
	annotationType, err := ats.annotationTypeRepo.GetByID(ctx, id)
	if err != nil {
		ats.logger.Error("Failed to retrieve annotation type", "error", err, "annotationTypeID", id)
		return nil, errors.NewInternalError("failed to retrieve annotation type", err)
	}
	if annotationType == nil {
		details := map[string]interface{}{"annotation_type_id": "Annotation type not found."}
		return nil, errors.NewValidationError("annotation type not found", details)
	}
	return annotationType, nil
}

type UpdateAnnotationTypeInput struct {
	Name        *string
	Description *string
}

func (ats *AnnotationTypeService) UpdateAnnotationType(ctx context.Context, id string, input *UpdateAnnotationTypeInput) error {
	updates := make(map[string]interface{})

	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Description != nil {
		updates["desc"] = *input.Description
	}

	if len(updates) == 0 {
		return nil
	}

	err := ats.annotationTypeRepo.Update(ctx, id, updates)
	if err != nil {
		ats.logger.Error("Failed to update annotation type", "error", err, "annotationTypeID", id)
		return errors.NewInternalError("failed to update annotation type", err)
	}

	return nil
}

func (ats *AnnotationTypeService) GetAllAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.AnnotationType], error) {
	annotationTypes, err := ats.annotationTypeRepo.GetByCriteria(ctx, []sharedQuery.Filter{}, paginationOpts)
	if err != nil {
		ats.logger.Error("Failed to retrieve annotation types", "error", err)
		return nil, errors.NewInternalError("failed to retrieve annotation types", err)
	}
	return annotationTypes, nil
}

func (ats *AnnotationTypeService) GetClassificationAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.AnnotationType], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    "ClassificationEnabled",
			Operator: sharedQuery.OpEqual,
			Value:    true,
		},
	}

	annotationTypes, err := ats.annotationTypeRepo.GetByCriteria(ctx, filters, paginationOpts)
	if err != nil {
		ats.logger.Error("Failed to retrieve classification annotation types", "error", err)
		return nil, errors.NewInternalError("failed to retrieve classification annotation types", err)
	}
	return annotationTypes, nil
}

func (ats *AnnotationTypeService) GetScoreAnnotationTypes(ctx context.Context, paginationOpts *sharedQuery.Pagination) (*sharedQuery.Result[model.AnnotationType], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    "ScoreEnabled",
			Operator: sharedQuery.OpEqual,
			Value:    true,
		},
	}

	annotationTypes, err := ats.annotationTypeRepo.GetByCriteria(ctx, filters, paginationOpts)
	if err != nil {
		ats.logger.Error("Failed to retrieve score annotation types", "error", err)
		return nil, errors.NewInternalError("failed to retrieve score annotation types", err)
	}
	return annotationTypes, nil
}
