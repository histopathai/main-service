package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationService struct {
	annotationRepo port.AnnotationRepository
}

func NewAnnotationService(
	annotationRepo port.AnnotationRepository,
	uow port.UnitOfWorkFactory,
) *AnnotationService {
	return &AnnotationService{
		annotationRepo: annotationRepo,
	}
}

func (as *AnnotationService) validateAnnotationInput(ctx context.Context, input *port.CreateAnnotationInput) error {
	if input.Score == nil && input.Class == nil {
		details := map[string]interface{}{"annotation": "At least one of score or class must be provided."}
		return errors.NewValidationError("invalid annotation input", details)
	}
	return nil
}

func (as *AnnotationService) CreateNewAnnotation(ctx context.Context, input *port.CreateAnnotationInput) (*model.Annotation, error) {
	if err := as.validateAnnotationInput(ctx, input); err != nil {
		return nil, err
	}

	annotation := &model.Annotation{
		ImageID:     input.ImageID,
		AnnotatorID: input.AnnotatorID,
		Polygon:     input.Polygon,
		Score:       input.Score,
		Class:       input.Class,
		Description: input.Description,
	}

	created, err := as.annotationRepo.Create(ctx, annotation)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (as *AnnotationService) GetAnnotationByID(ctx context.Context, id string) (*model.Annotation, error) {
	return as.annotationRepo.Read(ctx, id)
}

func (as *AnnotationService) GetAnnotationsByImageID(ctx context.Context, imageID string, pagination *sharedQuery.Pagination) (*sharedQuery.Result[*model.Annotation], error) {
	filters := []sharedQuery.Filter{
		{
			Field:    constants.AnnotationImageIDField,
			Operator: sharedQuery.OpEqual,
			Value:    imageID,
		},
	}

	result, err := as.annotationRepo.FindByFilters(ctx, filters, pagination)
	if err != nil {
		return nil, err
	}
	if result == nil {
		result = &sharedQuery.Result[*model.Annotation]{
			Data:    make([]*model.Annotation, 0),
			HasMore: false,
		}
	}
	return result, nil
}

func (as *AnnotationService) DeleteAnnotation(ctx context.Context, id string) error {
	return as.annotationRepo.Delete(ctx, id)
}

func (as *AnnotationService) BatchDeleteAnnotations(ctx context.Context, ids []string) error {
	return as.annotationRepo.BatchDelete(ctx, ids)
}

func (as *AnnotationService) CountAnnotations(ctx context.Context, filters []sharedQuery.Filter) (int64, error) {
	return as.annotationRepo.Count(ctx, filters)
}
