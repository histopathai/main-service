package firestore

import (
	"context"
	"reflect"
	"time"

	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/query"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/firestore"
)

type GenericRepositoryImpl[T port.Entity] struct {
	client        *firestore.Client
	collection    string
	hasUniqueName bool

	fnFromFirestoreDoc func(doc *firestore.DocumentSnapshot) (T, error)
	fnToFirestoreMap   func(entity T) map[string]interface{}
	fnMapUpdates       func(updates map[string]interface{}) (map[string]interface{}, error)
	fnMapFilters       func(filters []query.Filter) ([]query.Filter, error)
}

func NewGenericRepositoryImpl[T port.Entity](
	client *firestore.Client,
	collection string,
	hasUniqueName bool,
	fnFromFirestoreDoc func(doc *firestore.DocumentSnapshot) (T, error),
	fnToFirestoreMap func(entity T) map[string]interface{},
	fnMapUpdates func(updates map[string]interface{}) (map[string]interface{}, error),
	fnMapFilters func(filters []query.Filter) ([]query.Filter, error),
) *GenericRepositoryImpl[T] {
	return &GenericRepositoryImpl[T]{
		client:             client,
		collection:         collection,
		hasUniqueName:      hasUniqueName,
		fnFromFirestoreDoc: fnFromFirestoreDoc,
		fnToFirestoreMap:   fnToFirestoreMap,
		fnMapUpdates:       fnMapUpdates,
		fnMapFilters:       fnMapFilters,
	}
}

func (gr *GenericRepositoryImpl[T]) Create(ctx context.Context, entity T) (T, error) {

	if reflect.ValueOf(entity).IsNil() {
		var zero T
		return zero, ErrInvalidInput
	}

	if entity.GetID() == "" {

		entity.SetID(gr.client.Collection(gr.collection).NewDoc().ID)
	}

	now := time.Now()
	entity.SetCreatedAt(now)
	entity.SetUpdatedAt(now)

	entityMap := gr.fnToFirestoreMap(entity)

	docRef := gr.client.Collection(gr.collection).Doc(entity.GetID())

	var err error
	if tx := fromCtx(ctx); tx != nil {
		err = tx.Set(docRef, entityMap)
	} else {
		_, err = docRef.Set(ctx, entityMap)
	}
	if err != nil {
		var zero T
		return zero, mapFirestoreError(err)
	}

	return entity, nil
}

func (gr *GenericRepositoryImpl[T]) Read(ctx context.Context, id string) (T, error) {
	docRef := gr.client.Collection(gr.collection).Doc(id)
	var doc *firestore.DocumentSnapshot
	var err error

	if tx := fromCtx(ctx); tx != nil {
		doc, err = tx.Get(docRef)
	} else {
		doc, err = docRef.Get(ctx)
	}

	if err != nil {
		var zero T
		return zero, err
	}

	entity, err := gr.fnFromFirestoreDoc(doc)
	if err != nil {
		var zero T
		return zero, mapFirestoreError(err)
	}

	return entity, nil
}

func (gr *GenericRepositoryImpl[T]) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	updates, err := gr.fnMapUpdates(updates)
	if err != nil {
		return err
	}
	updates["updated_at"] = time.Now()

	docRef := gr.client.Collection(gr.collection).Doc(id)
	if tx := fromCtx(ctx); tx != nil {
		err = tx.Set(docRef, updates, firestore.MergeAll)
	} else {
		_, err = docRef.Set(ctx, updates, firestore.MergeAll)
	}

	if err != nil {
		return mapFirestoreError(err)
	}

	return nil
}

func (gr *GenericRepositoryImpl[T]) Delete(ctx context.Context, id string) error {
	docRef := gr.client.Collection(gr.collection).Doc(id)

	var err error
	if tx := fromCtx(ctx); tx != nil {
		err = tx.Delete(docRef)
	} else {
		_, err = docRef.Delete(ctx)
	}

	if err != nil {
		return mapFirestoreError(err)
	}

	return nil
}

func (gr *GenericRepositoryImpl[T]) Transfer(ctx context.Context, id string, newOwnerID string) error {
	// Not applicable for generic repository, implement if needed
	return nil
}

func (gr *GenericRepositoryImpl[T]) BatchTransfer(ctx context.Context, ids []string, newOwnerID string) error {
	// Not applicable for generic repository, implement if needed
	return nil
}

