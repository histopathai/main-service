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

// BlindTestInviteStoreImpl keeps the invitation links and the people who
// joined through them:
//
//	blind_test_invites/{sha256(token)}   token, set_id, max_participants, expires_at, active, created_by, created_at
//	blind_test_guests/{guest_id}         invite_id, set_id, name, name_key, pin_hash, session_hashes,
//	                                     failed_attempts, locked_until, consent_at, created_at
//
// Their answers live with everyone else's in blind_test_responses
// (user_id "guest_{guest_id}").
type BlindTestInviteStoreImpl struct {
	client  *firestore.Client
	invites string
	guests  string
}

func NewBlindTestInviteStore(client *firestore.Client) *BlindTestInviteStoreImpl {
	return &BlindTestInviteStoreImpl{client: client, invites: "blind_test_invites", guests: "blind_test_guests"}
}

func (s *BlindTestInviteStoreImpl) CreateInvite(ctx context.Context, inv port.BlindTestInvite) error {
	_, err := s.client.Collection(s.invites).Doc(inv.ID).Create(ctx, inviteToData(inv))
	return mapFirestoreError(err)
}

func (s *BlindTestInviteStoreImpl) GetInvite(ctx context.Context, id string) (*port.BlindTestInvite, error) {
	doc, err := s.client.Collection(s.invites).Doc(id).Get(ctx)
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, mapFirestoreError(err)
	}
	inv := inviteFromDoc(doc)
	if inv.Participants, err = s.countGuests(ctx, id); err != nil {
		return nil, err
	}
	return &inv, nil
}

func (s *BlindTestInviteStoreImpl) ListInvites(ctx context.Context, setID string) ([]port.BlindTestInvite, error) {
	iter := s.client.Collection(s.invites).Where("set_id", "==", setID).Documents(ctx)
	defer iter.Stop()
	out := []port.BlindTestInvite{}
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			if isCollectionNotFoundError(err) {
				return out, nil
			}
			return nil, mapFirestoreError(err)
		}
		out = append(out, inviteFromDoc(doc))
	}
	for i := range out {
		n, err := s.countGuests(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Participants = n
	}
	return out, nil
}

func (s *BlindTestInviteStoreImpl) UpdateInvite(ctx context.Context, id string, change func(*port.BlindTestInvite) error) (*port.BlindTestInvite, error) {
	ref := s.client.Collection(s.invites).Doc(id)
	var saved port.BlindTestInvite
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}
		inv := inviteFromDoc(doc)
		if inv.Participants, err = countIn(tx.Documents(s.guestsOf(id))); err != nil {
			return err
		}
		if err := change(&inv); err != nil {
			return err
		}
		saved = inv
		return tx.Set(ref, inviteToData(inv))
	})
	if err != nil {
		return nil, passAppError(err)
	}
	return &saved, nil
}

func (s *BlindTestInviteStoreImpl) AddGuest(ctx context.Context, g port.BlindTestGuest, max int) error {
	ref := s.client.Collection(s.guests).Doc(g.ID)
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		// Reads first (Firestore transactions): the name, then the head count.
		_, err := tx.Get(ref)
		if err == nil {
			return port.ErrBlindTestNameTaken
		}
		if !notFound(err) {
			return err
		}
		n, err := countIn(tx.Documents(s.guestsOf(g.InviteID)))
		if err != nil {
			return err
		}
		if n >= max {
			return port.ErrBlindTestInviteFull
		}
		return tx.Create(ref, guestToData(g))
	})
	if stderrors.Is(err, port.ErrBlindTestNameTaken) || stderrors.Is(err, port.ErrBlindTestInviteFull) {
		return err
	}
	return mapFirestoreError(err)
}

func (s *BlindTestInviteStoreImpl) GetGuest(ctx context.Context, id string) (*port.BlindTestGuest, error) {
	doc, err := s.client.Collection(s.guests).Doc(id).Get(ctx)
	if notFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, mapFirestoreError(err)
	}
	g := guestFromDoc(doc)
	return &g, nil
}

func (s *BlindTestInviteStoreImpl) UpdateGuest(ctx context.Context, id string, change func(*port.BlindTestGuest) error) (*port.BlindTestGuest, error) {
	ref := s.client.Collection(s.guests).Doc(id)
	var saved port.BlindTestGuest
	err := s.client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {
		doc, err := tx.Get(ref)
		if err != nil {
			return err
		}
		g := guestFromDoc(doc)
		if err := change(&g); err != nil {
			return err
		}
		saved = g
		return tx.Set(ref, guestToData(g))
	})
	if err != nil {
		return nil, passAppError(err)
	}
	return &saved, nil
}

