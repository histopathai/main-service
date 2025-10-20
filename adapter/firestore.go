package adapter

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	repo "github.com/histopathai/main-service/internal/repository"
)

type FirestoreTransctionAdapter struct {
	client *firestore.Client
	tx     *firestore.Transaction
}

func (fta *FirestoreTransctionAdapter) Create(col string, data map[string]interface{}) (string, error) {
	id, ok := data["id"].(string)
	if !ok || id == "" {
		id = fta.client.Collection(col).NewDoc().ID
		data["id"] = id
	}

	err := fta.tx.Set(fta.client.Collection(col).Doc(id), data)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (fta *FirestoreTransctionAdapter) Read(col string, docID string) (map[string]interface{}, error) {
	doc, err := fta.tx.Get(fta.client.Collection(col).Doc(docID))
	if err != nil {
		return nil, err
	}
	return doc.Data(), nil
}

func (fta *FirestoreTransctionAdapter) Update(col string, docID string, updates map[string]interface{}) error {
	return fta.tx.Set(fta.client.Collection(col).Doc(docID), updates, firestore.MergeAll)
}

func (fta *FirestoreTransctionAdapter) Delete(col string, docID string) error {
	return fta.tx.Delete(fta.client.Collection(col).Doc(docID))
}

func (fta *FirestoreTransctionAdapter) Set(col string, docID string, data map[string]interface{}) error {
	return fta.tx.Set(fta.client.Collection(col).Doc(docID), data)
}

type FirestoreAdapter struct {
	client *firestore.Client
}

func NewFirestoreAdapter(client *firestore.Client) *FirestoreAdapter {
	return &FirestoreAdapter{
		client: client,
	}
}

func (f *FirestoreAdapter) Create(ctx context.Context, col string, data map[string]interface{}) (string, error) {
	id, ok := data["id"].(string)
	if !ok || id == "" {
		id = f.client.Collection(col).NewDoc().ID
		data["id"] = id
	}

	_, err := f.client.Collection(col).Doc(id).Set(ctx, data)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (f *FirestoreAdapter) Read(ctx context.Context, col, docID string) (map[string]interface{}, error) {
	doc, err := f.client.Collection(col).Doc(docID).Get(ctx)
	if err != nil {
		return nil, err
	}
	return doc.Data(), nil
}

func (f *FirestoreAdapter) Update(ctx context.Context, col, docID string, updates map[string]interface{}) error {
	_, err := f.client.Collection(col).Doc(docID).Set(ctx, updates, firestore.MergeAll)
	return err
}

func (f *FirestoreAdapter) Delete(ctx context.Context, col, docID string) error {
	_, err := f.client.Collection(col).Doc(docID).Delete(ctx)
	return err
}

func (f *FirestoreAdapter) Set(ctx context.Context, col, docID string, data map[string]interface{}) error {
	_, err := f.client.Collection(col).Doc(docID).Set(ctx, data)
	return err
}

func (f *FirestoreAdapter) Query(ctx context.Context, col string, filters map[string]interface{}, pagination repo.Pagination) (*repo.QueryResult, error) {
	query := f.client.Collection(col).Query

	for field, value := range filters {
		query = query.Where(field, "==", value)
	}

	if pagination.Offset > 0 {
		query = query.Offset(pagination.Offset)
	}
	if pagination.Limit > 0 {
		query = query.Limit(pagination.Limit)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	results := make([]map[string]interface{}, 0)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		results = append(results, doc.Data())
	}

	return &repo.QueryResult{
		Data:    results,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: len(results) == pagination.Limit,
	}, nil
}

func (f *FirestoreAdapter) Exists(ctx context.Context, col, docID string) (bool, error) {
	_, err := f.client.Collection(col).Doc(docID).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
