package usecase_test

import (
	"context"
	stderrors "errors"
	"strings"
	"testing"
	"time"

	appusecase "github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─────────────────────────────────────────────────────────────────────────────
// Fakes
// ─────────────────────────────────────────────────────────────────────────────

type inviteStore struct {
	invites map[string]port.BlindTestInvite
	guests  map[string]port.BlindTestGuest
}

func newInviteStore() *inviteStore {
	return &inviteStore{invites: map[string]port.BlindTestInvite{}, guests: map[string]port.BlindTestGuest{}}
}

func (s *inviteStore) count(inviteID string) int {
	n := 0
	for _, g := range s.guests {
		if g.InviteID == inviteID {
			n++
		}
	}
	return n
}

func (s *inviteStore) CreateInvite(_ context.Context, inv port.BlindTestInvite) error {
	s.invites[inv.ID] = inv
	return nil
}

func (s *inviteStore) GetInvite(_ context.Context, id string) (*port.BlindTestInvite, error) {
	inv, ok := s.invites[id]
	if !ok {
		return nil, nil
	}
	inv.Participants = s.count(id)
	return &inv, nil
}

func (s *inviteStore) ListInvites(_ context.Context, setID string) ([]port.BlindTestInvite, error) {
	var out []port.BlindTestInvite
	for id, inv := range s.invites {
		if inv.SetID == setID {
			inv.Participants = s.count(id)
			out = append(out, inv)
		}
	}
	return out, nil
}

func (s *inviteStore) UpdateInvite(_ context.Context, id string, change func(*port.BlindTestInvite) error) (*port.BlindTestInvite, error) {
	inv := s.invites[id]
	inv.Participants = s.count(id)
	if err := change(&inv); err != nil {
		return nil, err
	}
	s.invites[id] = inv
	return &inv, nil
}

func (s *inviteStore) AddGuest(_ context.Context, g port.BlindTestGuest, max int) error {
	if _, ok := s.guests[g.ID]; ok {
		return port.ErrBlindTestNameTaken
	}
	if s.count(g.InviteID) >= max {
		return port.ErrBlindTestInviteFull
	}
	s.guests[g.ID] = g
	return nil
}

func (s *inviteStore) GetGuest(_ context.Context, id string) (*port.BlindTestGuest, error) {
	g, ok := s.guests[id]
	if !ok {
		return nil, nil
	}
	return &g, nil
}

func (s *inviteStore) UpdateGuest(_ context.Context, id string, change func(*port.BlindTestGuest) error) (*port.BlindTestGuest, error) {
	g := s.guests[id]
	if err := change(&g); err != nil {
		return nil, err
	}
	s.guests[id] = g
	return &g, nil
}

func (s *inviteStore) ListGuests(_ context.Context, setID string) ([]port.BlindTestGuest, error) {
	var out []port.BlindTestGuest
	for _, g := range s.guests {
		if g.SetID == setID {
			out = append(out, g)
		}
	}
	return out, nil
}

type clock struct{ t time.Time }

func (c *clock) now() time.Time { return c.t }

func code(err error) string {
	var e *errors.Err
	if stderrors.As(err, &e) {
		c, _ := e.Details["code"].(string)
		return c
	}
	return ""
}

func newInvites() (*appusecase.BlindTestInviteUseCase, *inviteStore, *blindStore, *clock) {
	tests, invites := newBlindStore(), newInviteStore()
	c := &clock{t: time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)}
	return appusecase.NewBlindTestInviteUseCaseForTest(invites, tests, c.now), invites, tests, c
}

func join(name, pin string) port.BlindTestJoin {
	return port.BlindTestJoin{Name: name, PIN: pin, Consent: true}
}

// ─────────────────────────────────────────────────────────────────────────────
// Tests
// ─────────────────────────────────────────────────────────────────────────────

func TestInvite_CreateAndInfo(t *testing.T) {
	uc, _, _, c := newInvites()
	ctx := context.Background()

	_, err := uc.Create(ctx, "s1", "admin", 0, nil)
	assert.Equal(t, "max_invalid", code(err))
	past := c.t.Add(-time.Hour)
	_, err = uc.Create(ctx, "s1", "admin", 10, &past)
	assert.Equal(t, "expiry_invalid", code(err))
	_, err = uc.Create(ctx, "off", "admin", 10, nil)
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err), "inactive set")

	inv, err := uc.Create(ctx, "s1", "admin", 2, nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(inv.Token), 32)
	assert.NotEqual(t, inv.Token, inv.ID, "the token is not the key")

	info, err := uc.Info(ctx, inv.Token)
	require.NoError(t, err)
	assert.Equal(t, port.BlindTestInviteInfo{SetName: "Set A", Images: 4, Max: 2, Joinable: true}, *info)

	_, err = uc.Info(ctx, "not-a-token")
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err))
}

