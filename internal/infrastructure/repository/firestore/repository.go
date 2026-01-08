package firestore

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
	"google.golang.org/api/iterator"
)

type Mapper[T port.Entity] interface {
	FromFirestoreDoc(doc *firestore.DocumentSnapshot) (T, error)
	ToFirestoreMap(entity T) map[string]interface{}
	MapUpdates(updates map[string]interface{}) (map[string]interface{}, error)
	MapFilters(filters []query.Filter) ([]query.Filter, error)
}

type RepositoryImpl[T port.Entity] struct {
	client     *firestore.Client
	collection string
	mapper     Mapper[T]
}

func NewRepositoryImpl[T port.Entity](
	client *firestore.Client,
	collection string,
	mapper Mapper[T],
) *RepositoryImpl[T] {
	return &RepositoryImpl[T]{
		client:     client,
		collection: collection,
		mapper:     mapper,
	}
}

type TransferableRepositoryImpl[T port.Entity] struct {
	*RepositoryImpl[T]
}

func NewTransferableRepositoryImpl[T port.Entity](
	client *firestore.Client,
	collection string,
	mapper Mapper[T],
) *TransferableRepositoryImpl[T] {
	return &TransferableRepositoryImpl[T]{
		RepositoryImpl: NewRepositoryImpl[T](client, collection, mapper),
	}
}

// Create, Read, Update, Delete ve diğer tüm metodlar aynı kalır
// Sadece type constraint değişti: model.Entity -> port.Entity

func (r *RepositoryImpl[T]) Create(ctx context.Context, entity T) (T, error) {
	if reflect.ValueOf(entity).IsNil() {
		var zero T
		return zero, errors.New("entity cannot be nil")
	}

	if entity.GetID() == "" {
		entity.SetID(r.client.Collection(r.collection).NewDoc().ID)
	}

	now := time.Now()
	entity.SetCreatedAt(now)
	entity.SetUpdatedAt(now)

	entityMap := r.mapper.ToFirestoreMap(entity)

	docRef := r.client.Collection(r.collection).Doc(entity.GetID())

	var err error
	if tx := FromCtx(ctx); tx != nil {
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

func (r *RepositoryImpl[T]) Read(ctx context.Context, id string) (T, error) {
	docRef := r.client.Collection(r.collection).Doc(id)
	var doc *firestore.DocumentSnapshot
	var err error

	if tx := FromCtx(ctx); tx != nil {
		doc, err = tx.Get(docRef)
	} else {
		doc, err = docRef.Get(ctx)
	}

	if err != nil {
		var zero T
		return zero, mapFirestoreError(err)
	}

	entity, err := r.mapper.FromFirestoreDoc(doc)
	if err != nil {
		var zero T
		return zero, fmt.Errorf("failed to map firestore document: %w", err)
	}

	return entity, nil
}

func (r *RepositoryImpl[T]) Update(ctx context.Context, id string, updates map[string]any) error {
	mappedUpdates, err := r.mapper.MapUpdates(updates)
	if err != nil {
		return err
	}

	mappedUpdates["updated_at"] = time.Now()

	docRef := r.client.Collection(r.collection).Doc(id)

	if tx := FromCtx(ctx); tx != nil {
		err = tx.Set(docRef, mappedUpdates, firestore.MergeAll)
	} else {
		_, err = docRef.Get(ctx)
		if err != nil {
			return mapFirestoreError(err)
		}

		_, err = docRef.Set(ctx, mappedUpdates, firestore.MergeAll)
	}
	if err != nil {
		return mapFirestoreError(err)
	}

	return nil
}

func (r *RepositoryImpl[T]) Delete(ctx context.Context, id string) error {
	docRef := r.client.Collection(r.collection).Doc(id)

	var err error
	if tx := FromCtx(ctx); tx != nil {
		err = tx.Delete(docRef)
	} else {
		_, err = docRef.Get(ctx)
		if err != nil {
			return mapFirestoreError(err)
		}

		_, err = docRef.Delete(ctx)
	}
	if err != nil {
		return mapFirestoreError(err)
	}

	return nil
}

func (r *RepositoryImpl[T]) FindByFilters(
	ctx context.Context,
	filters []query.Filter,
	paginationOpts *query.Pagination) (*query.Result[T], error) {
	mappedFilters, err := r.mapper.MapFilters(filters)
	if err != nil {
		return nil, err
	}

	q := r.client.Collection(r.collection).Query

	for _, f := range mappedFilters {
		q = q.Where(f.Field, string(f.Operator), f.Value)
	}

	if paginationOpts == nil {
		paginationOpts = &query.Pagination{Limit: 50, Offset: 0, SortBy: "created_at", SortDir: "desc"}
	}

	isLimited := paginationOpts.Limit > 0
	if isLimited {
		q = q.Limit(paginationOpts.Limit + 1).Offset(paginationOpts.Offset)
	}

	var firestoreSortOrder firestore.Direction
	if paginationOpts.SortDir == "asc" {
		firestoreSortOrder = firestore.Asc
	} else {
		firestoreSortOrder = firestore.Desc
	}

	q = q.OrderBy(paginationOpts.SortBy, firestore.Direction(firestoreSortOrder))

	var iter *firestore.DocumentIterator
	if tx := FromCtx(ctx); tx != nil {
		iter = tx.Documents(q)
	} else {
		iter = q.Documents(ctx)
	}
	defer iter.Stop()

	results := make([]T, 0, paginationOpts.Limit)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, mapFirestoreError(err)
		}

		entity, err := r.mapper.FromFirestoreDoc(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to map firestore document: %w", err)
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
		HasMore: hasMore,
		Limit:   paginationOpts.Limit,
		Offset:  paginationOpts.Offset,
	}, nil
}

func (r *RepositoryImpl[T]) Count(ctx context.Context, filters []query.Filter) (int64, error) {
	mappedFilters, err := r.mapper.MapFilters(filters)
	if err != nil {
		return 0, err
	}

	fQuery := r.client.Collection(r.collection).Query
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

func (r *RepositoryImpl[T]) ReadMany(ctx context.Context, ids []string, includeDeleted bool) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}

	const chunkSize = 30
	var allResults []T

	for i := 0; i < len(ids); i += chunkSize {
		end := i + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		chunk := ids[i:end]

		q := r.client.Collection(r.collection).Where(firestore.DocumentID, "in", chunk)

		if !includeDeleted {
			q = q.Where("deleted", "==", false)
		}

		var iter *firestore.DocumentIterator
		if tx := FromCtx(ctx); tx != nil {
			iter = tx.Documents(q)
		} else {
			iter = q.Documents(ctx)
		}

		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}

			if err != nil {
				iter.Stop()
				return nil, mapFirestoreError(err)
			}

			entity, err := r.mapper.FromFirestoreDoc(doc)
			if err != nil {
				iter.Stop()
				return nil, fmt.Errorf("failed to map firestore document: %w", err)
			}

			allResults = append(allResults, entity)
		}
		iter.Stop()
	}

	return allResults, nil
}

