package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/adapter/repository/firestore/mappers"
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
	workspaceRepo      port.WorkspaceRepository
	patientRepo        port.PatientRepository
	imageRepo          port.ImageRepository
	annotationRepo     port.AnnotationRepository
	annotationTypeRepo port.AnnotationTypeRepository
	contentRepo        port.ContentRepository
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

func (f *FirestoreUnitOfWorkFactory) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		txCtx := withTx(ctx, tx)
		return fn(txCtx)
	})
}

func (f *FirestoreUnitOfWorkFactory) GetWorkspaceRepo() port.WorkspaceRepository {
	return f.workspaceRepo
}

func (f *FirestoreUnitOfWorkFactory) GetPatientRepo() port.PatientRepository {
	return f.patientRepo
}

func (f *FirestoreUnitOfWorkFactory) GetImageRepo() port.ImageRepository {
	return f.imageRepo
}

func (f *FirestoreUnitOfWorkFactory) GetAnnotationRepo() port.AnnotationRepository {
	return f.annotationRepo
}

func (f *FirestoreUnitOfWorkFactory) GetAnnotationTypeRepo() port.AnnotationTypeRepository {
	return f.annotationTypeRepo
}

func (f *FirestoreUnitOfWorkFactory) GetContentRepo() port.ContentRepository {
	return f.contentRepo
}