func TestInvite_JoinValidatesAndTakesANameOnce(t *testing.T) {
	uc, _, _, _ := newInvites()
	ctx := context.Background()
	inv, err := uc.Create(ctx, "s1", "admin", 10, nil)
	require.NoError(t, err)

	for _, tc := range []struct {
		j    port.BlindTestJoin
		code string
	}{
		{join("A", "1234"), "name_invalid"},
		{join("...", "1234"), "name_invalid"},
		{join("Ayşe Yılmaz", "12a4"), "pin_invalid"},
		{join("Ayşe Yılmaz", "12345"), "pin_invalid"},
		{port.BlindTestJoin{Name: "Ayşe Yılmaz", PIN: "1234"}, "consent_required"},
		{port.BlindTestJoin{Name: "Ayşe Yılmaz", PIN: "1234", Consent: true, ExperienceYears: intPtr(-1)}, "experience_invalid"},
	} {
		_, err := uc.Join(ctx, inv.Token, tc.j)
		assert.Equal(t, tc.code, code(err), "%+v", tc.j)
	}

	s, err := uc.Join(ctx, inv.Token, port.BlindTestJoin{Name: "  Ayşe   Yılmaz ", PIN: "1234", Consent: true,
		Institution: "Uludağ Üniversitesi", ExperienceYears: intPtr(12)})
	require.NoError(t, err)
	assert.Equal(t, "Ayşe Yılmaz", s.Guest.Name, "spaces tidied")
	assert.Equal(t, "ayse-yilmaz", s.Guest.NameKey)
	assert.True(t, strings.HasPrefix(s.Token, s.Guest.ID+"."))
	assert.NotContains(t, s.Guest.PinHash, "1234")

	_, err = uc.Join(ctx, inv.Token, join("AYŞE YILMAZ", "9999"))
	assert.Equal(t, "name_taken", code(err), "same person, other spelling")
	_, err = uc.Join(ctx, inv.Token, join("ayse yilmaz", "9999"))
	assert.Equal(t, "name_taken", code(err), "ASCII spelling")
}

func TestInvite_CapClosedAndExpired(t *testing.T) {
	uc, _, _, c := newInvites()
	ctx := context.Background()
	soon := c.t.Add(time.Hour)
	inv, err := uc.Create(ctx, "s1", "admin", 1, &soon)
	require.NoError(t, err)

	_, err = uc.Join(ctx, inv.Token, join("Bir", "1111"))
	require.NoError(t, err)
	_, err = uc.Join(ctx, inv.Token, join("İki", "2222"))
	assert.Equal(t, "invite_full", code(err))
	info, _ := uc.Info(ctx, inv.Token)
	assert.False(t, info.Joinable)

	// Expired: no new people, but the ones who joined come back.
	c.t = soon.Add(time.Minute)
	_, err = uc.Update(ctx, "s1", inv.ID, nil, intPtr(5))
	require.NoError(t, err)
	_, err = uc.Join(ctx, inv.Token, join("Üç", "3333"))
	assert.Equal(t, "invite_expired", code(err))
	_, err = uc.Resume(ctx, inv.Token, "bir", "1111")
	require.NoError(t, err)

	// Closed: nobody joins, resumes or goes on.
	closed := false
	_, err = uc.Update(ctx, "s1", inv.ID, &closed, nil)
	require.NoError(t, err)
	_, err = uc.Resume(ctx, inv.Token, "bir", "1111")
	assert.Equal(t, "invite_closed", code(err))
	info, _ = uc.Info(ctx, inv.Token)
	assert.True(t, info.Closed)
}

func TestInvite_ResumeLocksAfterFiveWrongPINs(t *testing.T) {
	uc, _, _, c := newInvites()
	ctx := context.Background()
	inv, err := uc.Create(ctx, "s1", "admin", 10, nil)
	require.NoError(t, err)
	_, err = uc.Join(ctx, inv.Token, join("Ayşe Yılmaz", "1234"))
	require.NoError(t, err)

	_, err = uc.Resume(ctx, inv.Token, "Nobody Here", "1234")
	assert.Equal(t, "wrong_credentials", code(err), "unknown name says no more than a wrong PIN")

	for left := 4; left >= 1; left-- {
		_, err = uc.Resume(ctx, inv.Token, "ayşe yılmaz", "0000")
		require.Equal(t, "wrong_credentials", code(err))
		var e *errors.Err
		require.True(t, stderrors.As(err, &e))
		assert.Equal(t, left, e.Details["attempts_left"])
	}
	_, err = uc.Resume(ctx, inv.Token, "ayşe yılmaz", "0000")
	assert.Equal(t, "locked", code(err), "fifth wrong PIN")
	_, err = uc.Resume(ctx, inv.Token, "ayşe yılmaz", "1234")
	assert.Equal(t, "locked", code(err), "even the right PIN waits")

	c.t = c.t.Add(16 * time.Minute)
	s, err := uc.Resume(ctx, inv.Token, "Ayse Yilmaz", "1234")
	require.NoError(t, err)
	assert.Equal(t, 0, s.Guest.FailedAttempts)
	assert.Nil(t, s.Guest.LockedUntil)
}

