package service

import (
	"github.com/histopathai/main-service/internal/application/usecases/common"
	"github.com/histopathai/main-service/internal/application/usecases/composite"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type AnnotationService struct {
	*BaseService[*model.Annotation]
}

func NewAnnotationService(
	annotationRepo port.Repository[*model.Annotation],
	uowFactory port.UnitOfWorkFactory,
) *AnnotationService {

	createUc := composite.NewCreateUseCase[*model.Annotation](uowFactory)
	deleteUc := composite.NewDeleteUseCase(uowFactory)
	// updateUc := composite.NewUpdateUseCase[*model.Annotation](uowFactory)

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
		// updateUc,
		vobj.EntityTypeAnnotation,
	)

	return &AnnotationService{
		BaseService: baseSvc,
	}
}
