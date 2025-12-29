package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/repository"
)

type txKey struct{}

func withTx(ctx context.Context, tx *firestore.Transaction) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func FromCtx(ctx context.Context) *firestore.Transaction {
	tx, _ := ctx.Value(txKey{}).(*firestore.Transaction)
	return tx
}

type FirestoreUnitOfWorkFactory struct {
	client *firestore.Client
	repos  *repository.Repositories
}

func NewFirestoreUnitOfWorkFactory(
	client *firestore.Client,
	repos *repository.Repositories,
) repository.UnitOfWorkFactory {
	return &FirestoreUnitOfWorkFactory{
		client: client,
		repos:  repos,
	}
}

func (f *FirestoreUnitOfWorkFactory) WithTx(
	ctx context.Context,
	fn func(ctx context.Context, repos *repository.Repositories) error,
) error {
	return f.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		txCtx := withTx(ctx, tx)
		return fn(txCtx, f.repos)
	})
}