func (r *RepositoryImpl[T]) UpdateMany(ctx context.Context, updates map[string]any, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	mappedUpdates, err := r.mapper.MapUpdates(updates)
	if err != nil {
		return err
	}

	mappedUpdates["updated_at"] = time.Now()

	if tx := FromCtx(ctx); tx != nil {
		for _, id := range ids {
			docRef := r.client.Collection(r.collection).Doc(id)
			if err := tx.Set(docRef, mappedUpdates, firestore.MergeAll); err != nil {
				return mapFirestoreError(err)
			}
		}
		return nil
	}

	writer := r.client.BulkWriter(ctx)
	collRef := r.client.Collection(r.collection)

	jobs := make([]*firestore.BulkWriterJob, 0, len(ids))

	for _, id := range ids {
		docRef := collRef.Doc(id)
		job, err := writer.Set(docRef, mappedUpdates, firestore.MergeAll)
		if err != nil {
			writer.End()
			return mapFirestoreError(err)
		}
		jobs = append(jobs, job)
	}

	writer.Flush()

	for _, job := range jobs {
		_, err := job.Results()
		if err != nil {
			writer.End()
			return mapFirestoreError(err)
		}
	}

	writer.End()

	return nil
}

func (r *RepositoryImpl[T]) DeleteMany(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	if tx := FromCtx(ctx); tx != nil {
		for _, id := range ids {
			docRef := r.client.Collection(r.collection).Doc(id)
			if err := tx.Delete(docRef); err != nil {
				return mapFirestoreError(err)
			}
		}
		return nil
	}

	writer := r.client.BulkWriter(ctx)
	collRef := r.client.Collection(r.collection)

	jobs := make([]*firestore.BulkWriterJob, 0, len(ids))

	for _, id := range ids {
		docRef := collRef.Doc(id)
		job, err := writer.Delete(docRef)
		if err != nil {
			writer.End()
			return mapFirestoreError(err)
		}
		jobs = append(jobs, job)
	}

	writer.Flush()

	for _, job := range jobs {
		_, err := job.Results()
		if err != nil {
			writer.End()
			return mapFirestoreError(err)
		}
	}

	writer.End()

	return nil
}

