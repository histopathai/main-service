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
}

func NewFirestoreUnitOfWorkFactory(client *firestore.Client) *FirestoreUnitOfWorkFactory {
	return &FirestoreUnitOfWorkFactory{
		client:             client,
		workspaceRepo:      NewGenericRepositoryImpl[*model.Workspace](client, "workspaces", mappers.NewWorkspaceMapper()),
		patientRepo:        NewGenericRepositoryImpl[*model.Patient](client, "patients", mappers.NewPatientMapper()),
		imageRepo:          NewGenericRepositoryImpl[*model.Image](client, "images", mappers.NewImageMapper()),
		annotationRepo:     NewGenericRepositoryImpl[*model.Annotation](client, "annotations", mappers.NewAnnotationMapper()),
		annotationTypeRepo: NewGenericRepositoryImpl[*model.AnnotationType](client, "annotation_types", mappers.NewAnnotationTypeMapper()),
	}
}

func (f *FirestoreUnitOfWorkFactory) WithTx(ctx context.Context, fn func(ctx context.Context, repos map[vobj.EntityType]port.Repository[port.Entity]) error) error {
	return f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		txCtx := withTx(ctx, tx)

		repos := make(map[vobj.EntityType]port.Repository[port.Entity])
		repos[vobj.EntityTypeWorkspace] = f.workspaceRepo.(port.Repository[port.Entity])
		repos[vobj.EntityTypePatient] = f.patientRepo.(port.Repository[port.Entity])
		repos[vobj.EntityTypeImage] = f.imageRepo.(port.Repository[port.Entity])
		repos[vobj.EntityTypeAnnotation] = f.annotationRepo.(port.Repository[port.Entity])
		repos[vobj.EntityTypeAnnotationType] = f.annotationTypeRepo.(port.Repository[port.Entity])

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
