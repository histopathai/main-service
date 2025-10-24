package firestore

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/shared/query"
)

type BaseRepository[T any] struct {
	client     *firestore.Client
	collection string
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) (*T, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity cannot be nil")
	}

	idGetter, ok := any(entity).(interface{ GetID() string })
	if !ok {
		return nil, fmt.Errorf("entity must implement GetID method")
	}

	entityID := idGetter.GetID()

	if entityID == "" {
		entityID = r.client.Collection(r.collection).NewDoc().ID
		idSetter, ok := any(entity).(interface{ SetID(string) })
		if !ok {
			return nil, fmt.Errorf("entity must implement SetID method")
		}
		idSetter.SetID(entityID)
	}

	_, err := r.client.Collection(r.collection).Doc(entityID).Set(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return entity, nil
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id string) (*T, error) {
	doc, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	var entity T
	if err := doc.DataTo(&entity); err != nil {
		return nil, fmt.Errorf("failed to parse document data: %w", err)
	}

	return &entity, nil
}

func (r *BaseRepository[T]) Update(ctx context.Context, id string, entity *T) (*T, error) {
	if entity == nil {
		return nil, fmt.Errorf("entity cannot be nil")
	}

	_, err := r.client.Collection(r.collection).Doc(id).Set(ctx, entity)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return entity, nil
}

func (r *BaseRepository[T]) GetByCriteria(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error) {
	if filters == nil {
		filters = []query.Filter{}
	}
	if paginationOpts == nil {
		paginationOpts = &query.Pagination{
			Limit:  10,
			Offset: 0,
		}
	}

	// Toplam belge sayısını al
	countQuery := r.client.Collection(r.collection).Query
	for _, f := range filters {
		countQuery = countQuery.Where(f.Field, string(f.Operator), f.Value)
	}
	countSnapshot, err := countQuery.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}
	total := len(countSnapshot)

	// Sorguyu oluştur
	query := r.client.Collection(r.collection).Query
	for _, f := range filters {
		query = query.Where(f.Field, string(f.Operator), f.Value)
	}

	// Sorting
	if paginationOpts.SortBy != "" {
		direction := firestore.Asc
		if paginationOpts.SortDir == "desc" {
			direction = firestore.Desc
		}
		query = query.OrderBy(paginationOpts.SortBy, direction)
	}

	// Pagination
	if paginationOpts.Offset > 0 {
		query = query.Offset(paginationOpts.Offset)
	}
	if paginationOpts.Limit > 0 {
		query = query.Limit(paginationOpts.Limit)
	}

	// Execute query
	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch documents: %w", err)
	}

	// Parse documents
	var results []*T
	for _, doc := range docs {
		var entity T
		if err := doc.DataTo(&entity); err != nil {
			return nil, fmt.Errorf("failed to parse document: %w", err)
		}
		results = append(results, &entity)
	}

	// Calculate HasMore
	hasMore := (paginationOpts.Offset + paginationOpts.Limit) < total

	return &query.Result[T]{
		Data:    results,
		Total:   total,
		Limit:   paginationOpts.Limit,
		Offset:  paginationOpts.Offset,
		HasMore: hasMore,
	}, nil
}
