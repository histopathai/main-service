package firestore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/query"
	"google.golang.org/api/iterator"
)

type TelemetryRepositoryImpl struct {
	client     *firestore.Client
	collection string
}

func NewTelemetryRepositoryImpl(client *firestore.Client) *TelemetryRepositoryImpl {
	return &TelemetryRepositoryImpl{
		client:     client,
		collection: constants.TelemetryCollection,
	}
}

// Ensure interface compliance
var _ port.TelemetryRepository = (*TelemetryRepositoryImpl)(nil)

func (r *TelemetryRepositoryImpl) Create(ctx context.Context, message *events.TelemetryMessage) error {
	if message == nil {
		return ErrInvalidInput
	}

	// Use provided ID or generate new one
	if message.ID == "" {
		message.ID = r.client.Collection(r.collection).NewDoc().ID
	}

	data := telemetryToFirestoreMap(message)

	_, err := r.client.Collection(r.collection).Doc(message.ID).Set(ctx, data)
	return mapFirestoreError(err)
}

func (r *TelemetryRepositoryImpl) Read(ctx context.Context, id string) (*events.TelemetryMessage, error) {
	doc, err := r.client.Collection(r.collection).Doc(id).Get(ctx)
	if err != nil {
		return nil, mapFirestoreError(err)
	}

	return telemetryFromFirestoreDoc(doc)
}

func (r *TelemetryRepositoryImpl) FindByFilters(ctx context.Context, filters []query.Filter, pagination *query.Pagination) (*query.Result[*events.TelemetryMessage], error) {
	mappedFilters, err := telemetryMapFilters(filters)
	if err != nil {
		return nil, err
	}

	fQuery := r.client.Collection(r.collection).Query
	for _, f := range mappedFilters {
		fQuery = fQuery.Where(f.Field, string(f.Operator), f.Value)
	}

	if pagination == nil {
		pagination = &query.Pagination{Limit: 10, Offset: 0}
	}

	// Clone query for pagination
	pagedQuery := fQuery
	if pagination.Limit >= 0 {
		pagedQuery = pagedQuery.Limit(pagination.Limit + 1).Offset(pagination.Offset)
	}

	// Sort by timestamp by default if not specified
	if pagination.SortBy == "" {
		pagedQuery = pagedQuery.OrderBy("timestamp", firestore.Desc)
	} else {
		// Simple mapping for sort fields could be added here
		dir := firestore.Asc
		if pagination.SortDir == "desc" {
			dir = firestore.Desc
		}
		pagedQuery = pagedQuery.OrderBy(pagination.SortBy, dir)
	}

	iter := pagedQuery.Documents(ctx)
	defer iter.Stop()

	results := []*events.TelemetryMessage{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, mapFirestoreError(err)
		}

		entity, err := telemetryFromFirestoreDoc(doc)
		if err != nil {
			return nil, err
		}
		results = append(results, entity)
	}

	hasMore := false
	if pagination.Limit >= 0 && len(results) > pagination.Limit {
		hasMore = true
		results = results[:pagination.Limit]
	}

	return &query.Result[*events.TelemetryMessage]{
		Data:    results,
		Limit:   pagination.Limit,
		Offset:  pagination.Offset,
		HasMore: hasMore,
	}, nil
}

func (r *TelemetryRepositoryImpl) Count(ctx context.Context, filters []query.Filter) (int64, error) {
	mappedFilters, err := telemetryMapFilters(filters)
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

func (r *TelemetryRepositoryImpl) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(r.collection).Doc(id).Delete(ctx)
	return mapFirestoreError(err)
}

func (r *TelemetryRepositoryImpl) BatchDelete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	bulkWriter := r.client.BulkWriter(ctx)
	collRef := r.client.Collection(r.collection)

	for _, id := range ids {
		_, err := bulkWriter.Delete(collRef.Doc(id))
		if err != nil {
			bulkWriter.End()
			return mapFirestoreError(err)
		}
	}
	bulkWriter.End()
	return nil
}

// Helpers

func telemetryToFirestoreMap(t *events.TelemetryMessage) map[string]interface{} {
	m := map[string]interface{}{
		"timestamp":      t.Timestamp,
		"service":        t.Service,
		"operation":      t.Operation,
		"error_message":  t.ErrorMessage,
		"error_category": string(t.ErrorCategory),
		"error_severity": string(t.ErrorSeverity),
		"retry_count":    t.RetryCount,
	}

	if t.ImageID != nil {
		m["image_id"] = *t.ImageID
	}
	if t.PatientID != nil {
		m["patient_id"] = *t.PatientID
	}
	if t.UserID != nil {
		m["user_id"] = *t.UserID
	}
	if t.OriginalEventType != nil {
		m["original_event_type"] = string(*t.OriginalEventType)
	}
	if t.Metadata != nil {
		m["metadata"] = t.Metadata
	}

	return m
}

func telemetryFromFirestoreDoc(doc *firestore.DocumentSnapshot) (*events.TelemetryMessage, error) {
	data := doc.Data()
	t := &events.TelemetryMessage{
		ID: doc.Ref.ID,
	}

	if v, ok := data["timestamp"].(time.Time); ok {
		t.Timestamp = v
	}
	if v, ok := data["service"].(string); ok {
		t.Service = v
	}
	if v, ok := data["operation"].(string); ok {
		t.Operation = v
	}
	if v, ok := data["error_message"].(string); ok {
		t.ErrorMessage = v
	}
	if v, ok := data["error_category"].(string); ok {
		t.ErrorCategory = events.ErrorCategory(v)
	}
	if v, ok := data["error_severity"].(string); ok {
		t.ErrorSeverity = events.ErrorSeverity(v)
	}
	if v, ok := data["retry_count"].(int64); ok {
		t.RetryCount = int(v)
	}

	if v, ok := data["image_id"].(string); ok {
		t.ImageID = &v
	}
	if v, ok := data["patient_id"].(string); ok {
		t.PatientID = &v
	}
	if v, ok := data["user_id"].(string); ok {
		t.UserID = &v
	}
	if v, ok := data["original_event_type"].(string); ok {
		evt := events.EventType(v)
		t.OriginalEventType = &evt
	}
	if v, ok := data["metadata"].(map[string]interface{}); ok {
		t.Metadata = v
	}

	return t, nil
}

func telemetryMapFilters(filters []query.Filter) ([]query.Filter, error) {
	mapped := make([]query.Filter, 0, len(filters))
	for _, f := range filters {
		var dbField string
		switch f.Field {
		case "service":
			dbField = "service"
		case "operation":
			dbField = "operation"
		case "error_category":
			dbField = "error_category"
		case "error_severity":
			dbField = "error_severity"
		case "image_id":
			dbField = "image_id"
		case "patient_id":
			dbField = "patient_id"
		case "user_id":
			dbField = "user_id"
		default:
			return nil, fmt.Errorf("unknown filter field: %s", f.Field)
		}

		mapped = append(mapped, query.Filter{
			Field:    dbField,
			Operator: f.Operator,
			Value:    f.Value,
		})
	}
	return mapped, nil
}
