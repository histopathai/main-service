package firestore

import (
	"context"
	stderrors "errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/port"
	apperrors "github.com/histopathai/main-service/internal/shared/errors"
	"google.golang.org/api/iterator"
)

// RecheckStoreImpl keeps the Ek Kontrol requests, one document per image:
//
//	recheck_requests/{image_id}   image_name, patient_id, patient_name, ws_id, status (open | done),
//	                              reasons: [{code, note, requested_by, requested_at}],
//	                              created_at, updated_at, completed_by, completed_at
type RecheckStoreImpl struct {
	client     *firestore.Client
	collection string
}

func NewRecheckStore(client *firestore.Client) *RecheckStoreImpl {
	return &RecheckStoreImpl{client: client, collection: "recheck_requests"}
}

func (s *RecheckStoreImpl) List(ctx context.Context, status string) ([]port.RecheckRequest, error) {
	q := s.client.Collection(s.collection).Query
	if status != "" {
		q = q.Where("status", "==", status)
	}
	iter := q.Documents(ctx)
	defer iter.Stop()
	out := []port.RecheckRequest{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			return out, nil
		}
		if err != nil {
			if isCollectionNotFoundError(err) {
				return out, nil
			}
			return nil, mapFirestoreError(err)
		}
		out = append(out, recheckFromData(doc.Ref.ID, doc.Data()))
	}
}

func (s *RecheckStoreImpl) Update(ctx context.Context, imageID string,
	change func(*port.RecheckRequest) (*port.RecheckRequest, error)) (*port.RecheckRequest, error) {
	ref := s.client.Collection(s.collection).Doc(imageID)
	var saved *port.RecheckRequest
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var current *port.RecheckRequest
		doc, err := tx.Get(ref)
		switch {
		case notFound(err):
		case err != nil:
			return err
		default:
			r := recheckFromData(doc.Ref.ID, doc.Data())
			current = &r
		}
		next, err := change(current)
		if err != nil {
			return err
		}
		saved = next
		return tx.Set(ref, recheckToData(next))
	})
	if err != nil {
		var appErr *apperrors.Err
		if stderrors.As(err, &appErr) {
			return nil, err
		}
		return nil, mapFirestoreError(err)
	}
	return saved, nil
}

func (s *RecheckStoreImpl) Delete(ctx context.Context, imageID string) error {
	if _, err := s.client.Collection(s.collection).Doc(imageID).Delete(ctx); err != nil && !notFound(err) {
		return mapFirestoreError(err)
	}
	return nil
}

func recheckFromData(id string, data map[string]interface{}) port.RecheckRequest {
	r := port.RecheckRequest{ImageID: id, ImageName: str(data, "image_name"), PatientID: str(data, "patient_id"),
		PatientName: str(data, "patient_name"), WsID: str(data, "ws_id"), Status: str(data, "status"),
		CreatedAt: timeOf(data, "created_at"), UpdatedAt: timeOf(data, "updated_at"),
		CompletedBy: str(data, "completed_by")}
	if t, ok := data["completed_at"].(time.Time); ok {
		r.CompletedAt = &t
	}
	reasons, _ := data["reasons"].([]interface{})
	for _, raw := range reasons {
		m, _ := raw.(map[string]interface{})
		r.Reasons = append(r.Reasons, port.RecheckReason{Code: str(m, "code"), Note: str(m, "note"),
			RequestedBy: str(m, "requested_by"), RequestedAt: timeOf(m, "requested_at")})
	}
	return r
}

func recheckToData(r *port.RecheckRequest) map[string]interface{} {
	reasons := make([]interface{}, len(r.Reasons))
	for i, reason := range r.Reasons {
		reasons[i] = map[string]interface{}{"code": reason.Code, "note": reason.Note,
			"requested_by": reason.RequestedBy, "requested_at": reason.RequestedAt}
	}
	data := map[string]interface{}{
		"image_name": r.ImageName, "patient_id": r.PatientID, "patient_name": r.PatientName, "ws_id": r.WsID,
		"status": r.Status, "reasons": reasons, "created_at": r.CreatedAt, "updated_at": r.UpdatedAt,
		"completed_by": r.CompletedBy, "completed_at": nil,
	}
	if r.CompletedAt != nil {
		data["completed_at"] = *r.CompletedAt
	}
	return data
}
