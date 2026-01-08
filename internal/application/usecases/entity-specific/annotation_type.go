package entityspecific

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
)

type CreateAnnotationTypeUseCase struct {
	repo port.Repository[model.AnnotationType]
}

func NewCreateAnnotationTypeUseCase(repo port.Repository[model.AnnotationType]) *CreateAnnotationTypeUseCase {
	return &CreateAnnotationTypeUseCase{repo: repo}
}

func (uc *CreateAnnotationTypeUseCase) Execute(ctx context.Context, entity *model.AnnotationType) (*model.AnnotationType, error) {
	createdEntity, err := uc.repo.Create(ctx, *entity)
	if err != nil {
		return nil, err
	}
	return &createdEntity, nil
}

type UpdateAnnotationTypeUseCase struct {
	repo port.Repository[model.AnnotationType]
}

func NewUpdateAnnotationTypeUseCase(repo port.Repository[model.AnnotationType]) *UpdateAnnotationTypeUseCase {
	return &UpdateAnnotationTypeUseCase{repo: repo}
}

func (uc *UpdateAnnotationTypeUseCase) Execute(ctx context.Context, id string, updates map[string]any) (*model.AnnotationType, error) {
	// Will be extended with validation logic later
	err := uc.repo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	updatedEntity, err := uc.repo.Read(ctx, id)
	if err != nil {
		return nil, err
	}

	return &updatedEntity, nil
}
