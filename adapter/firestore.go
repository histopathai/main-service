package adapter

import (
	"context"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FirestoreAdapter struct {
	client *firestore.Client
}

func NewFirestoreAdapter(client *firestore.Client) *FirestoreAdapter {
	return &FirestoreAdapter{
		client: client,
	}
}

func (fa *FirestoreAdapter) Create(ctx context.Context, collection string, data map[string]interface{}) (string, error) {
	if id, ok := data["id"].(string); ok && id != "" {
		docRef := fa.client.Collection(collection).Doc(id)
		_, err := docRef.Set(ctx, data)
		if err != nil {
			return "", err
		}
		return docRef.ID, nil
	}
	docRef, _, err := fa.client.Collection(collection).Add(ctx, data)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}

func (fa *FirestoreAdapter) Read(ctx context.Context, collection, docID string) (map[string]interface{}, error) {
	docRef := fa.client.Collection(collection).Doc(docID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		return nil, err
	}
	return docSnap.Data(), nil
}

func (fa *FirestoreAdapter) Update(ctx context.Context, collection, docID string, updates map[string]interface{}) error {
	docRef := fa.client.Collection(collection).Doc(docID)
	_, err := docRef.Set(ctx, updates, firestore.MergeAll)
	return err
}

func (fa *FirestoreAdapter) Delete(ctx context.Context, collection, docID string) error {
	docRef := fa.client.Collection(collection).Doc(docID)
	_, err := docRef.Delete(ctx)
	return err
}

func (fa *FirestoreAdapter) Query(ctx context.Context, collection string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	colRef := fa.client.Collection(collection)
	query := colRef.Query

	for field, value := range filters {
		query = query.Where(field, "==", value)
	}

	iter := query.Documents(ctx)
	defer iter.Stop()

	docs, err := iter.GetAll()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0, len(docs))
	for _, d := range docs {
		results = append(results, d.Data())
	}
	return results, nil
}

func (fa *FirestoreAdapter) Exists(ctx context.Context, collection, docID string) (bool, error) {
	docRef := fa.client.Collection(collection).Doc(docID)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return false, nil
		}
		return false, err
	}

	return docSnap.Exists(), nil
}
