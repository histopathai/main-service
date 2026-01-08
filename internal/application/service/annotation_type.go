package service

import (
	"github.com/histopathai/main-service/internal/application/usecases/common"
	"github.com/histopathai/main-service/internal/application/usecases/composite"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type AnnotationTypeService struct {
	*BaseService[*model.AnnotationType]
}

func NewAnnotationTypeService(
	annotationTypeRepo port.Repository[*model.AnnotationType],
	uowFactory port.UnitOfWorkFactory,
) *AnnotationTypeService {

	createUc := composite.NewCreateUseCase[*model.AnnotationType](uowFactory)
	deleteUc := composite.NewDeleteUseCase(uowFactory)
	updateUc := composite.NewUpdateUseCase[*model.AnnotationType](uowFactory)

	baseSvc := NewBaseService(
		common.NewReadUseCase(annotationTypeRepo),
		common.NewListUseCase(annotationTypeRepo),
		common.NewCountUseCase(annotationTypeRepo),
		common.NewSoftDeleteUseCase(annotationTypeRepo),
		common.NewFilterUseCase(annotationTypeRepo),
		common.NewFilterByParentUseCase(annotationTypeRepo),
		common.NewFilterByCreatorUseCase(annotationTypeRepo),
		common.NewFilterByNameUseCase(annotationTypeRepo),
		deleteUc,
		createUc,
		updateUc,
		vobj.EntityTypeAnnotationType,
	)

	return &AnnotationTypeService{
		BaseService: baseSvc,
	}
}
