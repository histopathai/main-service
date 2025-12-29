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

func fromCtx(ctx context.Context) *firestore.Transaction {
	tx, _ := ctx.Value(txKey{}).(*firestore.Transaction)
	return tx
}

type FirestoreTransactionManager[T any] struct {
	client *firestore.Client
	repo   repository.Repository[T]
}

func NewFirestoreTransactionManager[T any](client *firestore.Client, repo repository.Repository[T]) repository.TransactionManager[T] {
	return &FirestoreTransactionManager[T]{
		client: client,
		repo:   repo,
	}
}

func (ftm *FirestoreTransactionManager[T]) WithTx(ctx context.Context, fn func(ctx context.Context, repo repository.Repository[T]) error) error {
	return ftm.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		txCtx := withTx(ctx, tx)

		return fn(txCtx, ftm.repo)
	})
}
