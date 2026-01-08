package service

import (
	"github.com/histopathai/main-service/internal/application/usecases/common"
	"github.com/histopathai/main-service/internal/application/usecases/composite"
	entityspecific "github.com/histopathai/main-service/internal/application/usecases/entity-specific"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type AnnotationService struct {
	*BaseService[model.Annotation]
}

func NewAnnotationService(
	uow port.UnitOfWorkFactory,
	annotationRepo port.Repository[model.Annotation],
	deleteUc *composite.DeleteUseCase,
) *AnnotationService {

	createUc := entityspecific.NewCreateAnnotationUseCase(uow)
	updateUc := entityspecific.NewUpdateAnnotationUseCase(uow)

	baseSvc := NewBaseService(
		common.NewReadUseCase(annotationRepo),
		common.NewListUseCase(annotationRepo),
		common.NewCountUseCase(annotationRepo),
		common.NewSoftDeleteUseCase(annotationRepo),
		common.NewFilterUseCase(annotationRepo),
		common.NewFilterByParentUseCase(annotationRepo),
		common.NewFilterByCreatorUseCase(annotationRepo),
		common.NewFilterByNameUseCase(annotationRepo),
		deleteUc,
		createUc,
		updateUc,
		vobj.EntityTypeAnnotation,
	)

	return &AnnotationService{
		BaseService: baseSvc,
	}
}
