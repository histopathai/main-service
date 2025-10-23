package repository

import (
	"context"
	"time"

	"github.com/histopathai/models"
)

const AnnotationTypesCollection = "annotation_types"

type AnnotationTypeQueryResult struct {
	AnnotationTypes []*models.AnnotationType
	Total           int
	Limit           int
	Offset          int
	HasMore         bool
}

type AnnotationTypeRepository struct {
	repo MainRepository
}

func NewAnnotationTypeRepository(repo MainRepository) *AnnotationTypeRepository {
	return &AnnotationTypeRepository{
		repo: repo,
	}
}
func (atr *AnnotationTypeRepository) GetMainRepository() *MainRepository {
	return &atr.repo
}

func (atr *AnnotationTypeRepository) Create(ctx context.Context, annotationType *models.AnnotationType) (string, error) {
	annotationType.CreatedAt = time.Now()
	annotationType.UpdatedAt = time.Now()
	return atr.repo.Create(ctx, AnnotationTypesCollection, annotationType.ToMap())
}

func (atr *AnnotationTypeRepository) Read(ctx context.Context, annotationTypeID string) (*models.AnnotationType, error) {
	data, err := atr.repo.Read(ctx, AnnotationTypesCollection, annotationTypeID)
	if err != nil {
		return nil, err
	}
	annotationType := &models.AnnotationType{}
	annotationType.FromMap(data)
	return annotationType, nil
}

func (atr *AnnotationTypeRepository) Update(ctx context.Context, annotationTypeID string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return atr.repo.Update(ctx, AnnotationTypesCollection, annotationTypeID, updates)
}

func (atr *AnnotationTypeRepository) Delete(ctx context.Context, annotationTypeID string) error {
	return atr.repo.Delete(ctx, AnnotationTypesCollection, annotationTypeID)
}

func (atr *AnnotationTypeRepository) List(ctx context.Context, filters []Filter, pagination Pagination) (*AnnotationTypeQueryResult, error) {
	result, err := atr.repo.List(ctx, AnnotationTypesCollection, filters, pagination)
	if err != nil {
		return nil, err
	}
	annotation_Types := make([]*models.AnnotationType, 0, len(result.Data))
	for _, item := range result.Data {
		annotationType := &models.AnnotationType{}
		annotationType.FromMap(item)
		annotation_Types = append(annotation_Types, annotationType)
	}

	return &AnnotationTypeQueryResult{
		AnnotationTypes: annotation_Types,
		Total:           result.Total,
		Limit:           result.Limit,
		Offset:          result.Offset,
		HasMore:         result.HasMore,
	}, nil
}

func (atr *AnnotationTypeRepository) Exists(ctx context.Context, annotationTypeID string) (bool, error) {
	return atr.repo.Exists(ctx, AnnotationTypesCollection, annotationTypeID)
}
