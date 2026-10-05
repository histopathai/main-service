package firestore

import (
	"context"
	stderrors "errors"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/main-service/internal/port"
	apperrors "github.com/histopathai/main-service/internal/shared/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BlindTestStoreImpl keeps the blind test in three collections:
//
//	blind_test_sets/{set_id}                    name, description, image_ids, active, created_at
//	blind_test_keys/{set_id}                    run_id, items: {image_id: {label, source}}   — admins only
//	blind_test_responses/{set_id}__{user_id}    set_id, user_id, user_role, answers: {image_id: {label, answered_at}},
//	                                            notes: {image_id: {text, updated_at}}, started_at, updated_at, completed_at
//
// The key lives in its own collection so that reading a set can never bring it
// along. Sets and keys are written by the experiments repository's upload
// script, not by this service.
type BlindTestStoreImpl struct {
	client    *firestore.Client
	sets      string
	keys      string
	responses string
}

func NewBlindTestStore(client *firestore.Client) *BlindTestStoreImpl {
	return &BlindTestStoreImpl{client: client, sets: "blind_test_sets", keys: "blind_test_keys",
		responses: "blind_test_responses"}
}

func blindTestResponseID(setID, userID string) string { return setID + "__" + userID }

func (s *BlindTestStoreImpl) ListSets(ctx context.Context) ([]port.BlindTestSet, error) {
	iter := s.client.Collection(s.sets).Where("active", "==", true).Documents(ctx)
	defer iter.Stop()
	sets := []port.BlindTestSet{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			return sets, nil
		}
		if err != nil {
			if isCollectionNotFoundError(err) {
				return sets, nil
			}
			return nil, mapFirestoreError(err)
		}
		sets = append(sets, blindTestSetFromDoc(doc))
	}
}

func (s *BlindTestStoreImpl) GetSet(ctx context.Context, setID string) (*port.BlindTestSet, error) {
	doc, err := s.client.Collection(s.sets).Doc(setID).Get(ctx)
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, mapFirestoreError(err)
	}
	set := blindTestSetFromDoc(doc)
	return &set, nil
}

func (s *BlindTestStoreImpl) GetKey(ctx context.Context, setID string) (*port.BlindTestKey, error) {
	doc, err := s.client.Collection(s.keys).Doc(setID).Get(ctx)
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, mapFirestoreError(err)
	}
	data := doc.Data()
	key := &port.BlindTestKey{SetID: doc.Ref.ID, RunID: str(data, "run_id"), Items: map[string]port.BlindTestKeyItem{}}
	items, _ := data["items"].(map[string]interface{})
	for id, raw := range items {
		item, _ := raw.(map[string]interface{})
		source := map[string]string{}
		if src, ok := item["source"].(map[string]interface{}); ok {
			for k, v := range src {
				if text, ok := v.(string); ok {
					source[k] = text
				}
			}
		}
		key.Items[id] = port.BlindTestKeyItem{Label: str(item, "label"), Source: source}
	}
	return key, nil
}

func (s *BlindTestStoreImpl) GetResponse(ctx context.Context, setID, userID string) (*port.BlindTestResponse, error) {
	doc, err := s.client.Collection(s.responses).Doc(blindTestResponseID(setID, userID)).Get(ctx)
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, mapFirestoreError(err)
	}
	return blindTestResponseFromData(doc.Data()), nil
}

func (s *BlindTestStoreImpl) ListResponses(ctx context.Context, setID string) ([]port.BlindTestResponse, error) {
	iter := s.client.Collection(s.responses).Where("set_id", "==", setID).Documents(ctx)
	defer iter.Stop()
	out := []port.BlindTestResponse{}
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
		out = append(out, *blindTestResponseFromData(doc.Data()))
	}
}

func (s *BlindTestStoreImpl) UpdateResponse(ctx context.Context, setID, userID string,
	change func(*port.BlindTestResponse) (*port.BlindTestResponse, error)) (*port.BlindTestResponse, error) {
	ref := s.client.Collection(s.responses).Doc(blindTestResponseID(setID, userID))
	var saved *port.BlindTestResponse
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		var current *port.BlindTestResponse
		doc, err := tx.Get(ref)
		switch {
		case notFound(err):
		case err != nil:
			return err
		default:
			current = blindTestResponseFromData(doc.Data())
		}
		next, err := change(current)
		if err != nil {
			return err
		}
		saved = next
		return tx.Set(ref, blindTestResponseToData(next))
	})
	if err != nil {
		var appErr *apperrors.Err
		if stderrors.As(err, &appErr) {
			return nil, err // refused by change (validation, conflict): passes through as it is
		}
		return nil, mapFirestoreError(err)
	}
	return saved, nil
}

func notFound(err error) bool { return err != nil && status.Code(err) == codes.NotFound }

func str(data map[string]interface{}, key string) string {
	text, _ := data[key].(string)
	return text
}

func timeOf(data map[string]interface{}, key string) time.Time {
	t, _ := data[key].(time.Time)
	return t
}

func blindTestSetFromDoc(doc *firestore.DocumentSnapshot) port.BlindTestSet {
	data := doc.Data()
	set := port.BlindTestSet{ID: doc.Ref.ID, Name: str(data, "name"), Description: str(data, "description"),
		CreatedAt: timeOf(data, "created_at")}
	set.Active, _ = data["active"].(bool)
	if ids, ok := data["image_ids"].([]interface{}); ok {
		for _, id := range ids {
			if text, ok := id.(string); ok {
				set.ImageIDs = append(set.ImageIDs, text)
			}
		}
	}
	return set
}

func blindTestResponseFromData(data map[string]interface{}) *port.BlindTestResponse {
	r := &port.BlindTestResponse{SetID: str(data, "set_id"), UserID: str(data, "user_id"),
		UserRole: str(data, "user_role"), StartedAt: timeOf(data, "started_at"), UpdatedAt: timeOf(data, "updated_at"),
		Answers: map[string]port.BlindTestAnswer{}, Notes: map[string]port.BlindTestNote{}}
	if t, ok := data["completed_at"].(time.Time); ok {
		r.CompletedAt = &t
	}
	answers, _ := data["answers"].(map[string]interface{})
	for id, raw := range answers {
		a, _ := raw.(map[string]interface{})
		r.Answers[id] = port.BlindTestAnswer{Label: str(a, "label"), AnsweredAt: timeOf(a, "answered_at")}
	}
	notes, _ := data["notes"].(map[string]interface{})
	for id, raw := range notes {
		n, _ := raw.(map[string]interface{})
		r.Notes[id] = port.BlindTestNote{Text: str(n, "text"), UpdatedAt: timeOf(n, "updated_at")}
	}
	return r
}

func blindTestResponseToData(r *port.BlindTestResponse) map[string]interface{} {
	answers := map[string]interface{}{}
	for id, a := range r.Answers {
		answers[id] = map[string]interface{}{"label": a.Label, "answered_at": a.AnsweredAt}
	}
	notes := map[string]interface{}{}
	for id, n := range r.Notes {
		notes[id] = map[string]interface{}{"text": n.Text, "updated_at": n.UpdatedAt}
	}
	data := map[string]interface{}{
		"set_id": r.SetID, "user_id": r.UserID, "user_role": r.UserRole, "answers": answers, "notes": notes,
		"started_at": r.StartedAt, "updated_at": r.UpdatedAt, "completed_at": nil,
	}
	if r.CompletedAt != nil {
		data["completed_at"] = *r.CompletedAt
	}
	return data
}
