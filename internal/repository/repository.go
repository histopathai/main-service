package repository

import (
	"context"
)

type Adapter interface {
	Read(ctx context.Context, collection, docID string) (map[string]interface{}, error)
	Create(ctx context.Context, collection string, data map[string]interface{}) (string, error)
	Update(ctx context.Context, collection, docID string, updates map[string]interface{}) error
	Delete(ctx context.Context, collection, docID string) error
	Query(ctx context.Context, collection string, filters map[string]interface{}) ([]map[string]interface{}, error)
	Exists(ctx context.Context, collection, docID string) (bool, error)
}

type Repository interface {
	Read(ctx context.Context, col string, docID string) (map[string]interface{}, error)
	Create(ctx context.Context, col string, data map[string]interface{}) (string, error)
	Update(ctx context.Context, col string, docID string, updates map[string]interface{}) error
	Delete(ctx context.Context, col string, docID string) error
	Query(ctx context.Context, col string, filters map[string]interface{}) ([]map[string]interface{}, error)
	Exists(ctx context.Context, col string, docID string) (bool, error)
}

type FirestoreRepository struct {
	adapter Adapter
}

func NewFirestoreRepository(adapter Adapter) Repository {
	return &FirestoreRepository{
		adapter: adapter,
	}
}

func (r *FirestoreRepository) Read(ctx context.Context, col string, docID string) (map[string]interface{}, error) {
	return r.adapter.Read(ctx, col, docID)
}

func (r *FirestoreRepository) Create(ctx context.Context, col string, data map[string]interface{}) (string, error) {
	return r.adapter.Create(ctx, col, data)
}

func (r *FirestoreRepository) Update(ctx context.Context, col string, docID string, updates map[string]interface{}) error {
	return r.adapter.Update(ctx, col, docID, updates)
}

func (r *FirestoreRepository) Delete(ctx context.Context, col string, docID string) error {
	return r.adapter.Delete(ctx, col, docID)
}

func (r *FirestoreRepository) Query(ctx context.Context, col string, filters map[string]interface{}) ([]map[string]interface{}, error) {
	return r.adapter.Query(ctx, col, filters)
}

func (r *FirestoreRepository) Exists(ctx context.Context, col string, docID string) (bool, error) {
	return r.adapter.Exists(ctx, col, docID)
}
