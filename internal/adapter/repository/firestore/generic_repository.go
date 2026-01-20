package firestore

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/query"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cloud.google.com/go/firestore"
)

var (
	ErrTooManyOperations = errors.New("too many operations for a single transaction")
)

type IMapper[T port.Entity] interface {
	ToFirestoreMap(entity T) map[string]interface{}
	FromFirestoreDoc(doc *firestore.DocumentSnapshot) (T, error)
	MapUpdates(updates map[string]interface{}) (map[string]interface{}, error)
	MapFilters(filters []query.Filter) ([]query.Filter, error)
}

type GenericRepositoryImpl[T port.Entity] struct {
	client     *firestore.Client
	collection string
	mapper     IMapper[T]
}

func NewGenericRepositoryImpl[T port.Entity](
	client *firestore.Client,
	collection string,
	mapper IMapper[T],
) *GenericRepositoryImpl[T] {
	return &GenericRepositoryImpl[T]{
		client:     client,
		collection: collection,
		mapper:     mapper,
	}
}

func (gr *GenericRepositoryImpl[T]) Create(ctx context.Context, entity T) (T, error) {
	if reflect.ValueOf(entity).IsNil() {
		var zero T
		return zero, ErrInvalidInput
	}

	entityInterface := port.Entity(entity)
	if entityInterface.GetID() == "" {
		entityInterface.SetID(gr.client.Collection(gr.collection).NewDoc().ID)
	}

	now := time.Now()
	entityInterface.SetCreatedAt(now)
	entityInterface.SetUpdatedAt(now)

	entityMap := gr.mapper.ToFirestoreMap(entity)

	docRef := gr.client.Collection(gr.collection).Doc(entityInterface.GetID())

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
		return zero, mapFirestoreError(err)
	}

	entity, err := gr.mapper.FromFirestoreDoc(doc)
	if err != nil {
		var zero T
		return zero, mapFirestoreError(err)
	}

	return entity, nil
}

func (gr *GenericRepositoryImpl[T]) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	updates, err := gr.mapper.MapUpdates(updates)
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
func (gr *GenericRepositoryImpl[T]) UpdateMany(ctx context.Context, ids []string, updates map[string]interface{}) error {
	if len(ids) == 0 {
		return nil
	}

	updates, err := gr.mapper.MapUpdates(updates)
	if err != nil {
		return err
	}
	updates["updated_at"] = time.Now()

	// Check if running in transaction
	if tx := fromCtx(ctx); tx != nil {
		// Transaction mode: Use sequential updates (slower but safe)
		return gr.updateManyInTransaction(ctx, tx, ids, updates)
	}

	// Non-transaction mode: Use BulkWriter (faster)
	return gr.updateManyWithBulkWriter(ctx, ids, updates)
}

