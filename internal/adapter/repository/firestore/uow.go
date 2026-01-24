package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/adapter/repository/firestore/mappers"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type txKey struct{}

func withTx(ctx context.Context, tx *firestore.Transaction) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func fromCtx(ctx context.Context) *firestore.Transaction {
	tx, _ := ctx.Value(txKey{}).(*firestore.Transaction)
	return tx
}

type FirestoreUnitOfWorkFactory struct {
	client             *firestore.Client
	workspaceRepo      port.Repository[*model.Workspace]
	patientRepo        port.Repository[*model.Patient]
	imageRepo          port.Repository[*model.Image]
	annotationRepo     port.Repository[*model.Annotation]
	annotationTypeRepo port.Repository[*model.AnnotationType]
	contentRepo        port.Repository[*model.Content]
}

func NewFirestoreUnitOfWorkFactory(client *firestore.Client) *FirestoreUnitOfWorkFactory {
	return &FirestoreUnitOfWorkFactory{
		client:             client,
		workspaceRepo:      NewGenericRepositoryImpl(client, "workspaces", mappers.NewWorkspaceMapper()),
		patientRepo:        NewGenericRepositoryImpl(client, "patients", mappers.NewPatientMapper()),
		imageRepo:          NewGenericRepositoryImpl(client, "images", mappers.NewImageMapper()),
		annotationRepo:     NewGenericRepositoryImpl(client, "annotations", mappers.NewAnnotationMapper()),
		annotationTypeRepo: NewGenericRepositoryImpl(client, "annotation_types", mappers.NewAnnotationTypeMapper()),
		contentRepo:        NewGenericRepositoryImpl(client, "contents", mappers.NewContentMapper()),
	}
}

func (f *FirestoreUnitOfWorkFactory) WithTx(ctx context.Context, fn func(ctx context.Context, repos map[vobj.EntityType]any) error) error {
	return f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		txCtx := withTx(ctx, tx)

		repos := map[vobj.EntityType]any{
			vobj.EntityTypeWorkspace:      f.workspaceRepo,
			vobj.EntityTypePatient:        f.patientRepo,
			vobj.EntityTypeImage:          f.imageRepo,
			vobj.EntityTypeAnnotation:     f.annotationRepo,
			vobj.EntityTypeAnnotationType: f.annotationTypeRepo,
			vobj.EntityTypeContent:        f.contentRepo,
		}

		return fn(txCtx, repos)
	})
}

func (f *FirestoreUnitOfWorkFactory) GetWorkspaceRepo() port.Repository[*model.Workspace] {
	return f.workspaceRepo
}

func (f *FirestoreUnitOfWorkFactory) GetPatientRepo() port.Repository[*model.Patient] {
	return f.patientRepo
}

func (f *FirestoreUnitOfWorkFactory) GetImageRepo() port.Repository[*model.Image] {
	return f.imageRepo
}

func (f *FirestoreUnitOfWorkFactory) GetAnnotationRepo() port.Repository[*model.Annotation] {
	return f.annotationRepo
}

func (f *FirestoreUnitOfWorkFactory) GetAnnotationTypeRepo() port.Repository[*model.AnnotationType] {
	return f.annotationTypeRepo
}

func (f *FirestoreUnitOfWorkFactory) GetContentRepo() port.Repository[*model.Content] {
	return f.contentRepo
}