// GetChildren ve GetChildrenPaginated ekle
func (r *RepositoryImpl[T]) GetChildren(ctx context.Context, parentID string, includeDeleted bool) ([]T, error) {
	q := r.client.Collection(r.collection).Where("parent_id", "==", parentID)

	if !includeDeleted {
		q = q.Where("deleted", "==", false)
	}

	var iter *firestore.DocumentIterator
	if tx := FromCtx(ctx); tx != nil {
		iter = tx.Documents(q)
	} else {
		iter = q.Documents(ctx)
	}
	defer iter.Stop()

	var results []T

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, mapFirestoreError(err)
		}

		entity, err := r.mapper.FromFirestoreDoc(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to map firestore document: %w", err)
		}

		results = append(results, entity)
	}

	return results, nil
}

func (r *RepositoryImpl[T]) GetChildrenPaginated(ctx context.Context, parentID string, includeDeleted bool, pagination *query.Pagination) (*query.Result[T], error) {
	if pagination == nil {
		pagination = &query.Pagination{Limit: 50, Offset: 0, SortBy: "created_at", SortDir: "desc"}
	}

	q := r.client.Collection(r.collection).Where("parent_id", "==", parentID)

	if !includeDeleted {
		q = q.Where("deleted", "==", false)
	}

	isLimited := pagination.Limit > 0
	if isLimited {
		q = q.Limit(pagination.Limit + 1).Offset(pagination.Offset)
	}

	var firestoreSortOrder firestore.Direction
	if pagination.SortDir == "asc" {
		firestoreSortOrder = firestore.Asc
	} else {
		firestoreSortOrder = firestore.Desc
	}

	q = q.OrderBy(pagination.SortBy, firestoreSortOrder)

	var iter *firestore.DocumentIterator
	if tx := FromCtx(ctx); tx != nil {
		iter = tx.Documents(q)
	} else {
		iter = q.Documents(ctx)
	}
	defer iter.Stop()

	results := make([]T, 0, pagination.Limit)

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, mapFirestoreError(err)
		}

		entity, err := r.mapper.FromFirestoreDoc(doc)
		if err != nil {
			return nil, fmt.Errorf("failed to map firestore document: %w", err)
		}

		results = append(results, entity)
	}

	hasMore := false
	if isLimited && len(results) > pagination.Limit {
		hasMore = true
		results = results[:pagination.Limit]
	}

	return &query.Result[T]{
		Data:    results,
		HasMore: hasMore,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
	}, nil
}

func (tr *TransferableRepositoryImpl[T]) Transfer(ctx context.Context, id, newOwnerID, transferField string) error {
	docRef := tr.client.Collection(tr.collection).Doc(id)

	updates := map[string]interface{}{
		transferField: newOwnerID,
		"updated_at":  time.Now(),
	}

	var err error
	if tx := FromCtx(ctx); tx != nil {
		err = tx.Set(docRef, updates, firestore.MergeAll)
	} else {
		_, err = docRef.Set(ctx, updates, firestore.MergeAll)
	}
	if err != nil {
		return mapFirestoreError(err)
	}

	return nil
}

func (tr *TransferableRepositoryImpl[T]) TransferMany(ctx context.Context, ids []string, newOwnerID, transferField string) error {
	if len(ids) == 0 {
		return nil
	}

	updates := map[string]interface{}{
		transferField: newOwnerID,
		"updated_at":  time.Now(),
	}

	if tx := FromCtx(ctx); tx != nil {
		for _, id := range ids {
			docRef := tr.client.Collection(tr.collection).Doc(id)
			if err := tx.Set(docRef, updates, firestore.MergeAll); err != nil {
				return mapFirestoreError(err)
			}
		}
		return nil
	}

	writer := tr.client.BulkWriter(ctx)
	collRef := tr.client.Collection(tr.collection)

	jobs := make([]*firestore.BulkWriterJob, 0, len(ids))

	for _, id := range ids {
		docRef := collRef.Doc(id)
		job, err := writer.Set(docRef, updates, firestore.MergeAll)
		if err != nil {
			writer.End()
			return mapFirestoreError(err)
		}
		jobs = append(jobs, job)
	}

	writer.Flush()

	for _, job := range jobs {
		_, err := job.Results()
		if err != nil {
			writer.End()
			return mapFirestoreError(err)
		}
	}

	writer.End()

	return nil
}
