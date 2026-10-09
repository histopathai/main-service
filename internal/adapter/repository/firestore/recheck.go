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
//	recheck_requests/{image_id}   image_name, patient_id, patient_name, ws_id, assignee_id, status (open | done),
//	                              reasons: [{code, note, requested_by, requested_at, resolved_at}], auto_completed,
//	                              created_at, updated_at, completed_by, completed_at,
//	                              outcome (corrected | no_change | undecided), completion_note
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
		if next == nil {
			return nil
		}
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

// recheckReadChunk is how many documents one GetAll reads.
const recheckReadChunk = 300

func (s *RecheckStoreImpl) UpdateMany(ctx context.Context, imageIDs []string,
	change func(string, *port.RecheckRequest) (*port.RecheckRequest, error)) error {
	coll := s.client.Collection(s.collection)
	bw := s.client.BulkWriter(ctx)
	var jobs []*firestore.BulkWriterJob
	for start := 0; start < len(imageIDs); start += recheckReadChunk {
		end := min(start+recheckReadChunk, len(imageIDs))
		refs := make([]*firestore.DocumentRef, 0, end-start)
		for _, id := range imageIDs[start:end] {
			refs = append(refs, coll.Doc(id))
		}
		docs, err := s.client.GetAll(ctx, refs)
		if err != nil {
			bw.End()
			return mapFirestoreError(err)
		}
		for i, doc := range docs {
			var current *port.RecheckRequest
			if doc.Exists() {
				r := recheckFromData(doc.Ref.ID, doc.Data())
				current = &r
			}
			next, err := change(refs[i].ID, current)
			if err != nil {
				bw.End()
				return err
			}
			var job *firestore.BulkWriterJob
			switch {
			case next != nil:
				job, err = bw.Set(refs[i], recheckToData(next))
			case current != nil:
				job, err = bw.Delete(refs[i])
			default:
				continue
			}
			if err != nil {
				bw.End()
				return mapFirestoreError(err)
			}
			jobs = append(jobs, job)
		}
	}
	bw.End()
	for _, job := range jobs {
		if _, err := job.Results(); err != nil {
			return mapFirestoreError(err)
		}
	}
	return nil
}

func (s *RecheckStoreImpl) Delete(ctx context.Context, imageID string) error {
	if _, err := s.client.Collection(s.collection).Doc(imageID).Delete(ctx); err != nil && !notFound(err) {
		return mapFirestoreError(err)
	}
	return nil
}

func recheckFromData(id string, data map[string]interface{}) port.RecheckRequest {
	r := port.RecheckRequest{ImageID: id, ImageName: str(data, "image_name"), PatientID: str(data, "patient_id"),
		PatientName: str(data, "patient_name"), WsID: str(data, "ws_id"), AssigneeID: str(data, "assignee_id"),
		Status:    str(data, "status"),
		CreatedAt: timeOf(data, "created_at"), UpdatedAt: timeOf(data, "updated_at"),
		CompletedBy: str(data, "completed_by"), Outcome: str(data, "outcome"),
		CompletionNote: str(data, "completion_note")}
	r.AutoCompleted, _ = data["auto_completed"].(bool)
	if t, ok := data["completed_at"].(time.Time); ok {
		r.CompletedAt = &t
	}
	reasons, _ := data["reasons"].([]interface{})
	for _, raw := range reasons {
		m, _ := raw.(map[string]interface{})
		reason := port.RecheckReason{Code: str(m, "code"), Note: str(m, "note"),
			RequestedBy: str(m, "requested_by"), RequestedAt: timeOf(m, "requested_at")}
		if t, ok := m["resolved_at"].(time.Time); ok {
			reason.ResolvedAt = &t
		}
		r.Reasons = append(r.Reasons, reason)
	}
	return r
}

func recheckToData(r *port.RecheckRequest) map[string]interface{} {
	reasons := make([]interface{}, len(r.Reasons))
	for i, reason := range r.Reasons {
		m := map[string]interface{}{"code": reason.Code, "note": reason.Note,
			"requested_by": reason.RequestedBy, "requested_at": reason.RequestedAt, "resolved_at": nil}
		if reason.ResolvedAt != nil {
			m["resolved_at"] = *reason.ResolvedAt
		}
		reasons[i] = m
	}
	data := map[string]interface{}{
		"image_name": r.ImageName, "patient_id": r.PatientID, "patient_name": r.PatientName, "ws_id": r.WsID,
		"assignee_id": r.AssigneeID, "status": r.Status, "reasons": reasons, "created_at": r.CreatedAt, "updated_at": r.UpdatedAt,
		"completed_by": r.CompletedBy, "completed_at": nil,
		"outcome": r.Outcome, "completion_note": r.CompletionNote, "auto_completed": r.AutoCompleted,
	}
	if r.CompletedAt != nil {
		data["completed_at"] = *r.CompletedAt
	}
	return data
}