func TestInvite_SessionsAuthenticateTheirOwnLinkOnly(t *testing.T) {
	uc, invites, _, _ := newInvites()
	ctx := context.Background()
	a, err := uc.Create(ctx, "s1", "admin", 10, nil)
	require.NoError(t, err)
	b, err := uc.Create(ctx, "s1", "admin", 10, nil)
	require.NoError(t, err)

	s, err := uc.Join(ctx, a.Token, join("Ayşe", "1234"))
	require.NoError(t, err)
	g, err := uc.Authenticate(ctx, a.Token, s.Token)
	require.NoError(t, err)
	assert.Equal(t, s.Guest.ID, g.ID)

	_, err = uc.Authenticate(ctx, b.Token, s.Token)
	assert.Equal(t, "session_invalid", code(err), "another link")
	_, err = uc.Authenticate(ctx, a.Token, s.Token+"x")
	assert.Equal(t, "session_invalid", code(err), "tampered secret")
	_, err = uc.Authenticate(ctx, a.Token, "no-dot")
	assert.Equal(t, "session_invalid", code(err))

	// Every resume is another device; the oldest of more than five is dropped.
	first := s.Token
	for i := 0; i < 5; i++ {
		_, err = uc.Resume(ctx, a.Token, "ayşe", "1234")
		require.NoError(t, err)
	}
	assert.Len(t, invites.guests[s.Guest.ID].SessionHashes, 5)
	_, err = uc.Authenticate(ctx, a.Token, first)
	assert.Equal(t, "session_invalid", code(err), "the first device fell out")
}

func TestInvite_CapCannotGoBelowParticipants(t *testing.T) {
	uc, _, _, _ := newInvites()
	ctx := context.Background()
	inv, err := uc.Create(ctx, "s1", "admin", 5, nil)
	require.NoError(t, err)
	for _, n := range []string{"Bir", "İki"} {
		_, err = uc.Join(ctx, inv.Token, join(n, "1234"))
		require.NoError(t, err)
	}
	_, err = uc.Update(ctx, "s1", inv.ID, nil, intPtr(1))
	assert.Equal(t, "max_below_participants", code(err))
	_, err = uc.Update(ctx, "other", inv.ID, nil, intPtr(3))
	assert.Equal(t, errors.ErrorTypeNotFound, errType(err), "wrong set")
	list, err := uc.List(ctx, "s1")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, 2, list[0].Participants)
}

func TestInvite_GuestsTakeTheTestAndAreNamedInResults(t *testing.T) {
	invUC, invites, tests, _ := newInvites()
	blind := appusecase.NewBlindTestUseCase(tests, &blindStorage{}).WithGuests(invites)
	ctx := context.Background()
	inv, err := invUC.Create(ctx, "s1", "admin", 10, nil)
	require.NoError(t, err)
	s, err := invUC.Join(ctx, inv.Token, port.BlindTestJoin{Name: "Ayşe Yılmaz", PIN: "1234", Consent: true,
		Institution: "Uludağ", ExperienceYears: intPtr(12)})
	require.NoError(t, err)

	user := appusecase.BlindTestGuestUserID(s.Guest)
	for _, id := range []string{"a", "b", "c", "d"} {
		_, err = blind.Answer(ctx, "s1", user, appusecase.BlindTestGuestRole, id, "real")
		require.NoError(t, err)
	}
	_, err = blind.Complete(ctx, "s1", user)
	require.NoError(t, err)

	res, err := blind.Results(ctx, "s1")
	require.NoError(t, err)
	require.Len(t, res.Users, 1)
	u := res.Users[0]
	assert.Equal(t, "guest", u.Response.UserRole)
	require.NotNil(t, u.Guest)
	assert.Equal(t, port.BlindTestGuestProfile{Name: "Ayşe Yılmaz", Institution: "Uludağ", ExperienceYears: intPtr(12)}, *u.Guest)
	assert.Equal(t, 2, u.Score.Correct)
}

func intPtr(i int) *int { return &i }
