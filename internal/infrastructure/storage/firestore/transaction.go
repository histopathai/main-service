package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	apperrors "github.com/histopathai/main-service-refactor/internal/shared/errors"
	sharedQuery "github.com/histopathai/main-service-refactor/internal/shared/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirestoreTransaction struct {
	client *firestore.Client
	tx     *firestore.Transaction
}

func NewFirestoreTransaction(client *firestore.Client, tx *firestore.Transaction) repository.Transaction {
	return &FirestoreTransaction{
		client: client,
		tx:     tx,
	}
}

func (t *FirestoreTransaction) Create(ctx context.Context, collection string, data interface{}) (string, error) {
	newDocRef := t.client.Collection(collection).NewDoc()
	err := t.tx.Set(newDocRef, data)
	if err != nil {
		return "", apperrors.NewInternalError(fmt.Sprintf("tx: failed to create doc in %s", collection), err)
	}
	return newDocRef.ID, nil
}

func (t *FirestoreTransaction) Get(ctx context.Context, collection string, id string, result interface{}) error {
	docRef := t.client.Collection(collection).Doc(id)
	docSnap, err := t.tx.Get(docRef)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return apperrors.NewNotFoundError(fmt.Sprintf("%s with id %s not found in transaction", collection, id))
		}
		return apperrors.NewInternalError(fmt.Sprintf("tx: failed to get doc %s from %s", id, collection), err)
	}

	if err := docSnap.DataTo(result); err != nil {
		return apperrors.NewInternalError(fmt.Sprintf("tx: failed to parse doc %s from %s", id, collection), err)
	}
	return nil
}

func (t *FirestoreTransaction) Update(ctx context.Context, collection string, id string, data interface{}) error {
	docRef := t.client.Collection(collection).Doc(id)

	switch v := data.(type) {
	case map[string]interface{}:
		updates := []firestore.Update{}
		for key, val := range v {
			updates = append(updates, firestore.Update{Path: key, Value: val})
		}
		err := t.tx.Update(docRef, updates)
		if err != nil {
			return apperrors.NewInternalError(fmt.Sprintf("tx: failed to update doc %s in %s", id, collection), err)
		}
	default:
		err := t.tx.Set(docRef, data)
		if err != nil {
			return apperrors.NewInternalError(fmt.Sprintf("tx: failed to set doc %s in %s", id, collection), err)
		}
	}

	return nil
}

func (t *FirestoreTransaction) Delete(ctx context.Context, collection string, id string) error {
	docRef := t.client.Collection(collection).Doc(id)
	err := t.tx.Delete(docRef)
	if err != nil {
		return apperrors.NewInternalError(fmt.Sprintf("tx: failed to delete doc %s from %s", id, collection), err)
	}
	return nil
}

func (t *FirestoreTransaction) FindByFilters(ctx context.Context, collection string, filters []sharedQuery.Filter, paginationOpts *sharedQuery.Pagination, result interface{}) (int, error) {
	query := t.client.Collection(collection).Query
	for _, f := range filters {
		query = query.Where(f.Field, string(f.Operator), f.Value)
	}

	limit := 10
	if paginationOpts != nil && paginationOpts.Limit > 0 {
		limit = paginationOpts.Limit
	}
	query = query.Limit(limit)

	iter := t.tx.Documents(query)
	docs, err := iter.GetAll()
	if err != nil {
		return 0, apperrors.NewInternalError(fmt.Sprintf("tx: failed to query collection %s", collection), err)
	}

	slicePtr, ok := result.(*[]interface{})
	if !ok {
		return 0, apperrors.NewInternalError("tx: result parameter must be a pointer to a slice", nil)
	}

	count := 0
	for _, docSnap := range docs {

		var item map[string]interface{}
		if err := docSnap.DataTo(&item); err == nil {
			*slicePtr = append(*slicePtr, item)
			count++
		}
	}

	return count, nil
}