func (gr *GenericRepositoryImpl[T]) SoftDelete(ctx context.Context, id string) error {
	updates := map[string]interface{}{
		"is_deleted": true,
		"updated_at": time.Now(),
	}

	docRef := gr.client.Collection(gr.collection).Doc(id)

	var err error
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
func (gr *GenericRepositoryImpl[T]) SoftDeleteMany(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	updates := map[string]interface{}{
		"is_deleted": true,
		"updated_at": time.Now(),
	}

	// Check if running in transaction
	if tx := fromCtx(ctx); tx != nil {
		// Transaction mode: Use sequential updates (slower but safe)
		return gr.updateManyInTransaction(ctx, tx, ids, updates)
	}

	// Non-transaction mode: Use BulkWriter (faster)
	return gr.softDeleteManyWithBulkWriter(ctx, ids, updates)
}

func (gr *GenericRepositoryImpl[T]) Transfer(ctx context.Context, id string, newOwnerID string) error {
	if newOwnerID == "" {
		return ErrInvalidInput
	}

	updates := map[string]interface{}{
		"parent_id":  newOwnerID,
		"updated_at": time.Now(),
	}

	docRef := gr.client.Collection(gr.collection).Doc(id)

	var err error
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
func (gr *GenericRepositoryImpl[T]) TransferMany(ctx context.Context, ids []string, newOwnerID string) error {
	if len(ids) == 0 {
		return nil
	}

	if newOwnerID == "" {
		return ErrInvalidInput
	}

	updates := map[string]interface{}{
		"parent_id":  newOwnerID,
		"updated_at": time.Now(),
	}

	// Check if running in transaction
	if tx := fromCtx(ctx); tx != nil {
		//  Transaction mode: Use sequential updates
		return gr.updateManyInTransaction(ctx, tx, ids, updates)
	}

	// Non-transaction mode: Use BulkWriter (faster)
	return gr.transferManyWithBulkWriter(ctx, ids, updates)
}

func (gr *GenericRepositoryImpl[T]) Count(ctx context.Context, filters []query.Filter) (int64, error) {
	var mappedFilters []query.Filter
	var err error

	if len(filters) > 0 {
		mappedFilters, err = gr.mapper.MapFilters(filters)
		if err != nil {
			return 0, err
		}
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

func (gr *GenericRepositoryImpl[T]) FindByFilters(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination) (*query.Result[T], error) {
	var mappedFilters []query.Filter
	var err error

	if len(filters) > 0 {
		mappedFilters, err = gr.mapper.MapFilters(filters)
		if err != nil {
			return nil, err
		}
	}

	if paginationOpts == nil {
		paginationOpts = &query.Pagination{Limit: 10, Offset: 0}
	}

	hasFilters := len(mappedFilters) > 0
	shouldSort := paginationOpts.SortBy != ""

	result, err := gr.executeQuery(ctx, mappedFilters, paginationOpts, shouldSort)
	if err != nil {
		if isIndexError(err) && hasFilters && shouldSort {
			return gr.executeQuery(ctx, mappedFilters, paginationOpts, false)
		}
		return nil, mapFirestoreError(err)
	}

	return result, nil
}

func (gr *GenericRepositoryImpl[T]) executeQuery(ctx context.Context, filters []query.Filter, paginationOpts *query.Pagination, withSort bool) (*query.Result[T], error) {
	fQuery := gr.client.Collection(gr.collection).Query

	for _, f := range filters {
		fQuery = fQuery.Where(f.Field, string(f.Operator), f.Value)
	}

	if withSort && paginationOpts.SortBy != "" {
		dir := firestore.Asc
		if paginationOpts.SortDir == "desc" {
			dir = firestore.Desc
		}
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
			return nil, err
		}

		entity, err := gr.mapper.FromFirestoreDoc(doc)
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

func isIndexError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "index") || strings.Contains(errMsg, "Index") {
		return true
	}

	if apiErr, ok := err.(*apierror.APIError); ok {
		grpcStatus := apiErr.GRPCStatus()
		if grpcStatus.Code() == codes.FailedPrecondition {
			return true
		}
	}

	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.FailedPrecondition {
			msg := st.Message()
			if strings.Contains(msg, "index") || strings.Contains(msg, "Index") {
				return true
			}
		}
	}

	return false
}

func (gr *GenericRepositoryImpl[T]) updateManyInTransaction(ctx context.Context, tx *firestore.Transaction, ids []string, updates map[string]interface{}) error {
	const maxOpsPerTx = 500

	if len(ids) > maxOpsPerTx {
		return ErrTooManyOperations
	}

	collRef := gr.client.Collection(gr.collection)

	for _, id := range ids {
		docRef := collRef.Doc(id)
		err := tx.Set(docRef, updates, firestore.MergeAll)
		if err != nil {
			return mapFirestoreError(err)
		}
	}

	return nil
}
func (gr *GenericRepositoryImpl[T]) updateManyWithBulkWriter(ctx context.Context, ids []string, updates map[string]interface{}) error {
	bulkWriter := gr.client.BulkWriter(ctx)
	defer bulkWriter.End()

	collRef := gr.client.Collection(gr.collection)

	for _, id := range ids {
		docRef := collRef.Doc(id)
		_, err := bulkWriter.Set(docRef, updates, firestore.MergeAll)
		if err != nil {
			return mapFirestoreError(err)
		}
	}

	return nil
}

func (gr *GenericRepositoryImpl[T]) softDeleteManyWithBulkWriter(ctx context.Context, ids []string, updates map[string]interface{}) error {
	bulkWriter := gr.client.BulkWriter(ctx)
	defer bulkWriter.End()

	collRef := gr.client.Collection(gr.collection)

	for _, id := range ids {
		docRef := collRef.Doc(id)
		_, err := bulkWriter.Set(docRef, updates, firestore.MergeAll)
		if err != nil {
			return mapFirestoreError(err)
		}
	}

	return nil
}

func (gr *GenericRepositoryImpl[T]) transferManyWithBulkWriter(ctx context.Context, ids []string, updates map[string]interface{}) error {
	bulkWriter := gr.client.BulkWriter(ctx)
	defer bulkWriter.End()

	collRef := gr.client.Collection(gr.collection)

	for _, id := range ids {
		docRef := collRef.Doc(id)
		_, err := bulkWriter.Set(docRef, updates, firestore.MergeAll)
		if err != nil {
			return mapFirestoreError(err)
		}
	}

	return nil
}
