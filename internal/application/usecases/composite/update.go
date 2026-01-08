package composite

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type UpdateUseCase[T port.Entity] struct {
	uowFactory port.UnitOfWorkFactory
}

func NewUpdateUseCase[T port.Entity](uowFactory port.UnitOfWorkFactory) *UpdateUseCase[T] {
	return &UpdateUseCase[T]{
		uowFactory: uowFactory,
	}
}

func (uc *UpdateUseCase[T]) Execute(ctx context.Context, id string, updates map[string]any) (T, error) {
	var zero T
	entity_type := zero.GetEntityType()

	switch entity_type {
	case vobj.EntityTypePatient:
		return uc.updatePatient(ctx, id, updates)
	case vobj.EntityTypeImage:
		return uc.updateImage(ctx, id, updates)
	case vobj.EntityTypeWorkspace:
		return uc.updateWorkspace(ctx, id, updates)
	default:
		return zero, errors.NewValidationError("unsupported entity type for update", nil)
	}
}

func (uc *UpdateUseCase[T]) updatePatient(ctx context.Context, id string, updates map[string]any) (T, error) {
	//Simple for now, can be optimized later
	var zero T

	repos, err := uc.uowFactory.WithoutTx(ctx)

	if err != nil {
		return zero, err
	}
	err = repos.PatientRepo.Update(ctx, id, updates)
	if err != nil {
		return zero, err
	}

	updatedEntity, err := repos.PatientRepo.Read(ctx, id)
	if err != nil {
		return zero, err
	}

	updated, ok := any(updatedEntity).(T)
	if !ok {
		return zero, errors.NewValidationError("updated entity is not of expected type", nil)
	}

	return updated, nil
}

func (uc *UpdateUseCase[T]) updateImage(ctx context.Context, id string, updates map[string]any) (T, error) {
	//Simple for now, can be optimized later
	var zero T

	repos, err := uc.uowFactory.WithoutTx(ctx)

	if err != nil {
		return zero, err
	}
	err = repos.ImageRepo.Update(ctx, id, updates)
	if err != nil {
		return zero, err
	}

	updatedEntity, err := repos.ImageRepo.Read(ctx, id)
	if err != nil {
		return zero, err
	}

	updated, ok := any(updatedEntity).(T)
	if !ok {
		return zero, errors.NewValidationError("updated entity is not of expected type", nil)
	}

	return updated, nil
}

func (uc *UpdateUseCase[T]) updateWorkspace(ctx context.Context, id string, updates map[string]any) (T, error) {
	//Simple for now, can be optimized later
	var zero T

	repos, err := uc.uowFactory.WithoutTx(ctx)

	if err != nil {
		return zero, err
	}
	err = repos.WorkspaceRepo.Update(ctx, id, updates)
	if err != nil {
		return zero, err
	}

	updatedEntity, err := repos.WorkspaceRepo.Read(ctx, id)
	if err != nil {
		return zero, err
	}

	updated, ok := any(updatedEntity).(T)
	if !ok {
		return zero, errors.NewValidationError("updated entity is not of expected type", nil)
	}

	return updated, nil
}