func (gr *GenericRepositoryImpl[T]) FindByFilters(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error) {
	// Map filters if needed
	mappedFilters, err := gr.fnMapFilters(filters)
	if err != nil {
		return nil, err
	}
	filters = mappedFilters

	fQuery := gr.client.Collection(gr.collection).Query

	for _, f := range filters {
		fQuery = fQuery.Where(f.Field, string(f.Operator), f.Value)
	}

	if paginationOpts == nil {
		paginationOpts = &query.Pagination{Limit: 10, Offset: 0}
	}

	// Apply sorting
	if paginationOpts.SortBy != "" {
		dir := firestore.Asc
		if paginationOpts.SortDir == "desc" {
			dir = firestore.Desc
		}
		// Map SortBy field if necessary. Assuming SortBy is already mapped or is a common field like created_at.
		// NOTE: You might need a mapper for SortBy field similar to filters if field names differ.
		fQuery = fQuery.OrderBy(paginationOpts.SortBy, dir)
	}

	isLimited := paginationOpts.Limit >= 0

	if isLimited {
		fQuery = fQuery.Limit(paginationOpts.Limit + 1).Offset(paginationOpts.Offset)
	}

	var iter *firestore.DocumentIterator
	if tx := fromCtx(ctx); tx != nil {
		iter = tx.Documents(fQuery)
	} else {
		iter = fQuery.Documents(ctx)
	}
	defer iter.Stop()

	results := []T{}

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, mapFirestoreError(err)
		}

		entity, err := gr.fnFromFirestoreDoc(doc)
		if err != nil {
			return nil, err
		}

		results = append(results, entity)
	}

	hasMore := false
	if isLimited && len(results) > paginationOpts.Limit {
		hasMore = true
		results = results[:paginationOpts.Limit]
	}

	return &query.Result[T]{
		Data:    results,
		Limit:   paginationOpts.Limit,
		Offset:  paginationOpts.Offset,
		HasMore: hasMore,
	}, nil

}

func (gr *GenericRepositoryImpl[T]) FindByName(ctx context.Context, name string) (T, error) {
	if !gr.hasUniqueName {
		var zero T
		return zero, ErrInvalidInput
	}

	filters := []query.Filter{
		{
			Field:    "Name", // used FindByFilters will map it
			Operator: query.OpEqual,
			Value:    name,
		},
	}

	result, err := gr.FindByFilters(ctx, filters, &query.Pagination{Limit: 1, Offset: 0})
	if err != nil {
		var zero T
		return zero, err
	}

	if len(result.Data) == 0 {
		var zero T
		return zero, nil
	}

	return result.Data[0], nil
}

func (gr *GenericRepositoryImpl[T]) Count(ctx context.Context, filters []query.Filter) (int64, error) {
	mappedFilters, err := gr.fnMapFilters(filters)
	if err != nil {
		return 0, err
	}

	fQuery := gr.client.Collection(gr.collection).Query
	for _, f := range mappedFilters {
		fQuery = fQuery.Where(f.Field, string(f.Operator), f.Value)
	}

	aggQuery := fQuery.NewAggregationQuery().WithCount("count")

	results, err := aggQuery.Get(ctx)
	if err != nil {
		return 0, mapFirestoreError(err)
	}

	if val, found := results["count"]; found {
		if count, ok := val.(int64); ok {
			return count, nil
		}
	}

	return 0, nil
}

func (gr *GenericRepositoryImpl[T]) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	bulkWriter := gr.client.BulkWriter(ctx)
	collRef := gr.client.Collection(gr.collection)

	for _, id := range ids {
		docRef := collRef.Doc(id)
		_, err := bulkWriter.Delete(docRef)
		if err != nil {
			bulkWriter.End()
			return mapFirestoreError(err)
		}
	}

	bulkWriter.End()

	return nil
}

func (gr *GenericRepositoryImpl[T]) BatchUpdate(ctx context.Context, updates map[string]map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	bulkWriter := gr.client.BulkWriter(ctx)
	collRef := gr.client.Collection(gr.collection)

	for id, updateFields := range updates {
		mappedUpdates, err := gr.fnMapUpdates(updateFields)
		if err != nil {
			bulkWriter.End()
			return err
		}
		mappedUpdates["updated_at"] = time.Now()

		docRef := collRef.Doc(id)
		_, err = bulkWriter.Set(docRef, mappedUpdates, firestore.MergeAll)
		if err != nil {
			bulkWriter.End()
			return mapFirestoreError(err)
		}
	}

	bulkWriter.End()

	return nil
}