func (s *BlindTestInviteStoreImpl) ListGuests(ctx context.Context, setID string) ([]port.BlindTestGuest, error) {
	iter := s.client.Collection(s.guests).Where("set_id", "==", setID).Documents(ctx)
	defer iter.Stop()
	out := []port.BlindTestGuest{}
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
		out = append(out, guestFromDoc(doc))
	}
}

func (s *BlindTestInviteStoreImpl) guestsOf(inviteID string) firestore.Query {
	return s.client.Collection(s.guests).Where("invite_id", "==", inviteID).Select()
}

func (s *BlindTestInviteStoreImpl) countGuests(ctx context.Context, inviteID string) (int, error) {
	return countIn(s.guestsOf(inviteID).Documents(ctx))
}

func countIn(iter *firestore.DocumentIterator) (int, error) {
	defer iter.Stop()
	n := 0
	for {
		_, err := iter.Next()
		if err == iterator.Done {
			return n, nil
		}
		if err != nil {
			if isCollectionNotFoundError(err) {
				return n, nil
			}
			return 0, err
		}
		n++
	}
}

// passAppError lets an error of the application (from a change function)
// through as it is and maps the rest.
func passAppError(err error) error {
	var appErr *apperrors.Err
	if stderrors.As(err, &appErr) {
		return err
	}
	return mapFirestoreError(err)
}

func inviteToData(inv port.BlindTestInvite) map[string]interface{} {
	data := map[string]interface{}{
		"token": inv.Token, "set_id": inv.SetID, "max_participants": inv.MaxParticipants, "active": inv.Active,
		"created_by": inv.CreatedBy, "created_at": inv.CreatedAt, "expires_at": nil,
	}
	if inv.ExpiresAt != nil {
		data["expires_at"] = *inv.ExpiresAt
	}
	return data
}

func inviteFromDoc(doc *firestore.DocumentSnapshot) port.BlindTestInvite {
	data := doc.Data()
	inv := port.BlindTestInvite{ID: doc.Ref.ID, Token: str(data, "token"), SetID: str(data, "set_id"),
		CreatedBy: str(data, "created_by"), CreatedAt: timeOf(data, "created_at")}
	inv.Active, _ = data["active"].(bool)
	inv.MaxParticipants = intOf(data, "max_participants")
	if t, ok := data["expires_at"].(time.Time); ok {
		inv.ExpiresAt = &t
	}
	return inv
}

func guestToData(g port.BlindTestGuest) map[string]interface{} {
	data := map[string]interface{}{
		"invite_id": g.InviteID, "set_id": g.SetID, "name": g.Name, "name_key": g.NameKey, "pin_hash": g.PinHash,
		"session_hashes": g.SessionHashes, "failed_attempts": g.FailedAttempts, "locked_until": nil,
		"consent_at": g.ConsentAt, "created_at": g.CreatedAt,
	}
	if g.LockedUntil != nil {
		data["locked_until"] = *g.LockedUntil
	}
	if g.SessionHashes == nil {
		data["session_hashes"] = []string{}
	}
	return data
}

func guestFromDoc(doc *firestore.DocumentSnapshot) port.BlindTestGuest {
	data := doc.Data()
	g := port.BlindTestGuest{ID: doc.Ref.ID, InviteID: str(data, "invite_id"), SetID: str(data, "set_id"),
		Name: str(data, "name"), NameKey: str(data, "name_key"), PinHash: str(data, "pin_hash"), FailedAttempts: intOf(data, "failed_attempts"),
		ConsentAt: timeOf(data, "consent_at"), CreatedAt: timeOf(data, "created_at")}
	if t, ok := data["locked_until"].(time.Time); ok {
		g.LockedUntil = &t
	}
	if hashes, ok := data["session_hashes"].([]interface{}); ok {
		for _, h := range hashes {
			if text, ok := h.(string); ok {
				g.SessionHashes = append(g.SessionHashes, text)
			}
		}
	}
	return g
}

func intOf(data map[string]interface{}, key string) int {
	switch v := data[key].(type) {
	case int64:
		return int(v)
	case float64:
		return int(v)
	}
	return 0
}
