package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/repository"
	// Gerekliyse eklendi
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
	client *firestore.Client
	repos  *repository.Repositories
}

func NewFirestoreUnitOfWorkFactory(client *firestore.Client) (repository.UnitOfWorkFactory, *repository.Repositories) {

	repos := &repository.Repositories{
		WorkspaceRepo:      NewWorkspaceRepositoryImpl(client, true),
		PatientRepo:        NewPatientRepositoryImpl(client, true),
		ImageRepo:          NewImageRepositoryImpl(client, false),
		AnnotationRepo:     NewAnnotationRepositoryImpl(client, false),
		AnnotationTypeRepo: NewAnnotationTypeRepositoryImpl(client, true),
	}

	factory := &FirestoreUnitOfWorkFactory{
		client: client,
		repos:  repos,
	}

	return factory, repos
}

func (f *FirestoreUnitOfWorkFactory) WithTx(ctx context.Context, fn func(ctx context.Context, repos *repository.Repositories) error) error {

	return f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		txCtx := withTx(ctx, tx)

		return fn(txCtx, f.repos)
	})
}
