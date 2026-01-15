package service

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
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

func (as *AnnotationService) CreateNewAnnotation(ctx context.Context, input *port.CreateAnnotationInput) (*model.Annotation, error) {

	entity, err := vobj.NewEntity(vobj.EntityTypeAnnotation, &input.Name, input.CreatorID, input.Parent)
	if err != nil {
		return nil, err
	}

	annotation := &model.Annotation{
		Entity:   *entity,
		Polygon:  input.Polygon,
		TagValue: input.TagValue,
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
			Field:    constants.ParentIDField,
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

func (as *AnnotationService) UpdateAnnotation(ctx context.Context, id string, input *port.UpdateAnnotationInput) error {
	updates := make(map[string]interface{})

	if input.Polygon != nil {
		updates[constants.PolygonField] = *input.Polygon
	}
	if input.TagType != nil {
		updates[constants.TagTypeField] = input.TagType.String()
	}
	if input.TagName != nil {
		updates[constants.TagNameField] = *input.TagName
	}
	if input.Value != nil {
		updates[constants.TagValueField] = *input.Value
	}
	if input.Color != nil {
		updates[constants.TagColorField] = *input.Color
	}
	if input.Global != nil {
		updates[constants.TagGlobalField] = *input.Global
	}

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	if err := as.annotationRepo.Update(ctx, id, updates); err != nil {
		return err
	}

	return nil
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
