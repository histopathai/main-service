package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
)

type AnnotationTypeUseCase struct {
	// Add necessary fields here, e.g., repositories, loggers, etc.
}

func (uc *AnnotationTypeUseCase) CreateAnnotationType(ctx context.Context, entity *model.AnnotationType) (*model.AnnotationType, error) {
	// Implement the logic to create an annotation type
	return nil, nil
}

func (uc *AnnotationTypeUseCase) UpdateAnnotationType(ctx context.Context, updates map[string]interface{}) error {
	// Implement the logic to update an annotation type
	return nil
}

func (uc *AnnotationTypeUseCase) DeleteAnnotationType(ctx context.Context, annotationTypeID string) error {
	// Implement the logic to delete an annotation type
	return nil
}
