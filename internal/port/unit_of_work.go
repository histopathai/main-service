package port

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
)

type Repositories struct {
	WorkspaceRepo      Repository[*model.Workspace]
	PatientRepo        TransferableRepository[*model.Patient]
	ImageRepo          TransferableRepository[*model.Image]
	AnnotationRepo     Repository[*model.Annotation]
	AnnotationTypeRepo Repository[*model.AnnotationType]
}

type UnitOfWorkFactory interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repos *Repositories) error) error
	WithoutTx(ctx context.Context) (*Repositories, error)
}
